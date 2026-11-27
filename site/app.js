(function () {
  'use strict';

  var STORAGE_THEME = 'videopress-theme';
  var STORAGE_LANG = 'videopress-lang';

  /* ==========================================================================
     Bilingual Copy Dictionary (zh / en)
     ========================================================================== */
  var COPY = {
    zh: {
      'meta.title': 'Videopress — Windows 视频压缩与动图导出工具',
      'meta.desc': '面向 Windows 的本地 FFmpeg 视频压缩工具，支持桌面界面、命令行、并发队列、GPU 硬件加速与 GIF/APNG/WebP 动图导出。',
      'a11y.skip': '跳到主要内容',
      'a11y.nav': '页面导航',
      'a11y.theme_group': '主题切换',
      'a11y.lang_group': '语言切换',
      'a11y.menu_open': '打开导航菜单',
      'a11y.menu_close': '关闭导航菜单',
      'a11y.facts': '基本规格',
      'a11y.showcase_tabs': '界面场景切换',
      'a11y.dec_concurrency': '减少并发数',
      'a11y.inc_concurrency': '增加并发数',
      'a11y.generated_cmd': '生成的命令行',

      'nav.features': '参数与能力',
      'nav.interface': '界面预览',
      'nav.cli': '命令行',
      'nav.integration': '系统集成',
      'nav.download': '下载',
      'theme.system': '跟随系统',
      'theme.light': '浅色',
      'theme.dark': '深色',
      'lang.zh': '中文',
      'lang.en': 'EN',

      'hero.title': '面向 Windows 的本地视频压缩工具',
      'hero.subtitle': '基于 FFmpeg，单文件集成原生桌面界面与命令行。提供三档实测预设、并发队列、GPU 硬件编码、几何裁剪以及 GIF / APNG / WebP 动图导出。',
      'hero.cta_download': '前往 Releases 下载',
      'hero.cta_demo': '观看演示 (0:26)',
      'hero.cta_cli': '命令行生成器',
      'hero.spec.os': 'Windows 10 / 11 (x64)',
      'hero.spec.dual': 'GUI + CLI 单文件',
      'hero.spec.engine': '调用本机 FFmpeg',
      'hero.spec.license': 'MIT 开源许可',
      'hero.alt_dark': 'Videopress 主工作区（深色主题）',
      'hero.alt_light': 'Videopress 配置面板与外观设置（浅色主题）',

      'specs.title': '预设参数与编码能力',
      'specs.subtitle': '所有预设对应确定的缩放、CRF 与音频码率参数，无需记忆 FFmpeg 命令，也支持在高级面板中逐项覆盖。',
      'specs.presets_heading': '内置三档压缩预设',
      'specs.col.preset': '预设',
      'specs.col.scale': '缩放比例',
      'specs.col.max_dim': '分辨率上限',
      'specs.col.crf': '默认 CRF',
      'specs.col.encoder_preset': '编码速度档',
      'specs.col.audio': '音频码率 (AAC)',
      'specs.col.usecase': '适用场景',
      'specs.preset.small_name': '小文件 (small)',
      'specs.preset.small_use': '聊天发送、低流量预览',
      'specs.preset.standard_name': '标准 (standard · 默认)',
      'specs.preset.standard_use': '网页嵌入、文档附带与日常归档',
      'specs.preset.quality_name': '高画质 (quality)',
      'specs.preset.quality_use': '保留原始分辨率的长期存档',
      'specs.no_cap': '保持原分辨率',
      'specs.mp4_note': '默认输出 MP4（附加 -movflags +faststart），保存于源视频同目录的 compressed/ 文件夹。',

      'specs.gpu_heading': 'GPU 硬件加速与编码器',
      'specs.gpu_desc': '按 NVENC → QSV → AMF 顺序并发探测本机显卡编码器（单项超时 1.5s），结果缓存至本地，不可用时自动回退 CPU 软编。',
      'specs.gpu.col_codec': '视频编码',
      'specs.gpu.col_hw': '硬件编码器',
      'specs.gpu.col_cpu': 'CPU 回退',
      'specs.gpu.col_crf': '建议 CRF',

      'specs.adv_heading': '高级参数与前处理',
      'specs.adv.crf_title': '自定义质量 (CRF)',
      'specs.adv.crf_desc': '可覆盖预设默认值，按编码格式提供建议参考区间，硬件编码自动映射至 QP / global_quality。',
      'specs.adv.audio_title': '音频三模式与帧率上限',
      'specs.adv.audio_desc': '支持 AAC 重编码、音频流无损直通（copy）或静音移除音轨（mute），并可限制最大帧率（如 30fps）缩减录屏体积。',
      'specs.adv.crop_title': '可选几何裁剪',
      'specs.adv.crop_desc': '在缩放前裁除画面多余边栏，支持精确像素（W:H:X:Y）或居中比例（如 16:9 配合 0.5–1.0 保留比例），默认关闭。',

      'showcase.title': '真实工作区与动图导出',
      'showcase.subtitle': '基于 MyGO 与 MujicaUI 纯 Go 自绘，无 WebView 开销；遵循同一套酒红与象牙配色，默认跟随系统明暗。',
      'showcase.sidebar_label': '界面视角',
      'showcase.tab.demo': '宣传片演示',
      'showcase.tab.workspace': '主工作区',
      'showcase.tab.batch': '批量进度',
      'showcase.tab.gif': '动图自定义',
      'showcase.tab.theme': '外观主题',
      'showcase.cap.demo': '26 秒完整功能演示（含配乐与音效）：涵盖批量队列、GPU 硬编探测、动图导出、几何裁剪与 CLI 终端工作流。',
      'showcase.cap.workspace': '拖放添加任务，顶部汇总任务数、已节省空间与整体压缩率；队列支持按全部、失败、已完成筛选。',
      'showcase.cap.batch': '多任务并发执行时显示整体进度条、单文件编码倍速（如 2.4x）与预计剩余时间，支持逐个取消单个文件而不中断其余队列。',
      'showcase.cap.gif': '内置流畅、均衡、高清三档与自定义参数，可同时勾选导出 GIF、APNG、WebP，并自动读取源视频时长提示截断与超体积。',
      'showcase.cap.theme': '配置面板按功能分区，外观支持跟随系统或固定浅色 / 深色，界面语言切换后同步更新中英文与系统右键菜单名称。',
      'showcase.video_unsupported': '当前浏览器不支持内嵌 MP4 播放，可直接访问视频链接观看。',
      'showcase.alt.workspace': 'Videopress 主工作区截图',
      'showcase.alt.batch': 'Videopress 批量压缩进度截图',
      'showcase.alt.gif': 'Videopress 动图自定义导出界面截图',
      'showcase.alt.theme': 'Videopress 配置面板与外观设置截图',

      'gif.table_heading': '动图导出规格',
      'gif.formats_note': '支持同时导出 GIF（128 色调色板）、APNG（无损 RGBA 自适应帧率）与 WebP（libwebp_anim）。',
      'gif.col.tier': '档位',
      'gif.col.width': '宽度上限',
      'gif.col.budget': '体积目标',
      'gif.col.fps': '帧率',
      'gif.col.duration': '片段时长',
      'gif.col.webp_q': 'WebP 质量',
      'gif.tier.smooth': '流畅 (smooth)',
      'gif.tier.balanced': '均衡 (balanced · 默认)',
      'gif.tier.hd': '高清 (hd)',
      'gif.tier.custom': '自定义 (GUI)',
      'gif.custom_dur': '1–30s（起点 0–600s）',

      'cli.title': '图形界面与命令行共用同一程序',
      'cli.subtitle': '双击或无选项启动进入图形界面；在终端传入任意选项即进入控制台模式，输出进度条、压缩汇总表，并在输出目录追加 compress_summary.log。',
      'cli.builder_title': '命令行生成器',
      'cli.builder_desc': '勾选所需参数，实时生成可直接运行的 videopress.exe 命令。',
      'cli.lbl.preset': '压缩预设',
      'cli.lbl.codec': '视频编码',
      'cli.lbl.concurrency': '并发任务数 (-c)',
      'cli.lbl.hw': '启用 GPU 硬件加速 (--hw)',
      'cli.lbl.audio': '音频处理',
      'cli.opt.preset_small': '小文件',
      'cli.opt.preset_standard': '标准',
      'cli.opt.preset_quality': '高画质',
      'cli.opt.codec_auto': '默认 (H.264)',
      'cli.opt.audio_compress': '重编码 (默认)',
      'cli.opt.audio_copy': '直通 (-a)',
      'cli.opt.audio_mute': '静音 (--audio mute)',
      'cli.lbl.crop': '居中比例裁剪',
      'cli.opt.crop_off': '不裁剪',
      'cli.lbl.crop_scale': '保留画面比例 (--crop-scale)',
      'cli.lbl.conflict': '同名文件策略',
      'cli.opt.conflict_number': '自动编号 (默认)',
      'cli.opt.conflict_force': '强制覆盖 (--force)',
      'cli.opt.conflict_skip': '增量跳过 (--skip-existing)',
      'cli.state.on': '开启',
      'cli.state.off': '关闭',
      'cli.btn.copy': '复制命令',
      'cli.btn.reset': '重置选项',
      'cli.toast.copied': '已复制到剪贴板',
      'cli.toast.reset': '已重置命令行选项',
      'cli.examples_title': '常用命令参考',
      'cli.ex1_label': '启用 GPU 加速与音频直通，按 small 预设压缩',
      'cli.ex2_label': '居中裁至 16:9 并保留 90% 画面后压缩',
      'cli.ex3_label': '3 并发批量处理并跳过已存在输出',

      'sys.title': 'Windows 系统集成与 FFmpeg 准备',
      'sys.subtitle': '所有系统入口均只写入当前用户配置，无需管理员权限；右键多选视频时自动合并到同一窗口队列。',
      'sys.int_heading': '五项当前用户集成',
      'sys.int.ctx_title': '右键菜单压缩',
      'sys.int.ctx_desc': '资源管理器右键视频选择「使用 Videopress 压缩」，切换界面语言时自动同步菜单名称。',
      'sys.int.sendto_title': '发送到菜单 (SendTo)',
      'sys.int.sendto_desc': '右键 → 发送到 →「快速压缩视频」，完成后弹出系统通知。',
      'sys.int.shortcuts_title': '桌面与开始菜单快捷方式',
      'sys.int.shortcuts_desc': '在配置面板按需开关，方便搜索启动或拖放图标。',
      'sys.int.path_title': '用户 Path 环境变量',
      'sys.int.path_desc': '将程序目录加入当前用户 Path，随后可在任意终端直接调用 videopress。',
      'sys.col.ffmpeg_title': 'FFmpeg 查找顺序与自动配置',
      'sys.col.ffmpeg_1': '首次启动若未检测到 FFmpeg，可在引导页点击「自动下载并配置」，从 npmmirror 获取 ffmpeg-static v6.0（约 28 MB），解压并校验可执行性。',
      'sys.col.ffmpeg_2': '与 Videopress.exe 同目录的 ffmpeg.exe（适合离线环境随身使用）。',
      'sys.col.ffmpeg_3': '系统 PATH 中的 ffmpeg.exe。',
      'sys.col.ffmpeg_4': '在引导页或「配置面板 → 关于」手动指定 ffmpeg.exe 路径，随时可切回自动检测。',
      'sys.col.cli_int_title': '命令行开关系统集成',
      'sys.col.cli_int_desc': '除配置面板外，也可直接通过终端管理 SendTo 与用户 Path：',
      'sys.col.formats_title': '支持的输入格式与日志归档',
      'sys.col.formats_body': '支持 MP4、MOV、MKV、AVI、M4V、WMV、WEBM、TS、FLV、MPG、MPEG、3GP。每次批处理会在输出目录追加 compress_summary.log，记录原始大小、压缩后大小、节省比例与耗时。',

      'dl.title': '获取 Videopress',
      'dl.subtitle': '在 GitHub Releases 提供安装版与便携版两种分发包，均附 SHA256SUMS.txt 校验文件。',
      'dl.installer.title': '安装版',
      'dl.installer.file': 'videopress-setup-<版本>-amd64.exe',
      'dl.installer.desc': '包含开始菜单入口与标准卸载流程，适合日常固定使用。',
      'dl.installer.steps': '双击安装 → 从开始菜单启动 → 按引导自动配置或指定 FFmpeg。',
      'dl.portable.title': '便携版',
      'dl.portable.file': 'Videopress.exe',
      'dl.portable.desc': '单文件可执行程序，无需安装或管理员权限，适合放工具目录或移动硬盘。',
      'dl.portable.steps': '放入独立目录直接运行；离线机器可将 ffmpeg.exe 置于同目录。',
      'dl.btn.releases': '前往 GitHub Releases',
      'dl.col.smartscreen_title': '关于 Windows SmartScreen 提示与校验',
      'dl.col.smartscreen_body': '当前发布包暂未进行代码签名，首次运行可能弹出 SmartScreen「Windows 已保护你的电脑」，点击「更多信息 → 仍要运行」即可。可使用 PowerShell 运行 Get-FileHash 核对下载页附带的 SHA256SUMS.txt。',
      'dl.col.req_title': '系统要求与本地配置路径',
      'dl.col.req_body': '支持 Windows 10 / 11（x64）。用户偏好保存在 %APPDATA%\\Videopress\\settings.json，GPU 探测缓存保存在 %LOCALAPPDATA%\\videopress_gpu.cache。',

      'footer.tagline': 'Windows 本地视频压缩与动图导出工具',
      'footer.license': 'MIT License · Copyright © 2026 Shirolin',
      'footer.link.repo': 'GitHub 仓库',
      'footer.link.releases': 'Releases'
    },

    en: {
      'meta.title': 'Videopress — Video Compression & Animated Export for Windows',
      'meta.desc': 'Local FFmpeg video compression tool for Windows with native GUI, CLI, concurrent queue, GPU hardware encoding, and GIF/APNG/WebP animated export.',
      'a11y.skip': 'Skip to main content',
      'a11y.nav': 'Primary navigation',
      'a11y.theme_group': 'Appearance theme',
      'a11y.lang_group': 'Display language',
      'a11y.menu_open': 'Open navigation menu',
      'a11y.menu_close': 'Close navigation menu',
      'a11y.facts': 'Key specifications',
      'a11y.showcase_tabs': 'Interface views',
      'a11y.dec_concurrency': 'Decrease concurrency',
      'a11y.inc_concurrency': 'Increase concurrency',
      'a11y.generated_cmd': 'Generated command',

      'nav.features': 'Specs',
      'nav.interface': 'Interface',
      'nav.cli': 'CLI',
      'nav.integration': 'Integration',
      'nav.download': 'Download',
      'theme.system': 'System',
      'theme.light': 'Light',
      'theme.dark': 'Dark',
      'lang.zh': '中文',
      'lang.en': 'EN',

      'hero.title': 'Local video compression for Windows',
      'hero.subtitle': 'Built on FFmpeg with both a native desktop GUI and CLI in one binary. Includes three calibrated presets, concurrent batch queues, GPU hardware encoding, geometry cropping, and GIF / APNG / WebP export.',
      'hero.cta_download': 'Download from Releases',
      'hero.cta_demo': 'Watch Demo (0:26)',
      'hero.cta_cli': 'CLI Builder',
      'hero.spec.os': 'Windows 10 / 11 (x64)',
      'hero.spec.dual': 'Single-binary GUI + CLI',
      'hero.spec.engine': 'Powered by local FFmpeg',
      'hero.spec.license': 'MIT License',
      'hero.alt_dark': 'Videopress Workspace (Dark theme)',
      'hero.alt_light': 'Videopress Settings & Appearance (Light theme)',

      'specs.title': 'Presets & Encoding Specs',
      'specs.subtitle': 'Each preset maps to explicit scale, CRF, and audio bitrate settings—no FFmpeg flags to memorize, with per-field overrides in Advanced.',
      'specs.presets_heading': 'Built-in Compression Presets',
      'specs.col.preset': 'Preset',
      'specs.col.scale': 'Scale',
      'specs.col.max_dim': 'Max Dimension',
      'specs.col.crf': 'Default CRF',
      'specs.col.encoder_preset': 'Encoder Preset',
      'specs.col.audio': 'Audio (AAC)',
      'specs.col.usecase': 'Typical Use',
      'specs.preset.small_name': 'Small (small)',
      'specs.preset.small_use': 'Messaging and low-bandwidth previews',
      'specs.preset.standard_name': 'Standard (standard · Default)',
      'specs.preset.standard_use': 'Web publishing, docs, and daily sharing',
      'specs.preset.quality_name': 'High Quality (quality)',
      'specs.preset.quality_use': 'Archival copies keeping native resolution',
      'specs.no_cap': 'Original resolution',
      'specs.mp4_note': 'Outputs MP4 with -movflags +faststart into a compressed/ folder beside each source video.',

      'specs.gpu_heading': 'GPU Hardware Encoding & Codecs',
      'specs.gpu_desc': 'Concurrently probes NVENC → QSV → AMF encoders (1.5s timeout each), caches the result locally, and falls back to CPU encoding when unavailable.',
      'specs.gpu.col_codec': 'Codec',
      'specs.gpu.col_hw': 'Hardware Encoders',
      'specs.gpu.col_cpu': 'CPU Fallback',
      'specs.gpu.col_crf': 'Suggested CRF',

      'specs.adv_heading': 'Advanced Controls',
      'specs.adv.crf_title': 'Custom Quality (CRF)',
      'specs.adv.crf_desc': 'Override the preset CRF with per-codec guidance; hardware encoders map the value to QP or global_quality.',
      'specs.adv.audio_title': 'Audio Modes & Frame Rate Cap',
      'specs.adv.audio_desc': 'Choose AAC compression, lossless pass-through (copy), or muted output (mute), plus optional frame-rate caps such as 30fps for screen recordings.',
      'specs.adv.crop_title': 'Optional Geometry Crop',
      'specs.adv.crop_desc': 'Trim excess borders before scaling via exact pixels (W:H:X:Y) or centered aspect ratio (e.g. 16:9 with 0.5–1.0 keep scale). Off by default.',

      'showcase.title': 'Workspace & Animated Export',
      'showcase.subtitle': 'Drawn natively in Go with MyGO and MujicaUI—no WebView overhead. Uses one burgundy-and-ivory theme that follows system appearance by default.',
      'showcase.sidebar_label': 'Views',
      'showcase.tab.demo': 'Video Demo',
      'showcase.tab.workspace': 'Workspace',
      'showcase.tab.batch': 'Batch Progress',
      'showcase.tab.gif': 'Animated Export',
      'showcase.tab.theme': 'Appearance',
      'showcase.cap.demo': '26-second full walkthrough with soundtrack: batch queue, GPU encoder probing, animated export, geometry crop, and CLI workflow.',
      'showcase.cap.workspace': 'Drop files to queue them. The top summary tracks task counts, space saved, and overall ratio, with All / Failed / Done queue filters.',
      'showcase.cap.batch': 'Concurrent batches display overall progress, per-file speed (e.g. 2.4x), and ETA, with per-file cancellation that leaves the rest of the queue running.',
      'showcase.cap.gif': 'Includes Smooth, Balanced, HD, and Custom tiers to export GIF, APNG, and WebP together, checking source duration for truncation and size budgets.',
      'showcase.cap.theme': 'Settings are organized by section. Pin Light or Dark or follow the system, and switch Chinese/English with live updates to the context menu label.',
      'showcase.video_unsupported': 'Your browser cannot play inline MP4 video; open the direct link instead.',
      'showcase.alt.workspace': 'Videopress Workspace screenshot',
      'showcase.alt.batch': 'Videopress batch compression progress screenshot',
      'showcase.alt.gif': 'Videopress custom animated export panel screenshot',
      'showcase.alt.theme': 'Videopress settings and appearance panel screenshot',

      'gif.table_heading': 'Animated Export Tiers',
      'gif.formats_note': 'Exports GIF (128-color palette), APNG (lossless RGBA with adaptive FPS), and WebP (libwebp_anim) in one pass.',
      'gif.col.tier': 'Tier',
      'gif.col.width': 'Max Width',
      'gif.col.budget': 'Size Budget',
      'gif.col.fps': 'Frame Rate',
      'gif.col.duration': 'Duration',
      'gif.col.webp_q': 'WebP Quality',
      'gif.tier.smooth': 'Smooth (smooth)',
      'gif.tier.balanced': 'Balanced (balanced · Default)',
      'gif.tier.hd': 'HD (hd)',
      'gif.tier.custom': 'Custom (GUI)',
      'gif.custom_dur': '1–30s (start 0–600s)',

      'cli.title': 'One Binary for GUI and CLI',
      'cli.subtitle': 'Launch without flags for the desktop window; pass any option in a terminal to run headless with progress output, a summary table, and compress_summary.log.',
      'cli.builder_title': 'Command Builder',
      'cli.builder_desc': 'Select options below to assemble a ready-to-run videopress.exe command.',
      'cli.lbl.preset': 'Preset',
      'cli.lbl.codec': 'Video Codec',
      'cli.lbl.concurrency': 'Concurrency (-c)',
      'cli.lbl.hw': 'GPU hardware encoding (--hw)',
      'cli.lbl.audio': 'Audio',
      'cli.opt.preset_small': 'Small',
      'cli.opt.preset_standard': 'Standard',
      'cli.opt.preset_quality': 'High Quality',
      'cli.opt.codec_auto': 'Default (H.264)',
      'cli.opt.audio_compress': 'Compress (default)',
      'cli.opt.audio_copy': 'Pass-through (-a)',
      'cli.opt.audio_mute': 'Mute (--audio mute)',
      'cli.lbl.crop': 'Center Crop',
      'cli.opt.crop_off': 'Off',
      'cli.lbl.crop_scale': 'Keep Scale (--crop-scale)',
      'cli.lbl.conflict': 'Existing Output',
      'cli.opt.conflict_number': 'Auto-number (default)',
      'cli.opt.conflict_force': 'Overwrite (--force)',
      'cli.opt.conflict_skip': 'Skip existing (--skip-existing)',
      'cli.state.on': 'Enabled',
      'cli.state.off': 'Off',
      'cli.btn.copy': 'Copy',
      'cli.btn.reset': 'Reset',
      'cli.toast.copied': 'Copied to clipboard',
      'cli.toast.reset': 'Command options reset',
      'cli.examples_title': 'Common Commands',
      'cli.ex1_label': 'Compress with small preset, GPU encoding, and audio pass-through',
      'cli.ex2_label': 'Center-crop to 16:9 keeping 90% of the frame before compressing',
      'cli.ex3_label': 'Process 3 files concurrently and skip existing outputs',

      'sys.title': 'Windows Integration & FFmpeg Setup',
      'sys.subtitle': 'All integrations write to the current user profile only—no administrator privileges required. Multi-selected files forward into one running window.',
      'sys.int_heading': 'Current-User Integrations',
      'sys.int.ctx_title': 'Right-Click Context Menu',
      'sys.int.ctx_desc': 'Right-click any video in Explorer and choose “Compress with Videopress”; label updates automatically when language changes.',
      'sys.int.sendto_title': 'SendTo Menu',
      'sys.int.sendto_desc': 'Right-click → Send to for background compression with a Windows notification on completion.',
      'sys.int.shortcuts_title': 'Desktop & Start Menu Shortcuts',
      'sys.int.shortcuts_desc': 'Toggle from Settings whenever you want Start search or desktop drag-and-drop launchers.',
      'sys.int.path_title': 'User Path Variable',
      'sys.int.path_desc': 'Adds the program folder to the current user Path so videopress is available in any terminal.',
      'sys.col.ffmpeg_title': 'FFmpeg Lookup Order & Setup',
      'sys.col.ffmpeg_1': 'On first launch without FFmpeg, click “Download and configure” in the setup wizard to fetch ffmpeg-static v6.0 (~28 MB) from npmmirror and verify it.',
      'sys.col.ffmpeg_2': 'ffmpeg.exe placed beside Videopress.exe (recommended for offline portable use).',
      'sys.col.ffmpeg_3': 'ffmpeg.exe found on the system PATH.',
      'sys.col.ffmpeg_4': 'Manually pin an ffmpeg.exe path in Setup or Settings → About, with one-click reset to auto-detect.',
      'sys.col.cli_int_title': 'Toggle Integration via CLI',
      'sys.col.cli_int_desc': 'Besides Settings, SendTo and user Path can be managed from a terminal:',
      'sys.col.formats_title': 'Supported Input Formats & Summary Logs',
      'sys.col.formats_body': 'Supports MP4, MOV, MKV, AVI, M4V, WMV, WEBM, TS, FLV, MPG, MPEG, and 3GP. Each batch appends compress_summary.log in the output folder with source size, output size, savings ratio, and duration.',

      'dl.title': 'Download Videopress',
      'dl.subtitle': 'Available on GitHub Releases as an installer or standalone portable binary, both accompanied by SHA256SUMS.txt.',
      'dl.installer.title': 'Installer',
      'dl.installer.file': 'videopress-setup-<version>-amd64.exe',
      'dl.installer.desc': 'Adds Start menu entries and standard uninstallation for regular desktop use.',
      'dl.installer.steps': 'Run installer → launch from Start menu → auto-download or pick FFmpeg on first run.',
      'dl.portable.title': 'Portable',
      'dl.portable.file': 'Videopress.exe',
      'dl.portable.desc': 'Single executable with no installation or admin rights required; ready for tool folders or USB drives.',
      'dl.portable.steps': 'Place in a folder and run directly; put ffmpeg.exe alongside it for offline machines.',
      'dl.btn.releases': 'Open GitHub Releases',
      'dl.col.smartscreen_title': 'Windows SmartScreen & Checksum Verification',
      'dl.col.smartscreen_body': 'Release binaries are currently unsigned, so Windows SmartScreen may prompt on first run—select “More info” → “Run anyway”. Verify integrity against SHA256SUMS.txt using PowerShell Get-FileHash.',
      'dl.col.req_title': 'System Requirements & Local Config Paths',
      'dl.col.req_body': 'Requires Windows 10 / 11 (x64). Preferences are stored in %APPDATA%\\Videopress\\settings.json and the GPU probe cache in %LOCALAPPDATA%\\videopress_gpu.cache.',

      'footer.tagline': 'Local video compression & animated export for Windows',
      'footer.license': 'MIT License · Copyright © 2026 Shirolin',
      'footer.link.repo': 'GitHub',
      'footer.link.releases': 'Releases'
    }
  };

  var SCENES = {
    demo:      { kind: 'video', src: 'https://media.shirolin.site/projects/videopress/demo.mp4', poster: 'assets/workspace-dark.png' },
    workspace: { kind: 'image', src: 'assets/workspace-dark.png',   width: 900, height: 820,  altKey: 'showcase.alt.workspace' },
    batch:     { kind: 'image', src: 'assets/batch-progress.png',   width: 900, height: 820,  altKey: 'showcase.alt.batch' },
    gif:       { kind: 'image', src: 'assets/gif-custom.png',       width: 900, height: 1250, altKey: 'showcase.alt.gif' },
    theme:     { kind: 'image', src: 'assets/appearance-light.png', width: 900, height: 820,  altKey: 'showcase.alt.theme' }
  };

  var state = {
    themeMode: document.documentElement.dataset.themePref || 'system',
    resolvedTheme: document.documentElement.dataset.theme || 'light',
    lang: document.documentElement.dataset.lang || 'zh',
    scene: 'demo',
    cli: {
      preset: 'standard',
      codec: 'h264',
      hw: false,
      audio: 'compress',
      concurrency: 1,
      crop: 'off',
      cropScale: 1.0,
      conflict: 'number'
    }
  };

  function t(key) {
    var dict = COPY[state.lang] || COPY.zh;
    return dict[key] !== undefined ? dict[key] : (COPY.zh[key] || key);
  }

  /* ==========================================================================
     Theme Management (system / light / dark)
     ========================================================================== */
  var mediaQuery = window.matchMedia ? window.matchMedia('(prefers-color-scheme: dark)') : null;

  function resolveTheme(mode) {
    if (mode === 'light' || mode === 'dark') return mode;
    return mediaQuery && mediaQuery.matches ? 'dark' : 'light';
  }

  function syncHeroImage() {
    var img = document.getElementById('hero-screenshot');
    if (!img) return;
    var isDark = state.resolvedTheme === 'dark';
    var nextSrc = isDark ? 'assets/workspace-dark.png' : 'assets/appearance-light.png';
    if (img.getAttribute('src') !== nextSrc) {
      img.setAttribute('src', nextSrc);
    }
    img.setAttribute('alt', t(isDark ? 'hero.alt_dark' : 'hero.alt_light'));
  }

  function applyTheme(mode, persist) {
    if (mode !== 'light' && mode !== 'dark' && mode !== 'system') mode = 'system';
    state.themeMode = mode;
    state.resolvedTheme = resolveTheme(mode);

    document.documentElement.dataset.theme = state.resolvedTheme;
    document.documentElement.dataset.themePref = mode;

    var metaColor = document.getElementById('meta-theme-color');
    if (metaColor) {
      metaColor.setAttribute('content', state.resolvedTheme === 'dark' ? '#191618' : '#F2EBDD');
    }

    var buttons = document.querySelectorAll('button[data-theme-mode]');
    for (var i = 0; i < buttons.length; i++) {
      var btn = buttons[i];
      var active = btn.getAttribute('data-theme-mode') === mode;
      btn.setAttribute('aria-pressed', active ? 'true' : 'false');
    }

    syncHeroImage();

    if (persist) {
      try {
        localStorage.setItem(STORAGE_THEME, mode);
      } catch (_) {}
    }
  }

  /* ==========================================================================
     Bilingual Switching (zh / en)
     ========================================================================== */
  function applyLanguage(lang, persist) {
    if (lang !== 'zh' && lang !== 'en') lang = 'zh';
    state.lang = lang;

    document.documentElement.dataset.lang = lang;
    document.documentElement.lang = lang === 'en' ? 'en' : 'zh-CN';
    document.title = t('meta.title');

    var metaDesc = document.getElementById('meta-desc');
    if (metaDesc) metaDesc.setAttribute('content', t('meta.desc'));

    var nodes = document.querySelectorAll('[data-i18n]');
    for (var i = 0; i < nodes.length; i++) {
      var key = nodes[i].getAttribute('data-i18n');
      nodes[i].textContent = t(key);
    }

    var ariaNodes = document.querySelectorAll('[data-i18n-aria]');
    for (var j = 0; j < ariaNodes.length; j++) {
      var ariaKey = ariaNodes[j].getAttribute('data-i18n-aria');
      ariaNodes[j].setAttribute('aria-label', t(ariaKey));
    }

    var themeBtns = document.querySelectorAll('button[data-theme-mode]');
    for (var k = 0; k < themeBtns.length; k++) {
      var tm = themeBtns[k].getAttribute('data-theme-mode');
      themeBtns[k].setAttribute('title', t('theme.' + tm));
    }

    var langBtns = document.querySelectorAll('[data-lang-set]');
    for (var m = 0; m < langBtns.length; m++) {
      var active = langBtns[m].getAttribute('data-lang-set') === lang;
      langBtns[m].setAttribute('aria-pressed', active ? 'true' : 'false');
    }

    var hwState = document.getElementById('cli-hw-state');
    if (hwState) {
      hwState.textContent = t(state.cli.hw ? 'cli.state.on' : 'cli.state.off');
    }

    syncHeroImage();
    syncShowcaseImage();
    syncMenuLabel();

    if (persist) {
      try {
        localStorage.setItem(STORAGE_LANG, lang);
      } catch (_) {}
    }
  }

  /* ==========================================================================
     Mobile Drawer
     ========================================================================== */
  var menuToggle = document.getElementById('menu-toggle');
  var mobileDrawer = document.getElementById('mobile-drawer');

  function syncMenuLabel() {
    if (!menuToggle) return;
    var open = menuToggle.getAttribute('aria-expanded') === 'true';
    menuToggle.setAttribute('aria-label', t(open ? 'a11y.menu_close' : 'a11y.menu_open'));
  }

  function setDrawerOpen(open) {
    if (!menuToggle || !mobileDrawer) return;
    menuToggle.setAttribute('aria-expanded', open ? 'true' : 'false');
    mobileDrawer.hidden = !open;
    syncMenuLabel();
  }

  /* ==========================================================================
     Interface Showcase Switcher
     ========================================================================== */
  function syncShowcaseImage() {
    var img = document.getElementById('showcase-img');
    var video = document.getElementById('showcase-video');
    var panel = document.getElementById('showcase-panel');
    var spec = SCENES[state.scene] || SCENES.demo;
    var isVideo = spec.kind === 'video';

    if (video) {
      if (isVideo) {
        video.hidden = false;
      } else {
        if (!video.paused) video.pause();
        video.hidden = true;
      }
    }

    if (img) {
      if (isVideo) {
        img.hidden = true;
      } else {
        img.hidden = false;
        if (img.getAttribute('src') !== spec.src) {
          img.setAttribute('src', spec.src);
        }
        img.setAttribute('width', String(spec.width));
        img.setAttribute('height', String(spec.height));
        img.setAttribute('alt', t(spec.altKey));
      }
    }

    if (panel) {
      panel.setAttribute('aria-labelledby', 'tab-' + state.scene);
    }
  }

  function selectScene(scene) {
    if (!SCENES[scene]) return;
    state.scene = scene;

    var tabs = document.querySelectorAll('#showcase-tabs [data-scene]');
    for (var i = 0; i < tabs.length; i++) {
      var active = tabs[i].getAttribute('data-scene') === scene;
      tabs[i].setAttribute('aria-selected', active ? 'true' : 'false');
    }

    var sideItems = document.querySelectorAll('.showcase-sidebar [data-scene]');
    for (var j = 0; j < sideItems.length; j++) {
      var isSideActive = sideItems[j].getAttribute('data-scene') === scene;
      sideItems[j].classList.toggle('is-active', isSideActive);
      sideItems[j].setAttribute('aria-current', isSideActive ? 'true' : 'false');
    }

    syncShowcaseImage();
  }

  /* ==========================================================================
     Interactive CLI Command Builder
     ========================================================================== */
  function formatScale(val) {
    var n = Math.round(val * 100) / 100;
    return n.toFixed(2).replace(/0$/, '');
  }

  function buildCommandParts() {
    var c = state.cli;
    var flags = [];

    if (c.preset !== 'standard') {
      flags.push('--preset', c.preset);
    }
    if (c.codec !== 'h264') {
      flags.push('--codec', c.codec);
    }
    if (c.hw) {
      flags.push('--hw');
    }
    if (c.audio === 'copy') {
      flags.push('-a');
    } else if (c.audio === 'mute') {
      flags.push('--audio', 'mute');
    }
    if (c.concurrency > 1) {
      flags.push('-c', String(c.concurrency));
    }
    if (c.crop !== 'off') {
      flags.push('--crop-center', c.crop);
      if (c.cropScale < 0.999) {
        flags.push('--crop-scale', formatScale(c.cropScale));
      }
    }
    if (c.conflict === 'force') {
      flags.push('--force');
    } else if (c.conflict === 'skip') {
      flags.push('--skip-existing');
    }

    // Passing at least one flag ensures Videopress starts in CLI mode rather than GUI mode.
    if (flags.length === 0) {
      flags.push('--preset', 'standard');
    }

    var files = ['"C:\\videos\\clip.mp4"'];
    if (c.concurrency === 2) {
      files = ['"C:\\videos\\a.mp4"', '"C:\\videos\\b.mp4"'];
    } else if (c.concurrency >= 3) {
      files = ['"C:\\videos\\a.mp4"', '"C:\\videos\\b.mp4"', '"C:\\videos\\c.mp4"'];
    }

    return {
      bin: 'videopress.exe',
      flags: flags,
      files: files,
      fullText: ['videopress.exe'].concat(flags, files).join(' ')
    };
  }

  function renderCliOutput() {
    var out = document.getElementById('cli-command-output');
    if (!out) return;
    var parts = buildCommandParts();

    out.textContent = '';

    var binSpan = document.createElement('span');
    binSpan.className = 'tok-bin';
    binSpan.textContent = parts.bin;
    out.appendChild(binSpan);

    if (parts.flags.length > 0) {
      out.appendChild(document.createTextNode(' ' + parts.flags.join(' ')));
    }

    var pathSpan = document.createElement('span');
    pathSpan.className = 'tok-path';
    pathSpan.textContent = ' ' + parts.files.join(' ');
    out.appendChild(pathSpan);

    out.dataset.command = parts.fullText;
  }

  function syncSegmentedGroup(containerId, activeValue) {
    var group = document.getElementById(containerId);
    if (!group) return;
    var btns = group.querySelectorAll('[data-value]');
    for (var i = 0; i < btns.length; i++) {
      var active = btns[i].getAttribute('data-value') === activeValue;
      btns[i].setAttribute('aria-pressed', active ? 'true' : 'false');
    }
  }

  function syncCliControls() {
    var c = state.cli;
    syncSegmentedGroup('cli-preset', c.preset);
    syncSegmentedGroup('cli-codec', c.codec);
    syncSegmentedGroup('cli-audio', c.audio);
    syncSegmentedGroup('cli-conflict', c.conflict);
    syncSegmentedGroup('cli-crop', c.crop);

    var concInput = document.getElementById('cli-concurrency-val');
    var decBtn = document.getElementById('concurrency-dec');
    var incBtn = document.getElementById('concurrency-inc');
    if (concInput) concInput.value = String(c.concurrency);
    if (decBtn) decBtn.disabled = c.concurrency <= 1;
    if (incBtn) incBtn.disabled = c.concurrency >= 8;

    var hwBtn = document.getElementById('cli-hw');
    var hwState = document.getElementById('cli-hw-state');
    if (hwBtn) hwBtn.setAttribute('aria-checked', c.hw ? 'true' : 'false');
    if (hwState) hwState.textContent = t(c.hw ? 'cli.state.on' : 'cli.state.off');

    var scaleInput = document.getElementById('cli-crop-scale');
    var scaleReadout = document.getElementById('cli-crop-scale-readout');
    if (scaleInput) {
      scaleInput.disabled = c.crop === 'off';
      scaleInput.value = String(c.cropScale);
    }
    if (scaleReadout) {
      scaleReadout.textContent = c.cropScale.toFixed(2);
    }

    renderCliOutput();
  }

  /* ==========================================================================
     Clipboard & MujicaUI Toast
     ========================================================================== */
  var toastTimer = null;

  function showToast(message) {
    var toaster = document.getElementById('toaster');
    if (!toaster) return;

    toaster.textContent = '';
    var toast = document.createElement('div');
    toast.className = 'toast';

    var svgNS = 'http://www.w3.org/2000/svg';
    var svg = document.createElementNS(svgNS, 'svg');
    svg.setAttribute('class', 'icon');
    svg.setAttribute('viewBox', '0 0 24 24');
    svg.setAttribute('aria-hidden', 'true');
    var poly = document.createElementNS(svgNS, 'polyline');
    poly.setAttribute('points', '20 6 9 17 4 12');
    svg.appendChild(poly);

    var label = document.createElement('span');
    label.textContent = message;

    toast.appendChild(svg);
    toast.appendChild(label);
    toaster.appendChild(toast);

    if (toastTimer) clearTimeout(toastTimer);
    toastTimer = setTimeout(function () {
      toaster.textContent = '';
    }, 2400);
  }

  function copyToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(function () {
        showToast(t('cli.toast.copied'));
      }, function () {
        fallbackCopy(text);
      });
    } else {
      fallbackCopy(text);
    }
  }

  function fallbackCopy(text) {
    var area = document.createElement('textarea');
    area.value = text;
    area.setAttribute('readonly', '');
    area.style.position = 'fixed';
    area.style.left = '-9999px';
    document.body.appendChild(area);
    area.select();
    try {
      document.execCommand('copy');
    } catch (_) {}
    document.body.removeChild(area);
    showToast(t('cli.toast.copied'));
  }

  /* ==========================================================================
     Event Binding & Initialization
     ========================================================================== */
  function bindSegmented(containerId, onChange) {
    var group = document.getElementById(containerId);
    if (!group) return;
    group.addEventListener('click', function (ev) {
      var btn = ev.target.closest('[data-value]');
      if (!btn) return;
      onChange(btn.getAttribute('data-value'));
    });
  }

  function init() {
    // Theme controls (desktop + mobile drawer)
    document.addEventListener('click', function (ev) {
      var themeBtn = ev.target.closest('button[data-theme-mode]');
      if (themeBtn) {
        applyTheme(themeBtn.getAttribute('data-theme-mode'), true);
        return;
      }

      var langBtn = ev.target.closest('[data-lang-set]');
      if (langBtn) {
        applyLanguage(langBtn.getAttribute('data-lang-set'), true);
        return;
      }

      var sceneBtn = ev.target.closest('[data-scene]');
      if (sceneBtn) {
        selectScene(sceneBtn.getAttribute('data-scene'));
        return;
      }

      var copyBtn = ev.target.closest('[data-copy-text]');
      if (copyBtn) {
        copyToClipboard(copyBtn.getAttribute('data-copy-text'));
        return;
      }

      var colTrigger = ev.target.closest('.collapsible-trigger');
      if (colTrigger) {
        var expanded = colTrigger.getAttribute('aria-expanded') === 'true';
        var targetId = colTrigger.getAttribute('aria-controls');
        var body = targetId ? document.getElementById(targetId) : null;
        colTrigger.setAttribute('aria-expanded', expanded ? 'false' : 'true');
        if (body) body.hidden = expanded;
      }
    });

    // Listen to OS theme changes when in system mode
    if (mediaQuery && mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', function () {
        if (state.themeMode === 'system') {
          applyTheme('system', false);
        }
      });
    }

    // Mobile drawer toggle
    if (menuToggle && mobileDrawer) {
      menuToggle.addEventListener('click', function () {
        var open = menuToggle.getAttribute('aria-expanded') === 'true';
        setDrawerOpen(!open);
      });

      var drawerLinks = mobileDrawer.querySelectorAll('.mobile-nav-link');
      for (var i = 0; i < drawerLinks.length; i++) {
        drawerLinks[i].addEventListener('click', function () {
          setDrawerOpen(false);
        });
      }
    }

    // CLI builder bindings
    bindSegmented('cli-preset', function (val) {
      state.cli.preset = val;
      syncCliControls();
    });
    bindSegmented('cli-codec', function (val) {
      state.cli.codec = val;
      syncCliControls();
    });
    bindSegmented('cli-audio', function (val) {
      state.cli.audio = val;
      syncCliControls();
    });
    bindSegmented('cli-conflict', function (val) {
      state.cli.conflict = val;
      syncCliControls();
    });
    bindSegmented('cli-crop', function (val) {
      state.cli.crop = val;
      syncCliControls();
    });

    var concInput = document.getElementById('cli-concurrency-val');
    var decBtn = document.getElementById('concurrency-dec');
    var incBtn = document.getElementById('concurrency-inc');

    function setConcurrency(next) {
      var n = parseInt(next, 10);
      if (isNaN(n)) n = 1;
      state.cli.concurrency = Math.max(1, Math.min(8, n));
      syncCliControls();
    }

    if (concInput) {
      concInput.addEventListener('change', function () {
        setConcurrency(concInput.value);
      });
    }
    if (decBtn) {
      decBtn.addEventListener('click', function () {
        setConcurrency(state.cli.concurrency - 1);
      });
    }
    if (incBtn) {
      incBtn.addEventListener('click', function () {
        setConcurrency(state.cli.concurrency + 1);
      });
    }

    var hwBtn = document.getElementById('cli-hw');
    if (hwBtn) {
      hwBtn.addEventListener('click', function () {
        state.cli.hw = !state.cli.hw;
        syncCliControls();
      });
    }

    var scaleInput = document.getElementById('cli-crop-scale');
    if (scaleInput) {
      scaleInput.addEventListener('input', function () {
        state.cli.cropScale = parseFloat(scaleInput.value) || 1.0;
        syncCliControls();
      });
    }

    var cliCopyBtn = document.getElementById('cli-copy-btn');
    if (cliCopyBtn) {
      cliCopyBtn.addEventListener('click', function () {
        var out = document.getElementById('cli-command-output');
        var cmd = out ? (out.dataset.command || out.textContent) : '';
        if (cmd) copyToClipboard(cmd);
      });
    }

    var cliResetBtn = document.getElementById('cli-reset-btn');
    if (cliResetBtn) {
      cliResetBtn.addEventListener('click', function () {
        state.cli = {
          preset: 'standard',
          codec: 'h264',
          hw: false,
          audio: 'compress',
          concurrency: 1,
          crop: 'off',
          cropScale: 1.0,
          conflict: 'number'
        };
        syncCliControls();
        showToast(t('cli.toast.reset'));
      });
    }

    var heroDemoBtn = document.getElementById('hero-demo-btn');
    if (heroDemoBtn) {
      heroDemoBtn.addEventListener('click', function () {
        selectScene('demo');
        var video = document.getElementById('showcase-video');
        if (video && video.play) {
          var p = video.play();
          if (p && typeof p.catch === 'function') p.catch(function () {});
        }
      });
    }

    // Initial render
    applyTheme(state.themeMode, false);
    applyLanguage(state.lang, false);
    selectScene(state.scene);
    syncCliControls();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
  } else {
    init();
  }
})();
