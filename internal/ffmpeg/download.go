package ffmpeg

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"videopress/internal/util"
)

// downloadURL 使用国内无障碍高速淘宝镜像源的 gzip 压缩版 ffmpeg 二进制 (约 28MB)
const downloadURL = "https://registry.npmmirror.com/-/binary/ffmpeg-static/v6.0/win32-x64"

// DownloadFFmpeg 从国内 npmmirror 镜像源流式下载并在本地解压安装 ffmpeg.exe
func DownloadFFmpeg(destDir string, onProgress func(percent float64)) error {
	destPath := filepath.Join(destDir, "ffmpeg.exe")
	tmpPath := destPath + ".tmp"

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建安装目录失败: %w", err)
	}

	// 网络错误最多重试 3 次，指数退避 1s/2s/4s
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}
		if lastErr = downloadAndExtract(tmpPath, destPath, onProgress); lastErr == nil {
			break
		}
	}
	if lastErr != nil {
		return lastErr
	}

	// 校验下载产物可执行，不可用则清理防止污染
	if err := verifyFFmpeg(destPath); err != nil {
		_ = os.Remove(destPath)
		return err
	}
	return nil
}

func downloadAndExtract(tmpPath string, destPath string, onProgress func(percent float64)) error {
	client := &http.Client{Timeout: 5 * time.Minute}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("创建下载请求失败: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求下载失败，请检查网络连接: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载服务器返回错误状态码: %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength
	if totalSize <= 0 {
		// 备用兜底估计大小为 30MB
		totalSize = 30 * 1024 * 1024
	}

	lastPercent := -1.0
	progressReader := &ProgressReader{
		Reader: resp.Body,
		Total:  totalSize,
		OnProgress: func(bytesRead int64, total int64) {
			if onProgress != nil {
				pct := float64(bytesRead) / float64(total) * 100.0
				if pct > 100 {
					pct = 100
				}
				// 只有当百分比变化超过 0.5% 或者已完成时，才发送更新，节省 IPC 性能
				if pct-lastPercent >= 0.5 || pct >= 100.0 {
					lastPercent = pct
					onProgress(pct)
				}
			}
		},
	}

	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}

	// 淘宝镜像源的 win32-x64 二进制经过了 gzip 压缩，我们直接流式解压写入
	gzipReader, gzipErr := gzip.NewReader(progressReader)
	if gzipErr == nil {
		defer gzipReader.Close()
		_, err = io.Copy(tmpFile, gzipReader)
		closeErr := tmpFile.Close()
		if err == nil {
			err = closeErr
		}
		if err != nil {
			return fmt.Errorf("解压并写入文件时出错: %w", err)
		}
	} else {
		// 源不是 gzip 格式，回退为直接当作普通二进制写入
		tmpFile.Close()
		tmpFile2, err2 := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err2 != nil {
			return fmt.Errorf("重新创建临时文件失败: %w", err2)
		}
		_, err2 = io.Copy(tmpFile2, resp.Body)
		tmpFile2.Close() // 必须立刻在这里关闭，防止 Windows 锁定临时文件导致下一步 Rename 失败！
		if err2 != nil {
			return fmt.Errorf("直接下载写入失败: %w", err2)
		}
	}

	// 如果 ffmpeg.exe 已存在，先移除它（Windows 下 rename 不覆盖已存在文件）
	if _, err := os.Stat(destPath); err == nil {
		_ = os.Remove(destPath)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("应用安装程序失败: %w", err)
	}

	return nil
}

// verifyFFmpeg 通过执行 ffmpeg -version 校验下载产物真实可运行
func verifyFFmpeg(ffmpegPath string) error {
	cmd := exec.Command(ffmpegPath, "-version")
	util.HideConsoleWindow(cmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("下载的 ffmpeg 无法正常执行: %w", err)
	}
	return nil
}

// ProgressReader 用于跟踪流式读取进度
type ProgressReader struct {
	Reader     io.Reader
	Total      int64
	Current    int64
	OnProgress func(bytesRead int64, total int64)
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	pr.Current += int64(n)
	if pr.OnProgress != nil && n > 0 {
		pr.OnProgress(pr.Current, pr.Total)
	}
	return n, err
}
