package cli

import "fmt"

// msgTable holds every user-facing CLI string as {zh, en}. The English half
// falls back to the Chinese one when empty, and entries may carry fmt verbs.
var msgTable = map[string][2]string{
	// 用法与帮助
	"usage.title":            {"用法: videopress.exe [选项] <视频文件...>", "Usage: videopress.exe [Options] <video files...>"},
	"usage.options":          {"选项:", "Options:"},
	"usage.number":           {"<数字>", "<number>"},
	"usage.preset":           {"压缩规格（默认 standard，大小写不敏感）", "Compression preset (default standard, case-insensitive)"},
	"usage.concurrency":      {"最大并发压缩任务数（默认 1）", "Max concurrent compression tasks (default 1)"},
	"usage.hw":               {"尝试使用 GPU 硬件加速编码", "Try to use GPU hardware acceleration"},
	"usage.force":            {"强制覆盖已存在的输出文件", "Force overwrite existing output files"},
	"usage.skip_existing":    {"如果输出文件已存在则跳过", "Skip compression if output file exists"},
	"usage.codec":            {"视频编码格式（默认 h264）", "Video encoding codec (default h264)"},
	"usage.max_fps":          {"限制最大帧率，0 为不限制（默认 0）", "Limit maximum frames per second, 0 for unlimited (default 0)"},
	"usage.audio":            {"音频模式（默认 compress）", "Audio mode (default compress)"},
	"usage.copy_audio":       {"直接复制音频流，不重编码", "Copy audio stream directly without re-encoding"},
	"usage.crf":              {"自定义视频质量 (CRF) 参数覆盖，0 为不覆盖（默认 0）", "Override CRF quality value, 0 for preset default (default 0)"},
	"usage.crop":             {"压缩前精确像素裁剪（与 --crop-center 互斥，默认不裁）", "Exact pixel crop before compression (mutually exclusive with --crop-center, off by default)"},
	"usage.crop_center":      {"压缩前居中裁到该比例的最大矩形（如 16:9）", "Center-crop to the largest rectangle of this ratio (e.g. 16:9)"},
	"usage.crop_scale":       {"在居中比例框内再保留的比例，需配合 --crop-center（默认 1.0）", "Keep scale inside the center crop, requires --crop-center (default 1.0)"},
	"usage.gif":              {"启用动图导出模式（GIF/APNG/WebP）", "Enable animated export mode (GIF/APNG/WebP)"},
	"usage.tier":             {"动图档位（默认 balanced）", "Animated tier (default balanced)"},
	"usage.format":           {"动图输出格式，可多次指定（默认全部三种）", "Animated output format, repeatable (default all three)"},
	"usage.install_sendto":   {"安装 SendTo 右键快捷方式", "Install SendTo context shortcut"},
	"usage.uninstall_sendto": {"移除 SendTo 快捷方式", "Remove SendTo shortcut"},
	"usage.install_path":     {"添加程序目录到用户 Path 环境变量", "Add app directory to user Path environment variable"},
	"usage.uninstall_path":   {"从用户 Path 环境变量移除程序目录", "Remove app directory from user Path environment variable"},
	"usage.version":          {"显示版本号", "Show version"},
	"usage.help":             {"显示此帮助", "Show this help"},
	"usage.exit_codes":       {"退出码:", "Exit Codes:"},
	"usage.exit_0":           {"全部成功或全部跳过", "All succeeded or skipped"},
	"usage.exit_1":           {"存在失败，或没有可处理的视频文件", "Failed, or no processable video files"},

	// 参数与集成错误
	"err.unknown_option":       {"未知选项:", "Unknown option:"},
	"err.add_to_path_missing":  {"未配置 AddToPath", "AddToPath is not configured"},
	"err.remove_path_missing":  {"未配置 RemoveFromPath", "RemoveFromPath is not configured"},
	"err.sendto_install_off":   {"当前构建未启用 SendTo 安装", "SendTo installation is not enabled in the current build"},
	"err.sendto_uninstall_off": {"当前构建未启用 SendTo 卸载", "SendTo uninstallation is not enabled in the current build"},
	"err.path_config_off":      {"当前构建未启用环境变量配置", "Path configuration is not enabled in the current build"},
	"err.crop_conflict":        {"裁剪模式冲突：--crop 与 --crop-center 不能同时使用", "Conflicting crop modes: --crop and --crop-center are mutually exclusive"},
	"err.crop_scale_alone":     {"裁剪参数无效：--crop-scale 仅在 --crop-center 模式下生效", "Invalid crop options: --crop-scale only applies with --crop-center"},
	"err.crop_scale_range":     {"裁剪参数无效：--crop-scale 必须在 0.5 到 1.0 之间", "Invalid crop option: --crop-scale must be between 0.5 and 1.0"},
	"err.crop_format":          {"裁剪格式无效: %s（应为 W:H:X:Y，例如 1600:900:160:90）", "Invalid crop format: %s (expected W:H:X:Y, e.g. 1600:900:160:90)"},
	"err.crop_numbers":         {"裁剪格式无效: %s（W:H:X:Y 必须为非负整数）", "Invalid crop format: %s (W:H:X:Y must be non-negative integers)"},
	"err.crop_ratio_format":    {"居中裁剪比例格式无效: %s（应为 W:H，例如 16:9）", "Invalid center-crop ratio: %s (expected W:H, e.g. 16:9)"},
	"err.crop_ratio_numbers":   {"居中裁剪比例格式无效: %s（W:H 必须为正整数）", "Invalid center-crop ratio: %s (W:H must be positive integers)"},
	"err.tier_format_ignored":  {"提示：--tier/--format 仅在 --gif 动图模式下生效，本次压缩已忽略", "Note: --tier/--format only apply in --gif animated mode, ignored for compression"},
	"err.gif_format":           {"未知的动图格式: %s（应为 gif|apng|webp）", "Unknown animated format: %s (expected gif|apng|webp)"},
	"err.engine_failed":        {"引擎运行失败: %v", "Engine failed: %v"},
	"err.summary_log":          {"无法写入摘要日志文件 %s: %v", "Could not write the summary log %s: %v"},

	// 系统集成结果
	"install.done":           {"\n处理完成。按回车键退出...", "\nDone. Press Enter to exit..."},
	"install.sendto_ok":      {"【成功】已安装 SendTo 快捷方式:", "[Success] Installed SendTo shortcut:"},
	"install.sendto_tip":     {"\n提示：现在您可以在资源管理器中右键任意视频，选择「发送到 > 快速压缩视频」进行一键压缩！", "\nTip: You can now right-click any video in Explorer and choose \"Send to > 快速压缩视频\" to compress instantly!"},
	"install.sendto_removed": {"【成功】已成功移除 SendTo 右键快捷方式。", "[Success] Successfully removed SendTo context shortcut."},
	"install.path_ok":        {"【成功】已将目录添加至当前用户的 Path 环境变量:", "[Success] Added directory to current user's Path environment variable:"},
	"install.path_tip":       {"\n提示：现在您可以在新打开的终端（CMD/PowerShell）中，直接通过 `videopress` 命令运行本软件！", "\nTip: You can now run the app by simply typing `videopress` in a newly opened terminal!"},
	"install.path_exists":    {"【提示】环境变量 Path 中已存在该目录:", "[Tip] Directory already exists in Path environment variable:"},
	"install.path_removed":   {"【成功】已成功从 Path 环境变量中移除该目录:", "[Success] Successfully removed directory from Path environment variable:"},
	"install.path_missing":   {"【提示】环境变量 Path 中未找到该目录:", "[Tip] Directory not found in Path environment variable:"},

	// 系统通知
	"notify.failed":         {"压缩失败", "Compression failed"},
	"notify.no_files":       {"未指定视频文件", "No video files specified"},
	"notify.invalid_preset": {"预设无效", "Invalid preset"},
	"notify.no_ffmpeg":      {"未找到 FFmpeg，请安装后重试", "FFmpeg not found, please install and retry"},
	"notify.done":           {"压缩完成", "Compression complete"},
	"notify.done_failures":  {"压缩完成，存在失败", "Compression finished with failures"},
	"notify.result_body":    {"成功 %d 个，失败 %d 个", "Succeeded: %d, Failed: %d"},

	// 压缩任务卡片
	"run.title":          {"         Videopress 视频压缩工具        ", "         Videopress video compressor      "},
	"run.preset":         {" 预设规格: %s\n", " Preset: %s\n"},
	"run.scale":          {" 缩放比例: %s\n", " Scale: %s\n"},
	"run.max_dim":        {" 最大分辨率限制: %s\n", " Max resolution: %s\n"},
	"run.concurrency":    {" 并发限制: %s\n", " Concurrency: %s\n"},
	"run.codec":          {" 编码格式: %s\n", " Codec: %s\n"},
	"run.codec_auto":     {"h264 (自动)", "h264 (auto)"},
	"run.crf":            {" 自定义质量 (CRF): %s\n", " Custom quality (CRF): %s\n"},
	"run.crf_overridden": {"%d (已覆盖预设)", "%d (preset overridden)"},
	"run.crf_off":        {"未启用", "off"},
	"run.crop_off":       {"未启用", "off"},
	"run.fps":            {" 帧率限制: %s\n", " Frame rate limit: %s\n"},
	"run.unlimited":      {"无限制", "unlimited"},
	"run.audio":          {" 音频模式: %s\n", " Audio: %s\n"},
	"run.audio_copy":     {"copy (直通)", "copy (pass-through)"},
	"run.audio_compress": {"compress (智能压缩)", "compress (re-encode)"},
	"run.hw":             {" 硬件加速: %s\n", " Hardware acceleration: %s\n"},
	"run.hw_disabled":    {"已禁用", "disabled"},
	"run.crop":           {" 几何裁剪: %s\n", " Crop: %s\n"},
	"run.crop_exact":     {"精确 %d:%d:%d:%d", "exact %d:%d:%d:%d"},
	"run.crop_center":    {"居中 %d:%d（保留 %.0f%%）", "center %d:%d (keep %.0f%%)"},

	// GPU 探测
	"gpu.detecting": {"正在检测可用 GPU 编码器...", "Detecting available GPU encoders..."},
	"gpu.detected":  {"检测到 GPU 编码器: %s，将启用硬件加速", "GPU encoder found: %s, enabling hardware acceleration"},
	"gpu.none":      {"未检测到可用 GPU 编码器，将使用 CPU 软编", "No GPU encoder available, falling back to CPU encoding"},

	// 逐文件日志
	"file.skip_non_video":     {"跳过非视频文件: %s\n", "Skipping non-video file: %s\n"},
	"file.unreadable":         {"警告:", "Warning:"},
	"file.unreadable_detail":  {"输入文件不存在或不可读: %s\n", "Input file missing or unreadable: %s\n"},
	"file.skip_existing":      {"跳过已存在的文件: %s\n", "Skipping existing file: %s\n"},
	"file.output_path_failed": {"错误:", "Error:"},
	"file.output_path_detail": {"生成输出路径失败 %s: %v\n", "Could not build the output path for %s: %v\n"},

	// 进度回调
	"progress.failed":    {"压缩失败:", "Compression failed:"},
	"progress.done":      {"[%s] 压缩完成\n", "[%s] done\n"},
	"progress.done_path": {"压缩完成: %s -> %s\n", "Done: %s -> %s\n"},
	"progress.start":     {"[%s] 开始压缩...\n", "[%s] compressing...\n"},

	// 摘要日志文件
	"log.start":     {"[%s] 开始压缩任务\n", "[%s] Compression run started\n"},
	"log.success":   {"成功: %s (%.1fMB -> %.1fMB, 节省 %.1f%%, 耗时 %s)\n", "Success: %s (%.1fMB -> %.1fMB, saved %.1f%%, took %s)\n"},
	"log.skipped":   {"跳过: %s (已存在)\n", "Skipped: %s (already exists)\n"},
	"log.failed":    {"失败: %s (错误信息: %s)\n", "Failed: %s (error: %s)\n"},
	"log.separator": {"----------------------------------------\n", "----------------------------------------\n"},

	// 汇总表格
	"table.file":       {"视频文件", "File"},
	"table.status":     {"状态  ", "Status"},
	"table.original":   {"原始大小", "Original"},
	"table.compressed": {"压缩大小", "Compressed"},
	"table.saved":      {"节省", "Saved"},
	"table.time":       {"耗时", "Time"},
	"table.success":    {"成功", "OK"},
	"table.skipped":    {"跳过", "Skipped"},
	"table.failed":     {"失败", "Failed"},
	"table.total":      {"处理完成: 成功 %s, 失败 %s\n", "Done: %s succeeded, %s failed\n"},

	// 动图导出
	"gif.title":        {"         Videopress 动图导出            ", "         Videopress animated export       "},
	"gif.tier":         {" 档位: %s\n", " Tier: %s\n"},
	"gif.max_width":    {" 宽度上限: %s\n", " Max width: %s\n"},
	"gif.max_size":     {" 体积上限: %s\n", " Size budget: %s\n"},
	"gif.fps":          {" 帧率: %s\n", " Frame rate: %s\n"},
	"gif.max_duration": {" 时长上限: %s\n", " Max duration: %s\n"},
	"gif.over_budget":  {" (超体积 over budget)", " (over budget)"},
	"gif.success":      {"成功", "OK"},
	"gif.skipped":      {"跳过", "Skipped"},
	"gif.failed":       {"失败", "Failed"},
	"gif.exists":       {"(已存在)", "(already exists)"},
	"gif.summary":      {"动图导出完成: 成功 %s, 跳过 %s, 失败 %s\n", "Animated export done: %s ok, %s skipped, %s failed\n"},

	// 交互式菜单
	"menu.title":          {"请选择压缩规格：", "Choose a compression preset:"},
	"menu.small":          {"小档案 (33% 比例 - 480px 限制)", "Small  (33% scale - 480px cap)"},
	"menu.standard":       {"中等   (50% 比例 - 720px 限制) [默认]", "Medium (50% scale - 720px cap) [default]"},
	"menu.quality":        {"大档案 (100% 比例 - 无限制)", "Large  (100% scale - no cap)"},
	"menu.auto_confirmed": {"已自动确认默认选项：中等", "Default confirmed automatically: Medium"},
	"menu.confirmed":      {"已选择并确认：%s", "Selected: %s"},
	"menu.hint":           {"【提示】使用上下方向键 %s 切换，%s 确认。按任意键停止倒计时。", "Tip: use %s to move, %s to confirm. Any key stops the countdown."},
	"menu.updown":         {"↑↓", "↑↓"},
	"menu.enter":          {"回车", "Enter"},
	"menu.countdown":      {"倒计时: %s 秒...   \r", "Countdown: %s s...   \r"},
	"menu.manual":         {"已启用手动选择。", "Manual selection enabled."},
}

// msg returns the copy for key in the active language, formatted with args.
func msg(key string, args ...any) string {
	entry, ok := msgTable[key]
	if !ok {
		return key
	}
	s := entry[0]
	if systemLanguage() == "en" && entry[1] != "" {
		s = entry[1]
	}
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}
