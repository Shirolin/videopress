<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import DropZone from './components/DropZone.svelte';
  import FileQueue, { type QueueItem } from './components/FileQueue.svelte';
  import Settings from './components/Settings.svelte';
  
  import { StartCompress, OpenFolder, DetectFFmpeg } from '../wailsjs/go/main/App.js';
  import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime.js';

  let queueItems: QueueItem[] = [];
  
  // Compression Settings
  let preset: string = 'standard';
  let concurrency: number = 1;
  let hwAccel: boolean = false;
  let copyAudio: boolean = false;
  let forceMode: boolean = false;
  let skipExisting: boolean = false;

  let isCompressing = false;
  let ffmpegError = '';
  let showSettings = false;

  onMount(async () => {
    // Check if FFmpeg is installed
    try {
      await DetectFFmpeg();
    } catch (e: any) {
      ffmpegError = e.message || '未找到 FFmpeg，请先安装并将 ffmpeg.exe 添加到系统环境变量中。';
    }

    // Set up Wails events listener for video compression progress
    EventsOn('progress', handleProgressEvent);
  });

  onDestroy(() => {
    EventsOff('progress');
  });

  function handleProgressEvent(ev: any) {
    // ev is a ProgressEvent: { File: string, Percent: float64, Done: bool, Error: string }
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

  function handleFileSelect(e: CustomEvent<string[]>) {
    const paths = e.detail;
    const newItems = paths.map(path => {
      // Get base name
      const name = path.split(/[/\\]/).pop() || path;
      return {
        path,
        name,
        size: 0, // Will be filled by Go reports
        percent: 0,
        status: 'waiting' as const,
      };
    });

    // Add only new unique paths
    const existingPaths = new Set(queueItems.map(item => item.path));
    const uniqueNewItems = newItems.filter(item => !existingPaths.has(item.path));

    queueItems = [...queueItems, ...uniqueNewItems];
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
      error: undefined,
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
        Concurrency: concurrency
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

  async function handleOpenOutputFolder() {
    // Open output folder of first successful compression
    const successItem = queueItems.find(item => item.status === 'success');
    if (successItem) {
      // Find output dir. By default it is in compressed/ subdirectory next to source file
      const sourceDir = successItem.path.substring(0, successItem.path.lastIndexOf('\\'));
      const outputDir = `${sourceDir}\\compressed`;
      try {
        await OpenFolder(outputDir);
      } catch (err) {
        console.error("Failed to open folder:", err);
      }
    }
  }
</script>

<div class="app-layout">
  <!-- Top Navigation -->
  <header class="app-header glass-panel">
    <div class="brand">
      <svg class="logo-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <polygon points="23 7 16 12 23 17 23 7"></polygon>
        <rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
      </svg>
      <h1>Videopress</h1>
      <span class="badge">v0.1.0</span>
    </div>
    
    <div class="nav-controls">
      <button 
        class="nav-btn {showSettings ? 'active' : ''}" 
        on:click={() => showSettings = !showSettings}
        disabled={isCompressing}
      >
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="nav-icon">
          <circle cx="12" cy="12" r="3"></circle>
          <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
        </svg>
        配置面板
      </button>
    </div>
  </header>

  <!-- Error banner if FFmpeg missing -->
  {#if ffmpegError}
    <div class="error-banner">
      <svg class="error-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
        <line x1="12" y1="9" x2="12" y2="13"></line>
        <line x1="12" y1="17" x2="12.01" y2="17"></line>
      </svg>
      <span>{ffmpegError}</span>
    </div>
  {/if}

  <!-- Main Grid Area -->
  <div class="workspace">
    {#if showSettings}
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
        <DropZone on:select={handleFileSelect} />
        
        <FileQueue 
          items={queueItems} 
          on:remove={handleRemove} 
          on:clear={handleClear} 
        />

        <!-- Action Panel -->
        <div class="action-panel">
          {#if queueItems.some(item => item.status === 'success')}
            <button class="btn-folder" on:click={handleOpenOutputFolder}>
              打开输出文件夹
            </button>
          {/if}

          <button 
            class="compress-trigger-btn {isCompressing ? 'compressing' : ''}" 
            disabled={queueItems.length === 0 || isCompressing || !!ffmpegError}
            on:click={triggerCompression}
          >
            {#if isCompressing}
              <span class="spinner"></span>
              正在压缩视频...
            {:else}
              开始视频压缩
            {/if}
          </button>
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
    padding: 1rem;
    gap: 1rem;
    max-width: 900px;
    margin: 0 auto;
  }

  .app-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 1.25rem;
    height: 60px;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .logo-icon {
    width: 24px;
    height: 24px;
    color: var(--accent-purple);
  }

  h1 {
    font-size: 1.15rem;
    font-weight: 850;
    letter-spacing: -0.02em;
    background: linear-gradient(135deg, #fff 0%, var(--text-secondary) 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .badge {
    font-size: 0.7rem;
    font-weight: 700;
    background: rgba(168, 85, 247, 0.15);
    color: var(--accent-purple);
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
    border: 1px solid rgba(168, 85, 247, 0.2);
  }

  .nav-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    font-size: 0.8rem;
    font-weight: 600;
    padding: 0.4rem 0.8rem;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .nav-btn:hover {
    border-color: rgba(255, 255, 255, 0.15);
    color: var(--text-primary);
  }

  .nav-btn.active {
    background: var(--accent-purple);
    color: white;
    border-color: var(--accent-purple);
    box-shadow: 0 0 10px rgba(168, 85, 247, 0.3);
  }

  .nav-icon {
    width: 14px;
    height: 14px;
  }

  /* Error Banner */
  .error-banner {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.2);
    padding: 0.75rem 1rem;
    border-radius: 8px;
    color: var(--accent-red);
    font-size: 0.8rem;
    font-weight: 600;
  }

  .error-icon {
    width: 18px;
    height: 18px;
    flex-shrink: 0;
  }

  /* Workspace */
  .workspace {
    flex: 1;
    overflow: hidden;
  }

  .dashboard {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .action-panel {
    display: flex;
    gap: 1rem;
    margin-top: 1rem;
  }

  .compress-trigger-btn {
    flex: 1;
    height: 48px;
    border: none;
    border-radius: 10px;
    background: var(--accent-purple-gradient);
    color: white;
    font-size: 0.95rem;
    font-weight: 750;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    box-shadow: 0 4px 15px rgba(168, 85, 247, 0.3);
    transition: all 0.25s ease;
  }

  .compress-trigger-btn:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(168, 85, 247, 0.45);
  }

  .compress-trigger-btn:active:not(:disabled) {
    transform: translateY(0);
  }

  .compress-trigger-btn:disabled {
    background: var(--border-color);
    color: var(--text-muted);
    cursor: not-allowed;
    box-shadow: none;
  }

  .btn-folder {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    padding: 0 1.25rem;
    border-radius: 10px;
    font-size: 0.85rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .btn-folder:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(255, 255, 255, 0.15);
  }

  .spinner {
    width: 18px;
    height: 18px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top: 2px solid white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }
</style>
