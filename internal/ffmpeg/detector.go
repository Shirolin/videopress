package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	cachedGPUEncoders map[string]string = make(map[string]string)
	cacheMu           sync.RWMutex
	EnableDebugLog    bool // 调试日志全局开关，由 GUI 端进行同步
)

// ResetGPUEncoderCache 重置 GPU 探测缓存。主要用于测试。
func ResetGPUEncoderCache() {
	cacheMu.Lock()
	cachedGPUEncoders = make(map[string]string)
	cacheMu.Unlock()
}

// DetectGPUEncoder 探测系统针对指定编码格式可用的硬件加速编码器。
// codec 支持: "h264", "h265" (或 "hevc"), "av1". 默认为 "h264".
// 探测成功则返回对应的硬件加速编码器名称，否则返回该格式的 CPU 软解编码器 (libx264/libx265/libsvtav1).
func DetectGPUEncoder(ffmpegPath string, codec string, runCmd func(name string, args []string) error) string {
	isTest := runCmd != nil

	// 标准化 codec 名称
	codec = strings.ToLower(codec)
	if codec == "" {
		codec = "h264"
	} else if codec == "hevc" {
		codec = "h265"
	}

	// 1. 如果非测试环境，先检查内存缓存，再检查磁盘缓存
	if !isTest {
		cacheMu.RLock()
		cached := cachedGPUEncoders[codec]
		cacheMu.RUnlock()
		if cached != "" {
			return cached
		}

		cacheDir, err := os.UserCacheDir()
		if err == nil {
			cacheFile := filepath.Join(cacheDir, "videopress_gpu.cache")
			if data, err := os.ReadFile(cacheFile); err == nil {
				// 尝试解析为 JSON
				var diskCache map[string]string
				if err := json.Unmarshal(data, &diskCache); err == nil && diskCache != nil {
					cachedDisk := diskCache[codec]
					if cachedDisk != "" {
						cacheMu.Lock()
						cachedGPUEncoders[codec] = cachedDisk
						cacheMu.Unlock()
						return cachedDisk
					}
				}
			}
		}
	}

	var encoders []string
	var detected string

	switch codec {
	case "h265":
		encoders = []string{"hevc_nvenc", "hevc_qsv", "hevc_amf"}
		detected = "libx265"
	case "av1":
		encoders = []string{"av1_nvenc", "av1_qsv", "av1_amf"}
		detected = "libsvtav1"
	default:
		encoders = []string{"h264_nvenc", "h264_qsv", "h264_amf"}
		detected = "libx264"
	}

	if isTest {
		// 测试模式下使用串行同步执行，以便单元测试计数正常且没有数据竞争
		for _, enc := range encoders {
			args := []string{
				"-f", "lavfi",
				"-i", "color=c=black:s=16x16",
				"-vframes", "1",
				"-c:v", enc,
				"-f", "null",
				"-",
			}
			if err := runCmd(ffmpegPath, args); err == nil {
				detected = enc
				break
			}
		}
	} else {
		// 生产环境：多 GPU 编码器并发探测 + 1.5s 严格超时控制，彻底避免驱动挂起造成的长久卡顿
		var wg sync.WaitGroup
		var mu sync.Mutex
		results := make(map[string]bool)
		var debugLogs []string

		for _, enc := range encoders {
			wg.Add(1)
			go func(e string) {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
				defer cancel()

				args := []string{
					"-f", "lavfi",
					"-i", "color=c=black:s=16x16",
					"-vframes", "1",
					"-c:v", e,
					"-f", "null",
					"-",
				}
				cmd := exec.CommandContext(ctx, ffmpegPath, args...)
				prepareCmd(cmd)
				
				var stderrBuf bytes.Buffer
				cmd.Stderr = &stderrBuf
				err := cmd.Run()

				mu.Lock()
				results[e] = (err == nil)
				if err != nil {
					errMsg := strings.TrimSpace(stderrBuf.String())
					if errMsg == "" {
						errMsg = err.Error()
					}
					// 替换多余的空行，整理为单行便于日志阅读
					errMsg = strings.ReplaceAll(errMsg, "\r\n", " ")
					errMsg = strings.ReplaceAll(errMsg, "\n", " ")
					debugLogs = append(debugLogs, fmt.Sprintf("- [%s]: %s", e, errMsg))
				}
				mu.Unlock()
			}(enc)
		}
		wg.Wait()

		// 按照优先级获取支持的编码器
		for _, enc := range encoders {
			if results[enc] {
				detected = enc
				break
			}
		}

		// 将探测结果写入本地磁盘缓存，下次启动 0ms 响应
		cacheDir, err := os.UserCacheDir()
		if err == nil {
			cacheFile := filepath.Join(cacheDir, "videopress_gpu.cache")
			
			// 读取现有缓存合并，保留其他 codec 的探测结果
			var diskCache map[string]string = make(map[string]string)
			if data, err := os.ReadFile(cacheFile); err == nil {
				_ = json.Unmarshal(data, &diskCache)
			}
			diskCache[codec] = detected

			if data, err := json.Marshal(diskCache); err == nil {
				_ = os.WriteFile(cacheFile, data, 0o644)
			}

			// 如果存在错误，并且启用了调试日志，写入调试日志文件方便定位硬件/驱动/FFmpeg配置问题
			if len(debugLogs) > 0 && EnableDebugLog {
				logFile := filepath.Join(cacheDir, "videopress_debug.log")
				logMsg := fmt.Sprintf("[%s] GPU %s 探测失败明细 (最终使用 %s):\n%s\n\n",
					time.Now().Format("2006-01-02 15:04:05"),
					codec,
					detected,
					strings.Join(debugLogs, "\n"),
				)
				f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
				if err == nil {
					_, _ = f.WriteString(logMsg)
					_ = f.Close()
				}
			}
		}
	}

	// 缓存探测结果至内存
	if !isTest {
		cacheMu.Lock()
		cachedGPUEncoders[codec] = detected
		cacheMu.Unlock()
	}

	return detected
}
