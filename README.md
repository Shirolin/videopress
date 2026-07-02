# Videopress

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Windows-0078d7?style=flat-square&logo=windows" alt="Platform">
  <img src="https://img.shields.io/badge/Language-Go%201.26+-00add8?style=flat-square&logo=go" alt="Language">
  <img src="https://img.shields.io/badge/FFmpeg-Required-green?style=flat-square&logo=ffmpeg" alt="FFmpeg">
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=flat-square" alt="License">
</p>

**Videopress** 是一个专为 Windows 控制台和资源管理器设计的、基于 FFmpeg 的高颜值视频压缩工具。  
它免去了记忆 FFmpeg 复杂参数的烦恼，只提供精心调校的常用预设规格，同时支持并发队列、硬件加速自动探测以及完美的 Windows 右键菜单绑定，是博客作者、社交媒体运营人员和前端开发者的效率利器。

---

## 🛠 运行工作流

```text
 ┌──────────┐      右键菜单 / 命令行      ┌────────────┐      自动探测 GPU      ┌─────────────┐
 │ 视频源文件 ├─────────────────────────>│ Videopress ├───────────────────────>│ 硬件加速编码 │
 └──────────┘                            └─────┬──────┘  (NVENC / QSV / AMF)   └──────┬──────┘
                                               │                                      │
                                               ▼                                      ▼
                                         ┌────────────┐                         ┌─────────────┐
                                         │  并发队列  │                         │ FFmpeg 压缩 │
                                         │ (Worker)   │                         └──────┬──────┘
                                         └─────┬──────┘                                │
                                               │                                      │
                                               ▼                                      ▼
                                         ┌────────────┐                         ┌─────────────┐
                                         │ 终端UI表格 │<────────────────────────┤ 压缩完成输出 │
                                         │  运行日志  │                         │(compressed/)│
                                         └────────────┘                         └─────────────┘
```

---

## ⚡️ 快速上手

只需三步，即可将软件集成进你的 Windows 系统：

### 1. 准备依赖
1. 下载并构建 [videopress.exe](https://github.com/Shirolin/videopress)。
2. 下载并安装 [FFmpeg](https://ffmpeg.org/download.html)。确保 `ffmpeg.exe` 已加入系统环境变量 `PATH`，或者直接将 `ffmpeg.exe` 放置在与 `videopress.exe` 相同的目录下。

### 2. 绑定右键菜单
在 `videopress.exe` 所在目录打开终端，运行：
```powershell
.\videopress.exe --install-sendto
```
提示安装成功后，你可以双击或回车关闭窗口。

### 3. 一键视频压缩
在资源管理器中，选择你想要压缩的视频文件（支持多选），**右键 -> 发送到 -> 快速压缩视频**。  
压缩任务会自动启动，你将在控制台中看到精美的实时进度条。

---

## ✨ 核心特性

* 🎨 **高颜值交互**：基于 ANSI 彩色高亮和零抖动的绿色平滑进度条设计。压缩完成后渲染出对齐的摘要表格（展示原始大小、压缩大小、存储节省率、耗时）。
* 🚀 **GPU 硬件加速**：自动检测你显卡支持的硬编技术（NVIDIA `h264_nvenc`、Intel `h264_qsv`、AMD `h264_amf`）。检测到后自动开启加速，未检测到时安全 fallback 到 CPU 编码。
* 🧵 **高效并发**：通过 Worker Pool 机制实现多视频文件并发压缩。并发数大于 1 时，进度条展示将优雅地降级为整洁的动态状态日志，避免终端闪烁。
* 📝 **自动日志归档**：压缩结束后，程序会在对应的输出目录自动创建/追加 `compress_summary.log`，供随时复盘压缩率和处理耗时。
* 🌐 **环境变量一键绑定**：支持通过 `--install-path` 自动将当前程序目录追加到 Windows 当前用户的 `Path` 环境变量中，同时支持 `--uninstall-path` 安全移除，实现跨终端随处调用。
* 🛠 **大白话错误提取**：FFmpeg 执行出错时自动丢弃上千行繁杂堆栈，只提取易读的中文错误原因（如“宽度或高度必须是偶数”）。

---

## ⚙️ 预设说明

Videopress 内置了 3 档精心调校的 CRF 质量参数，满足不同场景下的压缩需求：

* 📱 **`small`**：低码率，适合聊天预览、移动端低流量快速分享。
* 🖥 **`standard`**（默认）：平衡档，保留优异画质的同时极大限度缩减体积，适合博客、网站、社媒发布。
* 🎨 **`quality`**：高码率，极致的画质还原，适合高清视频存档与展示。

---

## 💻 命令行选项

如果你习惯使用命令行，Videopress 提供了极简且强大的命令行标志：

| 长选项 | 短选项 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `--preset <规格>` | - | `standard`| 压缩预设规格：`small` / `standard` / `quality` |
| `--concurrency <数>` | `-c` | `1` | 最大并发压缩任务数 |
| `--hw` | - | `false` | 自动探测并启用 GPU 硬件加速编码 |
| `--copy-audio` | `-a` | `false` | 音频流直通：不重新编码音频，保留原音质且加快速度 |
| `--force` | `-f` | `false` | 强制覆盖同名压缩文件，禁止创建自增副本 |
| `--skip-existing` | - | `false` | 增量模式：如果输出路径已存在同名压缩文件，则跳过 |
| `--install-sendto` | - | `false` | 安装 SendTo 右键菜单绑定 |
| `--uninstall-sendto`| - | `false` | 移除 SendTo 快捷方式（兼容清除旧版本） |
| `--install-path` | - | `false` | 将当前程序目录添加至当前用户的 Path 环境变量中 |
| `--uninstall-path`| - | `false` | 从当前用户的 Path 环境变量中安全移除程序目录 |

### 命令行进阶示例

```powershell
# 开启 GPU 加速，用 small 档压缩，并且直接 Copy 音轨（不编码音频）
videopress.exe --preset small --hw -a C:\videos\movie.mp4

# 开启 3 并发，增量压缩整个文件夹中的视频，如果已经压缩过则直接跳过
videopress.exe -c 3 --skip-existing C:\videos\a.mp4 C:\videos\b.mp4 C:\videos\c.mp4

# 一键配置环境变量（使你在任意目录下均可直接调用 videopress）
videopress.exe --install-path
```

---

## 📦 自行构建

若需从源码构建，请确保你的本地已安装 Go 1.26+ 环境：

```powershell
# 克隆仓库并构建
git clone https://github.com/Shirolin/videopress.git
cd videopress
go build -o .\bin\videopress.exe .\cmd\videopress
```


