<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import DropZone from './components/DropZone.svelte';
  import FileQueue, { type QueueItem } from './components/FileQueue.svelte';
  import Settings from './components/Settings.svelte';
  
  import { StartCompress, OpenFolder, DetectFFmpeg, SelectFolder, DownloadFFmpeg, GetInitialFiles, GetVersion, CancelCompress } from '../wailsjs/go/main/App.js';
  import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime.js';

  let queueItems: QueueItem[] = [];
  
  // Compression Settings initialized from localStorage
  let preset: string = localStorage.getItem('videopress_preset') || 'standard';
  let concurrency: number = parseInt(localStorage.getItem('videopress_concurrency') || '1', 10);
  let hwAccel: boolean = localStorage.getItem('videopress_hw_accel') === 'true';
  let copyAudio: boolean = localStorage.getItem('videopress_copy_audio') === 'true';
  let forceMode: boolean = localStorage.getItem('videopress_force_mode') === 'true';
  let skipExisting: boolean = localStorage.getItem('videopress_skip_existing') === 'true';

  // Persist settings reactively
  $: if (preset !== undefined) localStorage.setItem('videopress_preset', preset);
  $: if (concurrency !== undefined) localStorage.setItem('videopress_concurrency', concurrency.toString());
  $: if (hwAccel !== undefined) localStorage.setItem('videopress_hw_accel', hwAccel.toString());
  $: if (copyAudio !== undefined) localStorage.setItem('videopress_copy_audio', copyAudio.toString());
  $: if (forceMode !== undefined) localStorage.setItem('videopress_force_mode', forceMode.toString());
  $: if (skipExisting !== undefined) localStorage.setItem('videopress_skip_existing', skipExisting.toString());

  let isCompressing = false;
  let ffmpegError = '';
  let showSettings = false;
  let appVersion = 'v0.1.0';

  // FFmpeg auto-download state
  let isDownloadingFFmpeg = false;
  let downloadPercent = 0;
  let downloadError = '';

  // Custom output directory with localStorage persistence - optimized to avoid reactive trigger on initial load
  let customOutputDir = localStorage.getItem('videopress_custom_output_dir') || '';
  let lastOutputDir = '';

  function saveCustomOutputDir(dir: string) {
    customOutputDir = dir;
    localStorage.setItem('videopress_custom_output_dir', dir);
  }

  // Real-time stats calculation
  $: totalCount = queueItems.length;
  $: successItems = queueItems.filter(item => item.status === 'success');
  $: successCount = successItems.length;
  $: compressingCount = queueItems.filter(item => item.status === 'compressing').length;
  $: waitingCount = queueItems.filter(item => item.status === 'waiting').length;

  $: totalSavedBytes = successItems.reduce((sum, item) => {
    if (item.size && item.targetSize && item.size > item.targetSize) {
      return sum + (item.size - item.targetSize);
    }
    return sum;
  }, 0);

  $: totalOriginalBytes = successItems.reduce((sum, item) => sum + (item.size || 0), 0);
  
  $: averageRatio = totalOriginalBytes > 0 
    ? ((totalSavedBytes / totalOriginalBytes) * 100).toFixed(1) + '%' 
    : '0.0%';

  function formatSavedSize(bytes: number): string {
    if (bytes <= 0) return '0 MB';
    const mb = bytes / (1024 * 1024);
    if (mb < 1024) return mb.toFixed(1) + ' MB';
    return (mb / 1024).toFixed(2) + ' GB';
  }

  onMount(async () => {
    // Check if FFmpeg is installed
    try {
      await DetectFFmpeg();
    } catch (e: any) {
      ffmpegError = e.message || '未找到 FFmpeg，请先安装并将 ffmpeg.exe 添加到系统环境变量中。';
    }

    // Load application version
    try {
      const v = await GetVersion();
      if (v) appVersion = 'v' + v;
    } catch (e) {
      console.error("加载版本号失败:", e);
    }

    // Set up Wails events listener for video compression progress
    EventsOn('progress', handleProgressEvent);

    // 主动拉取启动时传入的视频文件并装载
    try {
      const files = await GetInitialFiles();
      if (files && files.length > 0) {
        addFilesToQueue(files);
      }
    } catch (e) {
      console.error("加载初始视频文件失败:", e);
    }
  });

  onDestroy(() => {
    EventsOff('progress');
  });

  function handleProgressEvent(ev: any) {
    const fileBase = ev.File;
    queueItems = queueItems.map(item => {
      if (item.name === fileBase) {
        let status = item.status;
        if (ev.Error) {
          status = 'failed';
        } else if (ev.Done) {
          status = 'success';
        } else {
          status = 'compressing';
        }
        return {
          ...item,
          percent: ev.Percent,
          status: status,
          error: ev.Error || undefined
        };
      }
      return item;
    });
  }

  async function handleDownloadFFmpeg() {
    isDownloadingFFmpeg = true;
    downloadPercent = 0;
    downloadError = '';

    EventsOn('download-progress', (percent: number) => {
      if (percent < 0) {
        downloadError = '下载或解压缩视频引擎 (FFmpeg) 失败，请检查网络并重试。';
        isDownloadingFFmpeg = false;
        EventsOff('download-progress');
      } else {
        downloadPercent = percent;
        if (percent >= 100) {
          isDownloadingFFmpeg = false;
          EventsOff('download-progress');
          setTimeout(async () => {
            try {
              await DetectFFmpeg();
              ffmpegError = '';
            } catch (e: any) {
              ffmpegError = e.message || '引擎加载失败，请尝试重启软件。';
            }
          }, 800);
        }
      }
    });

    try {
      await DownloadFFmpeg();
    } catch (err: any) {
      downloadError = err.message || '启动视频引擎配置任务失败。';
      isDownloadingFFmpeg = false;
      EventsOff('download-progress');
    }
  }

  function addFilesToQueue(paths: string[]) {
    const newItems = paths.map(path => {
      const name = path.split(/[/\\]/).pop() || path;
      const isAlreadyCompressed = name.includes('.compressed');
      return {
        path,
        name,
        size: 0,
        percent: 0,
        status: 'waiting' as const,
        error: isAlreadyCompressed ? '警告：该视频可能已经被压缩过，二次压缩会损害画质。' : undefined,
        isWarning: isAlreadyCompressed,
      };
    });

    // Add only new unique paths
    const existingPaths = new Set(queueItems.map(item => item.path));
    const uniqueNewItems = newItems.filter(item => !existingPaths.has(item.path));

    if (uniqueNewItems.length > 0) {
      queueItems = [...queueItems, ...uniqueNewItems];
    }
  }

  function handleFileSelect(e: CustomEvent<string[]>) {
    addFilesToQueue(e.detail);
  }

  function handleRemove(e: CustomEvent<number>) {
    const index = e.detail;
    queueItems = queueItems.filter((_, i) => i !== index);
  }

  function handleClear() {
    if (isCompressing) return;
    queueItems = [];
  }

  async function triggerCompression() {
    if (isCompressing || queueItems.length === 0) return;
    
    isCompressing = true;
    
    // Reset status of non-completed items to waiting
    queueItems = queueItems.map(item => ({
      ...item,
      percent: 0,
      status: 'waiting' as const,
      error: item.isWarning ? '警告：该视频可能已经被压缩过，二次压缩会损害画质。' : undefined,
      targetSize: undefined,
      duration: undefined
    }));

    const files = queueItems.map(item => item.path);

    try {
      const reports = await StartCompress({
        Files: files,
        Preset: preset,
        HWAccel: hwAccel,
        CopyAudio: copyAudio,
        ForceMode: forceMode,
        SkipExisting: skipExisting,
        Concurrency: concurrency,
        OutputDir: customOutputDir
      });

      // Update queue items with target results
      queueItems = queueItems.map(item => {
        const report = reports.find((r: any) => r.InputName === item.name);
        if (report) {
          let status = 'success' as const;
          if (report.Status === '跳过') {
            status = 'skipped' as const;
          } else if (report.Status === '失败') {
            status = 'failed' as const;
          }
          if (report.Status === '成功' && report.OutputDir) {
            lastOutputDir = report.OutputDir;
          }
          return {
            ...item,
            size: report.SourceSize,
            targetSize: report.TargetSize,
            status: status,
            duration: report.Duration / 1000000, // convert ns to ms
            error: report.ErrMessage || undefined
          };
        }
        return item;
      });
    } catch (err: any) {
      console.error("Compression execution error:", err);
    } finally {
      isCompressing = false;
    }
  }

  async function handleCancelCompression() {
    try {
      await CancelCompress();
    } catch (err) {
      console.error("Failed to cancel compression:", err);
    }
  }

  async function handleOpenOutputFolder() {
    const dirToOpen = customOutputDir || lastOutputDir;
    if (dirToOpen) {
      try {
        await OpenFolder(dirToOpen);
      } catch (err) {
        console.error("Failed to open output folder:", err);
      }
      return;
    }

    const successItem = queueItems.find(item => item.status === 'success');
    if (successItem) {
      const lastSlashIndex = Math.max(successItem.path.lastIndexOf('\\'), successItem.path.lastIndexOf('/'));
      const sourceDir = lastSlashIndex !== -1 ? successItem.path.substring(0, lastSlashIndex) : '';
      const separator = successItem.path.includes('/') ? '/' : '\\';
      const outputDir = sourceDir ? `${sourceDir}${separator}compressed` : 'compressed';
      try {
        await OpenFolder(outputDir);
      } catch (err) {
        console.error("Failed to open folder:", err);
      }
    }
  }

  async function handleChangeOutputDir() {
    try {
      const dir = await SelectFolder();
      if (dir) {
        saveCustomOutputDir(dir);
      }
    } catch (err) {
      console.error("Failed to select output folder:", err);
    }
  }
