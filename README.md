# Videopress

Windows 下基于 FFmpeg 的快速视频压缩工具。

用于把本地视频快速压缩成更适合网站、博客、社交媒体发布的 MP4 文件，不暴露复杂参数，只保留三档常用规格。

## 功能

- 三档压缩规格：`small`、`standard`、`quality`（大小写不敏感）
- **高颜值终端界面**：彩色日志、高保真动态进度条、对齐的批量任务总结表格
- **GPU 硬件加速**：支持自动探测 NVENC/QSV/AMF 并启用显卡加速，大幅缩短耗时
- **多并发控制**：支持通过 Goroutine Pool 并发压缩多个文件
- **音频流直通 (Copy)**：支持跳过音频重编码，无损复制原片 AAC 音轨
- **丰富的文件策略**：支持 `--force` 强制覆盖或 `--skip-existing` 增量跳过
- **运行日志归档**：在压缩目录下自动生成 `compress_summary.log` 运行报告与体积占比统计
- **Windows SendTo 原生快捷绑定**：自动生成优雅的系统 `.lnk` 快捷菜单（无 `.cmd` 后缀且带应用图标）

## 依赖

- Windows 10/11
- Go 1.26+（自行构建时）
- 已安装 `ffmpeg`，可通过以下任一方式提供：
  - 系统 `PATH` 中可找到 `ffmpeg`
  - 与 `videopress.exe` 同目录放置 `ffmpeg.exe`

## 预设说明

- `small`：小文件，适合聊天预览和快速分享
- `standard`：默认档，适合博客、网站、社媒发布
- `quality`：更高观感，适合高清压缩

## 命令行选项

| 长选项 | 短选项 | 说明 |
| :--- | :--- | :--- |
| `--preset <规格>` | - | 压缩规格：`small` / `standard` / `quality`（默认 `standard`） |
| `--concurrency <数>` | `-c` | 最大并发压缩任务数（默认 `1`，并发大于 1 时终端输出简洁多行进度） |
| `--hw` | - | 自动检测并启用 GPU 硬件加速编码（支持 Nvidia / Intel / AMD） |
| `--copy-audio` | `-a` | 音频流直通：直接复制源片音频流，不重编码音频，速度更快 |
| `--force` | `-f` | 强制覆盖已存在的压缩文件，禁止重命名为 `xxx.compressed-1.mp4` 副本 |
| `--skip-existing` | - | 增量压缩：如果对应压缩目标路径已存在，则自动跳过压缩 |
| `--install-sendto` | - | 安装 SendTo 右键快捷方式（创建原生 `.lnk`，防止一闪即逝） |
| `--uninstall-sendto`| - | 移除 SendTo 快捷方式（一并清除旧版 `.cmd` 和新版 `.lnk`） |
| `--version` | - | 显示版本号 |
| `-h`, `--help` | - | 显示帮助菜单 |

### 命令行示例

```powershell
# 压缩单个视频（默认 preset standard）
videopress.exe C:\videos\clip.mp4

# 使用 GPU 加速并保留音频复制，使用 small 规格压缩多个视频
videopress.exe --preset small --hw -a C:\videos\a.mp4 C:\videos\b.mov

# 开启 3 个并发任务进行增量批量压缩
videopress.exe -c 3 --skip-existing C:\videos\a.mp4 C:\videos\b.mp4 C:\videos\c.mp4

# 安装 SendTo 右键菜单
videopress.exe --install-sendto
```

默认输出文件名格式：

```text
原文件名.<preset>.compressed.mp4
```

如果目标文件已存在且未开启 `--force`，会自动追加 `-1`、`-2` 后缀。

### 退出码

- `0`：全部压缩成功（或已成功跳过）
- `1`：存在失败、输入无效，或所有文件均为非视频格式被跳过

### 支持的视频格式

`.mp4` `.mov` `.mkv` `.avi` `.m4v` `.wmv` `.webm` `.ts` `.flv` `.mpg` `.mpeg` `.3gp`

## 构建

```powershell
go build -o .\bin\videopress.exe .\cmd\videopress
```

## SendTo 使用

先运行一次安装以绑定右键：

```powershell
.\bin\videopress.exe --install-sendto
```

之后在 Windows 资源管理器中右键视频文件，选择：

```text
发送到 > 快速压缩视频
```

卸载：

```powershell
.\bin\videopress.exe --uninstall-sendto
```

> 注意：SendTo 快捷方式记录的是安装时 `videopress.exe` 的绝对路径。若移动了可执行文件，重新运行一次 `--install-sendto` 即可自动更新路径。

