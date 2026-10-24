package gui

import "strings"

// copyTable holds every user-facing string as {zh, en}.
var copyTable = map[string][2]string{
	"nav.workspace": {"主工作区", "Workspace"},
	"nav.settings":  {"配置面板", "Settings"},

	"dropzone.title":     {"拖入视频文件", "Drop video files here"},
	"dropzone.sub":       {"或点击选择", "or click to select"},
	"dropzone.add_more":  {"点击或拖入视频以添加", "Click or drop videos to add"},
	"dropzone.desc":      {"支持多选，或点击选择文件", "Multi-select supported, or click to choose files"},
	"dropzone.formats":   {"支持 MP4, MOV, MKV, AVI, WEBM 等视频格式", "Supports MP4, MOV, MKV, AVI, WEBM, etc."},
	"dropzone.disabled":  {"压缩进行中，请稍候", "Compression in progress, please wait"},
	"dropzone.drag_hint": {"松开即可添加", "Release to add"},
	"dropzone.browse":    {"选择文件", "Choose files"},

	"stats.total":          {"任务总数", "Total Tasks"},
	"stats.processing":     {"处理中", "Processing"},
	"stats.idle":           {"待命", "Idle"},
	"stats.completed":      {"完成", "Completed"},
	"stats.saved_space":    {"共节省空间", "Space Saved"},
	"stats.avg_ratio":      {"平均压缩率", "Avg. Ratio"},
	"stats.current_status": {"当前状态", "Status"},

	"controls.preset":              {"压缩预设", "Preset"},
	"controls.output":              {"保存目录", "Output Folder"},
	"controls.default_output_path": {"默认 (原视频目录下的 compressed/ 文件夹)", "Default (compressed/ folder under the source directory)"},
	"controls.default_output_tip":  {"原视频同目录下的 compressed/ 文件夹", "compressed/ folder under the source video directory"},

	"preset.small":    {"小文件", "Small"},
	"preset.standard": {"标准", "Standard"},
	"preset.quality":  {"高画质", "High Quality"},

	"action.summary": {"已选 {num} 个 · {preset}", "{num} selected · {preset}"},
	"action.empty":   {"选择文件后可开始压缩", "Choose files to start compression"},

	"btn.change":             {"更改", "Change"},
	"btn.reset":              {"重置", "Reset"},
	"btn.reset_path_title":   {"恢复默认路径", "Reset output path"},
	"btn.open_output_folder": {"打开输出文件夹", "Open Output Folder"},
	"btn.cancel_compress":    {"取消当前任务", "Cancel Task"},
	"btn.start_compress":     {"开始压缩", "Start Compression"},
	"btn.enable":             {"一键开启", "Enable"},
	"btn.disable":            {"卸载移除", "Remove"},

	"app.ffmpeg_not_found": {"未检测到 FFmpeg，请先完成核心引擎配置。", "FFmpeg not found; finish the core engine setup first."},

	"queue.title":                  {"任务队列", "Task Queue"},
	"queue.clear":                  {"清空队列", "Clear Queue"},
	"queue.empty":                  {"暂无待压缩的视频任务", "No video tasks in queue"},
	"queue.status_idle":            {"待命中", "Idle"},
	"queue.status_waiting":         {"等待中", "Waiting"},
	"queue.status_succeeded":       {"压缩成功", "Succeeded"},
	"queue.status_failed":          {"压缩失败", "Failed"},
	"queue.status_skipped":         {"已跳过", "Skipped"},
	"queue.status_compressing":     {"压缩中", "Compressing"},
	"queue.open_folder":            {"打开所在目录", "Open Folder"},
	"queue.remove":                 {"移除", "Remove"},
	"queue.percent_done":           {"100% (完成)", "100% (Done)"},
	"queue.compress_failed_simple": {"失败", "Failed"},
	"queue.skip_tip":               {"已跳过：输出目录已存在同名文件", "Skipped: output file already exists"},
	"queue.origin_size":            {"原始", "Source"},
	"queue.target_size":            {"压缩后", "Compressed"},
	"queue.save_ratio":             {"压缩率", "Ratio"},
	"queue.duration":               {"耗时", "Time"},
	"queue.warn_recompressed":      {"该视频可能已经被压缩过，二次压缩会损害画质。", "This video may already be compressed; compressing again will hurt quality."},

	"settings.title.general":      {"常规预设与性能", "Presets & Performance"},
	"settings.title.sys":          {"Windows 系统集成", "Windows System Integration"},
	"settings.title.lang":         {"界面语言", "Interface Language"},
	"settings.title.debug":        {"调试与排查日志", "Debug & Diagnostic Logs"},
	"settings.section.preset":     {"压缩预设", "Compression Preset"},
	"settings.section.perf":       {"性能与行为", "Performance & Behavior"},
	"settings.preset_desc":        {"预设了不同的码率与分辨率配置", "Presets for target bitrate and resolution"},
	"settings.concurrency":        {"并发任务数", "Concurrency"},
	"settings.concurrency_desc":   {"同时压缩的视频数量", "How many videos compress at once"},
	"settings.concurrency_all":    {"{num}（全部）", "{num} (all)"},
	"settings.hw_accel":           {"硬件加速", "Hardware Acceleration"},
	"settings.hw_accel_active":    {"已启用显卡加速: {gpu}", "GPU acceleration enabled: {gpu}"},
	"settings.hw_accel_off":       {"无显卡加速，回退至 CPU 编码", "No GPU acceleration, falling back to CPU encoding"},
	"settings.hw_accel_probing":   {"正在探测显卡编码器…", "Probing GPU encoders…"},
	"settings.copy_audio":         {"只复制音频流", "Copy audio stream"},
	"settings.copy_audio_desc":    {"直接复制原片音轨，节省音频重编码开销", "Copy the original track to skip audio re-encoding"},
	"settings.force_mode":         {"强制覆盖同名文件", "Overwrite existing files"},
	"settings.force_mode_desc":    {"直接重写已有的同名输出文件，不生成序号副本", "Rewrite an existing output instead of numbering a copy"},
	"settings.skip_existing":      {"增量跳过模式", "Skip existing"},
	"settings.skip_existing_desc": {"输出文件夹已有同名压缩文件时跳过", "Skip when the output folder already holds that file"},
	"settings.lang.desc":          {"选择显示语言，右键菜单等集成项会在切换后自动热更新。", "Pick the display language; system integration menus follow the change."},
	"settings.debug.desc":         {"开启后记录 GPU 探测错误、各模块耗时及底层 FFmpeg 出错堆栈。", "Records GPU probe errors, module timings and FFmpeg failure details."},
	"settings.version":            {"版本", "Version"},
	"settings.debug.manage":       {"日志操作", "Log actions"},
	"settings.lang.display":       {"显示语言", "Display language"},
	"settings.debug.toggle":       {"环境检测与探测性能日志", "Environment and probe diagnostics"},
	"settings.debug.file_desc":    {"在默认编辑器中打开当前日志，或清空积累的探测日志及硬件加速缓存。", "Open the log in your editor, or clear accumulated logs and the GPU cache."},
	"btn.open_log":                {"打开日志文件", "Open log file"},
	"btn.clear_log":               {"清空日志与缓存", "Clear logs & cache"},

	"integration.contextmenu":      {"右键菜单压缩", "Right-click compression"},
	"integration.contextmenu_desc": {"在资源管理器右键视频文件，选择「使用 Videopress 压缩」。", "Right-click any video in Explorer and pick “Compress with Videopress”."},
	"integration.sendto":           {"发送到菜单 (SendTo)", "SendTo menu"},
	"integration.sendto_desc":      {"右键文件 → 发送到 → 快速压缩视频。", "Right-click a file → Send to → compress it."},
	"integration.desktop":          {"桌面快捷方式", "Desktop shortcut"},
	"integration.desktop_desc":     {"在桌面上创建 Videopress 的启动方式。", "Put a Videopress launcher on the desktop."},
	"integration.startmenu":        {"开始菜单", "Start menu"},
	"integration.startmenu_desc":   {"加入开始菜单程序列表，可在搜索中唤醒。", "Add Videopress to the Start menu program list."},
	"integration.path":             {"用户 Path 环境变量", "User Path variable"},
	"integration.path_desc":        {"把程序所在目录加入 Path，可在任意终端直接运行。", "Add the program folder to Path so any terminal can run it."},

	"status.enabled":       {"已开启", "Enabled"},
	"status.disabled":      {"已关闭", "Disabled"},
	"status.checking":      {"检测中…", "Checking…"},
	"status.adding":        {"正在创建…", "Creating…"},
	"status.removing":      {"正在移除…", "Removing…"},
	"status.up_to_date":    {"已是最新", "Up to date"},
	"status.not_installed": {"未安装", "Not installed"},

	"setup.title":        {"需要配置核心视频引擎", "Video Engine Setup Required"},
	"setup.desc":         {"未检测到 FFmpeg 视频引擎。Videopress 需要它进行视频解码与压缩。", "FFmpeg was not found. Videopress needs it to decode and compress video."},
	"setup.downloading":  {"正在下载核心组件… {percent}%", "Downloading core components… {percent}%"},
	"setup.download_tip": {"请勿关闭软件，完成后将自动进入主界面。", "Do not close the app; the workspace opens when this finishes."},
	"setup.btn_auto":     {"自动下载并配置（推荐）", "Download and configure (recommended)"},
	"setup.btn_manual":   {"查看系统集成设置", "Open system integration settings"},
	"setup.failed":       {"下载或解压 FFmpeg 失败，请检查网络后重试。", "Downloading or unpacking FFmpeg failed. Check your network and retry."},

	"advanced.title":               {"高级参数", "Advanced"},
	"advanced.enabled_indicator":   {"已启用自定义高级配置", "Custom settings active"},
	"advanced.video_codec":         {"视频编码格式", "Video codec"},
	"advanced.video_codec_auto":    {"自动（按硬件推荐）", "Auto (hardware-based)"},
	"advanced.video_codec_desc":    {"自动按硬件推荐编码器；H.265 体积更小，AV1 为较新编码。", "Auto picks an encoder for your hardware. H.265 yields smaller files; AV1 is newer."},
	"advanced.max_fps":             {"帧率限制", "Frame rate limit"},
	"advanced.max_fps_auto":        {"不限制（保留原片帧率）", "Unlimited (keep original)"},
	"advanced.max_fps_desc":        {"限制至 30fps 可大幅削减体积，适合录屏压缩。", "Capping at 30fps shrinks recordings a lot."},
	"advanced.audio_mode":          {"音频处理", "Audio"},
	"advanced.audio_mode_auto":     {"压缩音频（推荐）", "Compress audio (recommended)"},
	"advanced.audio_mode_copy":     {"音频直通（无损保留）", "Pass through (lossless)"},
	"advanced.audio_mode_mute":     {"视频静音（去除音轨）", "Mute (drop the track)"},
	"advanced.audio_mode_desc":     {"直通保留原音轨，压缩可减小体积。", "Pass-through keeps the track; compressing saves space."},
	"advanced.crf.title":           {"自定义质量 (CRF)", "Custom quality (CRF)"},
	"advanced.crf.desc.h264":       {"H.264 建议 18 - 28。数值越小画质越好，体积越大。", "H.264 suggested 18–28. Lower is better quality and larger."},
	"advanced.crf.desc.h265":       {"H.265 建议 20 - 30。HEVC 能以更大数值保留优质画面。", "H.265 suggested 20–30. HEVC holds quality at higher values."},
	"advanced.crf.desc.av1":        {"AV1 建议 24 - 38。AV1 在高数值下依然保持卓越画质。", "AV1 suggested 24–38. AV1 stays sharp at higher values."},
	"advanced.crf.preset_override": {"分辨率预设 (画质已自定义)", "Resolution preset (quality overridden)"},
	"advanced.crf.status.high":     {"超清/体积大", "Ultra / large"},
	"advanced.crf.status.balanced": {"清晰/均衡", "Balanced"},
	"advanced.crf.status.low":      {"高压缩/体积小", "Compressed / small"},
	"advanced.crop.title":          {"几何裁剪", "Geometry crop"},
	"advanced.crop.desc":           {"压缩前先裁掉画面多余空边（如 AI 生成视频的安全边）。默认不裁。", "Trim excess padding before compressing. Off by default."},
	"advanced.crop.mode":           {"裁剪方式", "Crop mode"},
	"advanced.crop.mode_off":       {"不裁剪", "Off"},
	"advanced.crop.mode_exact":     {"精确像素 (W:H:X:Y)", "Exact pixels (W:H:X:Y)"},
	"advanced.crop.mode_center":    {"居中比例 (W:H)", "Centered ratio (W:H)"},
	"advanced.crop.exact":          {"精确裁剪", "Exact crop"},
	"advanced.crop.center":         {"居中裁剪", "Center crop"},
	"advanced.crop.scale":          {"保留比例", "Keep scale"},
	"advanced.crop.scale_desc":     {"在居中比例框内再保留的画面比例。", "Fraction of the centered frame to keep."},
	"advanced.crop.size_label":     {"宽:高", "W:H"},
	"advanced.crop.offset_label":   {"X:Y 偏移", "X:Y offset"},

	"gif.title":              {"动图导出", "Animated export"},
	"gif.subtitle":           {"为社媒分享导出小体积动图", "Small animated clips for sharing"},
	"gif.tier":               {"档位", "Tier"},
	"gif.tier_smooth":        {"流畅", "Smooth"},
	"gif.tier_balanced":      {"均衡", "Balanced"},
	"gif.tier_hd":            {"高清", "HD"},
	"gif.spec":               {"≤%dMB · %dpx · %dfps · %s", "≤%dMB · %dpx · %dfps · %s"},
	"gif.tier_desc_smooth":   {"小体积流畅档，适合极速分享", "Smallest tier, for the fastest sharing"},
	"gif.tier_desc_balanced": {"均衡档（默认模板），画质与体积平衡", "Balanced (default), quality against size"},
	"gif.tier_desc_hd":       {"高清档，保留更多细节", "HD tier, keeps more detail"},
	"gif.format":             {"输出格式", "Formats"},
	"gif.export_btn":         {"导出动图", "Export animated"},
	"gif.cancel":             {"取消导出", "Cancel export"},
	"gif.need_format":        {"请至少选择一种输出格式", "Select at least one format"},
	"gif.ok":                 {"成功", "OK"},
	"gif.failed":             {"失败", "Failed"},
	"gif.skipped":            {"跳过", "Skipped"},
	"gif.over_budget":        {"超体积", "Over budget"},
	"gif.pending":            {"将为 {n} 个文件各导出 {m} 种格式", "Exports {m} formats for each of {n} files"},
	"gif.busy_note":          {"压缩进行中，完成后可导出动图", "Compression running; export when it finishes"},

	"toast.log_opened":          {"已打开本地调试与性能排查日志", "Opened the local diagnostic log"},
	"toast.log_open_failed":     {"打开日志文件失败: {err}", "Failed to open the log: {err}"},
	"toast.log_cleared":         {"已清空日志并重置硬件加速缓存", "Cleared the log and the GPU cache"},
	"toast.log_clear_failed":    {"清空日志与缓存失败: {err}", "Failed to clear logs and cache: {err}"},
	"toast.action_failed":       {"操作失败: {err}", "Action failed: {err}"},
	"toast.contextmenu_added":   {"已把「使用 Videopress 压缩」加入右键菜单", "Added “Compress with Videopress” to the context menu"},
	"toast.contextmenu_removed": {"已卸载右键压缩菜单", "Removed the right-click compression menu"},
	"toast.sendto_added":        {"已创建 SendTo 快捷方式", "Created the SendTo shortcut"},
	"toast.sendto_removed":      {"已移除 SendTo 快捷方式", "Removed the SendTo shortcut"},
	"toast.desktop_added":       {"已创建桌面快捷方式", "Created the desktop shortcut"},
	"toast.desktop_removed":     {"已删除桌面快捷方式", "Removed the desktop shortcut"},
	"toast.startmenu_added":     {"已加入开始菜单", "Added to the Start menu"},
	"toast.startmenu_removed":   {"已从开始菜单移除", "Removed from the Start menu"},
	"toast.path_added":          {"已把目录加入用户 Path", "Added the folder to the user Path"},
	"toast.path_removed":        {"已从用户 Path 移除目录", "Removed the folder from the user Path"},
	"toast.done":                {"压缩完成", "Compression finished"},
	"toast.cancelled":           {"已取消", "Cancelled"},
	"toast.gpu_ready":           {"显卡加速可用: {gpu}", "GPU acceleration available: {gpu}"},
}

// t returns the copy for key in the active language.
func (a *App) t(key string) string {
	return pick(copyTable[key], a.lang() == "zh")
}

// tf returns the copy for key with {name} placeholders filled from vals.
func (a *App) tf(key string, vals map[string]string) string {
	s := a.t(key)
	for k, v := range vals {
		s = strings.ReplaceAll(s, "{"+k+"}", v)
	}
	return s
}

// pick chooses the Chinese or English half of a copy entry.
func pick(entry [2]string, zh bool) string {
	if zh {
		if entry[0] != "" {
			return entry[0]
		}
		return entry[1]
	}
	if entry[1] != "" {
		return entry[1]
	}
	return entry[0]
}
