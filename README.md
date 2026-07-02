# Videopress

Windows 下基于 FFmpeg 的快速视频压缩工具。

用于把本地视频快速压缩成更适合网站、博客、社交媒体发布的 MP4 文件，不暴露复杂参数，只保留三档常用规格。

## 功能

- 三档压缩规格：`small`、`standard`、`quality`（大小写不敏感）
- 默认输出到源文件旁的 `compressed` 子目录
- 永不覆盖原文件
- 支持多文件批量处理
- 支持 Windows `SendTo` 右键发送到

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

## 命令行

```powershell
videopress.exe C:\videos\clip.mp4
videopress.exe --preset small C:\videos\a.mp4 C:\videos\b.mov
videopress.exe --install-sendto
videopress.exe --uninstall-sendto
videopress.exe --version
videopress.exe -h
```

默认输出文件名格式：

```text
原文件名.<preset>.compressed.mp4
```

如果目标文件已存在，会自动追加 `-1`、`-2`。

### 退出码

- `0`：全部压缩成功
- `1`：存在失败、输入无效，或所有文件均为非视频格式被跳过

### 支持的视频格式

`.mp4` `.mov` `.mkv` `.avi` `.m4v` `.wmv` `.webm` `.ts` `.flv` `.mpg` `.mpeg` `.3gp`

## 构建

```powershell
go build -o .\bin\videopress.exe .\cmd\videopress
```

## 发布版

GitHub [Releases](https://github.com/Shirolin/videopress/releases) 提供预编译的 `videopress.exe`，无需安装 Go。仍需单独安装或放置 `ffmpeg.exe`。

## SendTo 使用

先安装：

```powershell
.\bin\videopress.exe --install-sendto
```

之后在资源管理器中右键视频文件，选择：

```text
发送到 > 快速压缩视频
```

卸载：

```powershell
.\bin\videopress.exe --uninstall-sendto
```

> 注意：SendTo 快捷方式记录的是安装时的 `videopress.exe` 绝对路径。移动可执行文件后需重新运行 `--install-sendto`。
