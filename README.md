# Videopress

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Windows-0078d7?style=flat-square&logo=windows" alt="Platform">
  <img src="https://img.shields.io/badge/Language-Go%201.26+-00add8?style=flat-square&logo=go" alt="Language">
  <img src="https://img.shields.io/badge/FFmpeg-Required-green?style=flat-square&logo=ffmpeg" alt="FFmpeg">
  <img src="https://img.shields.io/badge/Framework-Wails%20v2-red?style=flat-square" alt="Wails">
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=flat-square" alt="License">
</p>

**Videopress** 是一个专为 Windows 系统设计的、基于 FFmpeg 的高颜值视频压缩工具。
它完美融合了**桌面图形界面 (GUI)** 与**控制台命令行 (CLI)** 双入口，不仅能满足日常拖拽和右键菜单快速压缩，还能完美适配终端效率工具流。
它免去了记忆 FFmpeg 复杂参数的烦恼，只提供精心调校的常用预设规格，同时支持并发队列、硬件加速自动探测以及系统右键集成。

---

## 🛠 运行工作流

```text
 ┌──────────┐      右键菜单 / 命令行 / GUI 拖拽      ┌────────────┐      自动探测 GPU      ┌─────────────┐
 │ 视频源文件 ├────────────────────────────────────>│ Videopress ├───────────────────────>│ 硬件加速编码 │
 └──────────┘                                        └─────┬──────┘  (NVENC / QSV / AMF)   └──────┬──────┘
                                                           │                                      │
                                                           ▼                                      ▼
                                                     ┌────────────┐                         ┌─────────────┐
                                                     │  并发队列  │                         │ FFmpeg 压缩 │
                                                     │ (Worker)   │                         └──────┬──────┘
                                                     └─────┬──────┘                                │
                                                           │                                      │
                                                           ▼                                      ▼
                                                     ┌────────────┐                         ┌─────────────┐
                                                     │ GUI / CLI  │<────────────────────────┤ 压缩完成输出 │
                                                     │ 进度与摘要  │                         │(compressed/)│
                                                     └────────────┘                         └─────────────┘
```

---

## ⚡️ 快速上手

只需三步，即可将软件集成进你的 Windows 系统：

### 1. 准备依赖
1. 下载或编译得到 `videopress.exe`。
2. 下载并安装 [FFmpeg](https://ffmpeg.org/download.html)。确保 `ffmpeg.exe` 已加入系统环境变量 `PATH`，或者直接将 `ffmpeg.exe` 放置在与 `videopress.exe` 相同的目录下。

### 2. 集成到系统（二选一或同时配置）

#### 💡 选项 A：绑定右键“发送到”菜单（适合日常交互）
在 `videopress.exe` 所在目录打开终端，运行：
```powershell
.\videopress.exe --install-sendto
```
* **如何使用**：在资源管理器中，选择你想要压缩的视频文件（支持多选），**右键 -> 发送到 -> 快速压缩视频** 即可自动开始压缩。程序会自动识别调用环境，在控制台弹出精美的动态进度条。

#### 💻 选项 B：配置环境变量 Path（适合命令行效率用户）
在 `videopress.exe` 所在目录打开终端，运行：
```powershell
.\videopress.exe --install-path
```
* **如何使用**：重新打开一个终端（CMD/PowerShell）后，你可以在**任意目录**下直接运行 `videopress <视频路径>` 启动命令行压缩，或直接运行 `videopress` 打开高颜值图形界面。

---

## ✨ 核心特性

* 🖥 **双入口无缝切换**：
  * **GUI 模式**：双击程序或不带参数运行时，展示基于 Wails + Svelte 开发的现代化深色毛玻璃美感图形界面。支持文件拖拽、压缩队列管理、全局设置配置及 GPU 加速检测切换。
  * **CLI 模式**：在终端附带视频文件路径运行时，自动进入控制台模式，实时渲染无抖动的平滑进度条，并在完成后输出直观的压缩率摘要表格。
* 🚀 **GPU 硬件加速**：自动检测你显卡支持的硬编技术（NVIDIA `h264_nvenc`、Intel `h264_qsv`、AMD `h264_amf`）。检测到后自动开启加速，未检测到时安全 fallback 到 CPU 编码。
* 🧵 **高效并发**：通过 Worker Pool 机制实现多视频文件并发压缩。CLI 模式下并发数大于 1 时，进度条展示将优雅地降级为整洁的动态状态日志，避免终端闪烁。
* 📝 **自动日志归档**：压缩结束后，程序会在对应的输出目录自动创建/追加 `compress_summary.log`，供随时复盘压缩率和处理耗时。
* 🌐 **环境变量一键绑定**：支持通过 `--install-path` 自动将当前程序目录追加到 Windows 当前用户的 `Path` 环境变量中，同时支持 `--uninstall-path` 安全移除，实现跨终端随处调用。
* 🌐 **多语言与国际化支持 (i18n)**：
  * **界面与命令行全覆盖**：程序在 GUI（配置面板、弹窗、核心设置）与 CLI（终端用法、参数帮助、进度提示）中完全实现中/英文国际化。
  * **系统级 UI 语言自适应**：Go 后端会在启动时通过 Windows `GetUserDefaultUILanguage` API 检测系统 UI 语言并自动加载中文或英文。
  * **右键菜单名称热重载**：在配置面板中切换语言时，如果已安装右键快捷菜单，程序会自动在注册表中覆盖修改 `MUIVerb`，实现右键菜单名称的热切换（中文：“使用 Videopress 压缩” / 英文：“Compress with Videopress”）。
  * **自定义语言选择**：配置面板提供“界面语言”下拉菜单，便于用户灵活手动指定首选语言。

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

### 命令行版本直接编译（不包含 GUI 界面）

若你仅需要纯命令行版本，可直接使用 Go 编译：
```powershell
# 编译 CLI 静态程序
go build -o .\bin\videopress.exe .\main.go
```

### GUI+CLI 双入口完整版编译（推荐）

完整版需要使用 [Wails CLI](https://wails.io/zh-Hans/docs/gettingstarted/installation) 工具进行构建（需本地安装 Node.js 以编译前端 Svelte）：

```powershell
# 1. 安装 Wails CLI（如未安装）
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 2. 启动本地热重载开发环境
wails dev

# 3. 编译发布 Windows 完整版单文件程序
wails build
```
编译生成的 `videopress.exe` 将输出在根目录下，双击即可启动高颜值 GUI 界面，拖入命令行即可执行 CLI 压缩。