</script>

<div class="app-layout" on:dragover|preventDefault on:drop|preventDefault>
  <!-- Top Navigation Bar -->
  <header class="app-header glass-panel">
    <div class="brand">
      <div class="logo-box">
        <svg class="logo-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <polygon points="23 7 16 12 23 17 23 7"></polygon>
          <rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
        </svg>
      </div>
      <div class="brand-text">
        <h1>Videopress</h1>
        <span class="badge">{appVersion}</span>
      </div>
    </div>
    
    <div class="nav-controls">
      <button 
        class="nav-btn {showSettings ? 'active' : ''}" 
        on:click={() => showSettings = !showSettings}
        disabled={isCompressing}
      >
        {#if showSettings}
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="nav-icon">
            <rect x="3" y="3" width="7" height="9"></rect><rect x="14" y="3" width="7" height="5"></rect><rect x="14" y="12" width="7" height="9"></rect><rect x="3" y="16" width="7" height="5"></rect>
          </svg>
          主工作区
        {:else}
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="nav-icon">
            <circle cx="12" cy="12" r="3"></circle>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
          </svg>
          配置面板
        {/if}
      </button>
    </div>
  </header>

  <!-- Main Responsive Workspace -->
  <div class="workspace">
    {#if ffmpegError && !showSettings}
      <!-- FFmpeg 缺失自配置面板 -->
      <div class="setup-container glass-panel">
        <div class="setup-header">
          <svg class="setup-logo-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line>
          </svg>
          <h2>需要配置核心视频引擎</h2>
        </div>
        <p class="setup-desc">
          检测到系统未安装 <strong>FFmpeg</strong> 视频引擎依赖。<br>
          Videopress 需要该依赖来进行高效率的视频解码与压缩任务。
        </p>

        {#if isDownloadingFFmpeg}
          <div class="download-progress-container">
            <span class="download-status">正在极速下载核心组件... ({downloadPercent.toFixed(1)}%)</span>
            <div class="setup-progress-bar">
              <div class="setup-progress-fill" style="width: {downloadPercent}%"></div>
            </div>
            <p class="download-tip">请不要关闭软件，这可能需要几十秒时间。配置成功后会自动进入主界面。</p>
          </div>
        {:else}
          <div class="setup-actions">
            <button class="setup-download-btn" on:click={handleDownloadFFmpeg}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="btn-icon-down">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line>
              </svg>
              一键闪电自动配置 (推荐)
            </button>
            <button class="setup-config-btn" on:click={() => showSettings = true}>
              手动指定环境变量或寻找本地文件
            </button>
          </div>
        {/if}

        {#if downloadError}
          <div class="setup-error-msg">
            ⚠️ {downloadError}
          </div>
        {/if}
      </div>
    {:else if showSettings}
      <Settings 
        bind:preset 
        bind:concurrency 
        bind:hwAccel 
        bind:copyAudio 
        bind:forceMode 
        bind:skipExisting
      />
    {:else}
      <div class="dashboard">
        <!-- Stats Dashboard Banner (Only visible when items exist) -->
        {#if totalCount > 0}
          <div class="stats-banner glass-panel">
            <div class="stat-card">
              <span class="val">{totalCount}</span>
              <span class="label">任务总数</span>
            </div>
            <div class="stat-card">
              <span class="val text-purple">
                {#if compressingCount > 0}
                  {compressingCount} <span class="mini-status">处理中</span>
                {:else if waitingCount > 0}
                  {waitingCount} <span class="mini-status">待命</span>
                {:else}
                  {successCount} <span class="mini-status">完成</span>
                {/if}
              </span>
              <span class="label">当前状态</span>
            </div>
            <div class="stat-card">
              <span class="val text-green">{formatSavedSize(totalSavedBytes)}</span>
              <span class="label">共节省空间</span>
            </div>
            <div class="stat-card">
              <span class="val text-magenta">{averageRatio}</span>
              <span class="label">平均压缩率</span>
            </div>
          </div>
        {/if}

        <!-- Quick Controls Panel (Preset selector and custom output path) -->
        <div class="quick-controls-panel glass-panel">
          <!-- Preset tabs -->
          <div class="control-group">
            <span class="control-label">压缩预设</span>
            <div class="segmented-control">
              <button 
                class="segment-btn {preset === 'small' ? 'active' : ''}" 
                on:click={() => preset = 'small'}
                disabled={isCompressing}
              >
                小文件
              </button>
              <button 
                class="segment-btn {preset === 'standard' ? 'active' : ''}" 
                on:click={() => preset = 'standard'}
                disabled={isCompressing}
              >
                标准
              </button>
              <button 
                class="segment-btn {preset === 'quality' ? 'active' : ''}" 
                on:click={() => preset = 'quality'}
                disabled={isCompressing}
              >
                高画质
              </button>
            </div>
          </div>

          <!-- Custom output folder -->
          <div class="control-group flex-1">
            <span class="control-label">保存目录</span>
            <div class="path-bar-container">
              <div class="path-display" title={customOutputDir || '原视频同目录下的 compressed/ 文件夹'}>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="path-icon">
                  <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
                </svg>
                <span class="path-text">
                  {#if customOutputDir}
                    {customOutputDir}
                  {:else}
                    默认 (保存至原视频目录下 compressed/ 文件夹)
                  {/if}
                </span>
              </div>
              <div class="path-actions">
                <button 
                  class="path-btn change" 
                  on:click={handleChangeOutputDir}
                  disabled={isCompressing}
                >
                  更改
                </button>
                {#if customOutputDir}
                  <button 
                    class="path-btn reset" 
                    on:click={() => customOutputDir = ''} 
                    disabled={isCompressing}
                    title="恢复默认路径"
                  >
                    重置
                  </button>
                {/if}
              </div>
            </div>
          </div>
        </div>

        <!-- Drop Zone (Collapses to small bar if queue not empty) -->
        <DropZone compact={totalCount > 0} disabled={isCompressing} on:select={handleFileSelect} />
        
        <!-- Scrollable File Queue (Automatically expands to take all space) -->
        <FileQueue 
          items={queueItems} 
          isCompressing={isCompressing}
          on:remove={handleRemove} 
          on:clear={handleClear} 
        />

        <!-- Action Panel (Fixed size, locked at the bottom) -->
        <div class="action-panel">
          {#if queueItems.some(item => item.status === 'success' || item.status === 'failed' || item.status === 'skipped')}
            <button class="btn-folder" on:click={handleOpenOutputFolder}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="btn-icon"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
              打开输出文件夹
            </button>
          {/if}

          {#if isCompressing}
            <button class="compress-trigger-btn cancel-btn" on:click={handleCancelCompression}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="btn-icon"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect></svg>
              取消压缩任务
            </button>
          {:else}
            <button 
              class="compress-trigger-btn" 
              disabled={queueItems.length === 0 || !!ffmpegError}
              on:click={triggerCompression}
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="btn-icon"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>
              开始视频压缩
            </button>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .app-layout {
    display: flex;
    flex-direction: column;
    height: 100vh;
    padding: 0.8rem 1rem;
    gap: 0.8rem;
    max-width: 960px;
    margin: 0 auto;
    overflow: hidden; /* Lock the global viewport */
  }

  .app-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.6rem 1.1rem;
    height: 52px;
    flex-shrink: 0;      /* Header never shrinks */
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .logo-box {
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, rgba(168, 85, 247, 0.2) 0%, rgba(217, 70, 239, 0.2) 100%);
    border: 1px solid rgba(168, 85, 247, 0.3);
    padding: 0.4rem;
    border-radius: 8px;
    box-shadow: 0 0 12px rgba(168, 85, 247, 0.1);
  }

  .logo-icon {
    width: 16px;
    height: 16px;
    color: var(--accent-purple);
  }

  .brand-text {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  h1 {
    font-size: 1.05rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    background: linear-gradient(135deg, #ffffff 30%, #94a3b8 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .badge {
    font-size: 0.62rem;
    font-weight: 750;
    background: rgba(168, 85, 247, 0.12);
    color: var(--accent-purple);
    padding: 0.1rem 0.35rem;
    border-radius: 4px;
    border: 1px solid rgba(168, 85, 247, 0.15);
  }

  .nav-btn {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    font-size: 0.76rem;
    font-weight: 650;
    padding: 0.35rem 0.75rem;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .nav-btn:hover {
    border-color: rgba(255, 255, 255, 0.12);
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.04);
  }

  .nav-btn.active {
    background: var(--accent-purple);
    color: white;
    border-color: var(--accent-purple);
    box-shadow: 0 0 12px rgba(168, 85, 247, 0.35);
  }

  .nav-icon {
    width: 12px;
    height: 12px;
  }

  /* Error Banner */
  .error-banner {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    background: rgba(244, 63, 94, 0.08);
    border: 1px solid rgba(244, 63, 94, 0.18);
    padding: 0.6rem 0.8rem;
    border-radius: 8px;
    color: var(--accent-red);
    font-size: 0.75rem;
    font-weight: 600;
    flex-shrink: 0;
  }

  .error-icon {
    width: 16px;
    height: 16px;
    flex-shrink: 0;
  }

  /* Workspace Container */
  .workspace {
    flex: 1;             /* Grow to fill the layout */
    min-height: 0;       /* Allow container to scale down */
    overflow: hidden;    /* Keep inner scrolls independent */
    display: flex;
    flex-direction: column;
  }

  .dashboard {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    gap: 0.8rem;
  }

  /* Statistics Dashboard Banner */
  .stats-banner {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    padding: 0.75rem 0.5rem;
    background: rgba(255, 255, 255, 0.015);
    border-radius: 12px;
    flex-shrink: 0;
  }

  .stat-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    border-right: 1px solid var(--border-color);
  }

  .stat-card:last-child {
    border-right: none;
  }

  .stat-card .val {
    font-size: 0.95rem;
    font-weight: 800;
    color: var(--text-primary);
    display: flex;
    align-items: baseline;
    gap: 0.15rem;
  }

  .mini-status {
    font-size: 0.62rem;
    font-weight: 600;
    color: var(--text-muted);
  }

  .stat-card .label {
    font-size: 0.65rem;
    color: var(--text-muted);
    font-weight: 550;
    margin-top: 0.15rem;
  }

  .text-purple { color: var(--accent-purple) !important; }
  .text-green { color: var(--accent-green) !important; }
  .text-magenta { color: var(--accent-magenta) !important; }

  /* Quick Controls Panel Styling */
  .quick-controls-panel {
    display: flex;
    gap: 1.2rem;
    padding: 0.8rem 1.1rem;
    align-items: flex-end;
    background: rgba(255, 255, 255, 0.012) !important;
    border-radius: 12px;
    flex-shrink: 0;
  }

  @media (max-width: 650px) {
    .quick-controls-panel {
      flex-direction: column;
      align-items: stretch;
      gap: 0.8rem;
    }
  }

  .control-group {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .flex-1 {
    flex: 1;
  }

  .control-label {
    font-size: 0.72rem;
    font-weight: 700;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }

  /* Segmented Preset Switcher */
  .segmented-control {
    display: flex;
    background: rgba(0, 0, 0, 0.2);
    border: 1px solid var(--border-color);
    padding: 2px;
    border-radius: 8px;
    height: 32px;
  }

  .segment-btn {
    border: none;
    background: transparent;
    color: var(--text-secondary);
    font-size: 0.76rem;
    font-weight: 600;
    padding: 0 0.85rem;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    white-space: nowrap;
  }

  .segment-btn:hover:not(:disabled) {
    color: var(--text-primary);
  }

  .segment-btn.active {
    background: var(--accent-purple-gradient);
    color: white;
    box-shadow: var(--shadow-sm);
  }

  .segment-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Output Path Bar Container */
  .path-bar-container {
    display: flex;
    background: rgba(0, 0, 0, 0.2);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    height: 32px;
    padding: 2px;
    align-items: center;
    justify-content: space-between;
    overflow: hidden;
  }

  .path-display {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    padding-left: 0.6rem;
    overflow: hidden;
    flex: 1;
  }

  .path-icon {
    width: 13px;
    height: 13px;
    color: var(--accent-purple);
    flex-shrink: 0;
  }

  .path-text {
    font-size: 0.75rem;
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: 550;
  }

  .path-actions {
    display: flex;
    gap: 2px;
    height: 100%;
    align-items: center;
    padding-right: 2px;
  }

  .path-btn {
    border: none;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    font-size: 0.72rem;
    font-weight: 700;
    height: 24px;
    padding: 0 0.6rem;
    border-radius: 5px;
    cursor: pointer;
    transition: all 0.15s;
    white-space: nowrap;
  }

  .path-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.12);
  }

  .path-btn.reset {
    color: var(--accent-red);
    background: rgba(244, 63, 94, 0.04);
    border-color: rgba(244, 63, 94, 0.1);
  }

  .path-btn.reset:hover:not(:disabled) {
    background: rgba(244, 63, 94, 0.12);
    border-color: rgba(244, 63, 94, 0.2);
  }

  .path-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Action Panel Area */
  .action-panel {
    display: flex;
    gap: 0.8rem;
    flex-shrink: 0;      /* Lock action buttons height */
  }

  .compress-trigger-btn {
    flex: 1;
    height: 44px;
    border: none;
    border-radius: 10px;
    background: var(--accent-purple-gradient);
    color: white;
    font-size: 0.85rem;
    font-weight: 750;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    box-shadow: 0 4px 14px rgba(168, 85, 247, 0.25);
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .compress-trigger-btn.cancel-btn {
    background: linear-gradient(135deg, #f43f5e 0%, #e11d48 100%);
    box-shadow: 0 4px 14px rgba(244, 63, 94, 0.25);
  }

  .compress-trigger-btn.cancel-btn:hover:not(:disabled) {
    box-shadow: 0 6px 18px rgba(244, 63, 94, 0.4);
  }

  .compress-trigger-btn:hover:not(:disabled) {
    transform: translateY(-1.5px);
    box-shadow: 0 6px 18px rgba(168, 85, 247, 0.4);
  }

  .compress-trigger-btn:active:not(:disabled) {
    transform: translateY(0);
  }

  .compress-trigger-btn:disabled {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--border-color);
    color: var(--text-muted);
    cursor: not-allowed;
    box-shadow: none;
  }

  .btn-folder {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    padding: 0 1rem;
    border-radius: 10px;
    font-size: 0.8rem;
    font-weight: 650;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.4rem;
    transition: all 0.2s;
  }

  .btn-folder:hover {
    background: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.12);
  }

  .btn-icon {
    width: 13px;
    height: 13px;
  }

  .spinner {
    width: 14px;
    height: 14px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top: 2px solid white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  /* FFmpeg Autoconfig Setup panel styles */
  .setup-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem 2rem;
    max-width: 580px;
    margin: 4rem auto;
    text-align: center;
    border-radius: 16px;
    background: rgba(18, 18, 24, 0.4);
    box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.3);
  }

  .setup-header {
    display: flex;
    align-items: center;
    gap: 0.8rem;
    margin-bottom: 1.2rem;
    color: var(--accent-magenta);
  }

  .setup-logo-icon {
    width: 32px;
    height: 32px;
    color: var(--accent-magenta);
  }

  .setup-header h2 {
    font-size: 1.3rem;
    font-weight: 700;
    letter-spacing: -0.02em;
    color: var(--text-primary);
  }

  .setup-desc {
    font-size: 0.85rem;
    color: var(--text-secondary);
    line-height: 1.6;
    margin-bottom: 2.2rem;
  }

  .setup-desc strong {
    color: var(--accent-purple);
  }

  .setup-actions {
    display: flex;
    flex-direction: column;
    gap: 0.8rem;
    width: 100%;
    max-width: 320px;
  }

  .setup-download-btn {
    height: 46px;
    background: var(--accent-purple-gradient);
    border: none;
    border-radius: 10px;
    color: white;
    font-size: 0.85rem;
    font-weight: 750;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    box-shadow: 0 4px 15px rgba(168, 85, 247, 0.3);
    transition: all 0.2s;
  }

  .setup-download-btn:hover {
    transform: translateY(-1.5px);
    box-shadow: 0 6px 20px rgba(168, 85, 247, 0.45);
  }

  .btn-icon-down {
    width: 15px;
    height: 15px;
  }

  .setup-config-btn {
    height: 40px;
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid var(--border-color);
    border-radius: 10px;
    color: var(--text-secondary);
    font-size: 0.78rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .setup-config-btn:hover {
    background: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.12);
  }

  .download-progress-container {
    width: 100%;
    max-width: 380px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.6rem;
  }

  .download-status {
    font-size: 0.8rem;
    font-weight: 650;
    color: var(--accent-purple);
  }

  .setup-progress-bar {
    width: 100%;
    height: 8px;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 10px;
    overflow: hidden;
  }

  .setup-progress-fill {
    height: 100%;
    background: var(--accent-purple-gradient);
    border-radius: 10px;
    transition: width 0.1s linear;
  }

  .download-tip {
    font-size: 0.7rem;
    color: var(--text-muted);
    margin-top: 0.3rem;
  }

  .setup-error-msg {
    margin-top: 1.5rem;
    padding: 0.6rem 1rem;
    background: rgba(244, 63, 94, 0.08);
    border: 1px solid rgba(244, 63, 94, 0.2);
    color: var(--accent-red);
    font-size: 0.75rem;
    border-radius: 8px;
    font-weight: 600;
  }
</style>
