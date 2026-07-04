<script lang="ts">
  import { onMount } from 'svelte';
  import CustomSelect from './CustomSelect.svelte';
  import { 
    GetPresets, 
    DetectGPUEncoder, 
    InstallSendTo, 
    UninstallSendTo, 
    AddToPath, 
    RemoveFromPath,
    InstallDesktopShortcut,
    UninstallDesktopShortcut,
    InstallStartMenuShortcut,
    UninstallStartMenuShortcut,
    InstallContextMenu,
    UninstallContextMenu,
    GetIntegrationStatus,
    OpenDebugLogFile,
    ClearDebugLogs,
    SetDebugMode
  } from '../../wailsjs/go/main/App.js';

  export let preset: string = 'standard';
  export let concurrency: number = 1;
  export let hwAccel: boolean = false;
  export let copyAudio: boolean = false;
  export let forceMode: boolean = false;
  export let skipExisting: boolean = false;
  export let enableDebugLog: boolean = false;

  $: if (enableDebugLog !== undefined) {
    SetDebugMode(enableDebugLog).catch(console.error);
  }
  
  let presetsList: any[] = [];
  let detectedGPU: string = '';
  let statusMessage: string = '';
  let statusType: 'success' | 'info' | 'error' = 'info';

  // System integration status loading
  let loadingIntegration = true;
  let isSendToInstalled = false;
  let isDesktopInstalled = false;
  let isStartMenuInstalled = false;
  let isContextMenuInstalled = false;
  let isPathConfigured = false;

  const presetLabels: Record<string, string> = {
    small: '小文件规格',
    standard: '标准规格',
    quality: '高画质规格'
  };

  // Map presets list to CustomSelect format
  $: presetOptions = presetsList.map(p => ({
    value: p.name,
    label: presetLabels[p.name] || p.name,
    desc: p.description
  }));

  // Concurrency list for CustomSelect based on actual CPU cores
  const maxCores = typeof navigator !== 'undefined' ? (navigator.hardwareConcurrency || 4) : 4;
  const concurrencyOptions = [
    { value: 1, label: '1 线程', desc: '单任务队列，最佳进度条体验' }
  ];
  if (maxCores >= 2) {
    concurrencyOptions.push({ value: 2, label: '2 并发', desc: '支持 2 任务并发压缩' });
  }
  if (maxCores >= 4) {
    concurrencyOptions.push({ value: 4, label: '4 并发', desc: '支持 4 任务并发压缩' });
  }
  if (maxCores >= 8) {
    concurrencyOptions.push({ value: 8, label: '8 并发', desc: '高并发，建议多核处理器使用' });
  }
  if (maxCores > 8) {
    concurrencyOptions.push({ value: maxCores, label: `${maxCores} 并发`, desc: '利用本机全部 CPU 核心进行高速编码' });
  }

  async function updateIntegrationStatus() {
    try {
      const status = await GetIntegrationStatus();
      isSendToInstalled = status.sendto;
      isDesktopInstalled = status.desktop;
      isStartMenuInstalled = status.startmenu;
      isContextMenuInstalled = status.contextmenu;
      isPathConfigured = status.path;
    } catch (e) {
      console.error("Failed to query integration status:", e);
    }
  }

  onMount(async () => {
    try {
      loadingIntegration = true;
      presetsList = await GetPresets();
      
      // 1. 先更新集成状态，让 UI 上的系统集成（右键菜单、快捷方式等）瞬间渲染出来
      await updateIntegrationStatus();
      loadingIntegration = false; // 集成状态加载完毕，立即关掉 shimmer

      // 2. 异步探测 GPU 编码器，不阻塞其他设置的显示
      DetectGPUEncoder().then(gpu => {
        detectedGPU = gpu;
        if (detectedGPU !== 'libx264') {
          if (localStorage.getItem('videopress_hw_accel') === null) {
            hwAccel = true; // 仅在用户未曾手动设置过硬件加速时，才默认自动开启
          }
        }
      }).catch(gpuErr => {
        console.warn("GPU detector error:", gpuErr);
        detectedGPU = 'libx264';
      });

    } catch (e) {
      console.error("Mount error:", e);
      loadingIntegration = false;
    }
  });

  function showStatus(msg: string, type: 'success' | 'info' | 'error' = 'success') {
    statusMessage = msg;
    statusType = type;
    setTimeout(() => {
      statusMessage = '';
    }, 4000);
  }

  async function handleOpenDebugLog() {
    try {
      await OpenDebugLogFile();
      showStatus("已打开本地调试与性能排查日志", "success");
    } catch (e: any) {
      showStatus("打开日志文件失败: " + e.message, "error");
    }
  }

  async function handleClearDebugLog() {
    try {
      await ClearDebugLogs();
      showStatus("已成功清空日志并重置硬件加速缓存", "success");
      // 重新触发一次后台探测，重新填充缓存
      DetectGPUEncoder().then(gpu => {
        detectedGPU = gpu;
      });
    } catch (e: any) {
      showStatus("清空日志与缓存失败: " + e.message, "error");
    }
  }

  async function toggleSendTo() {
    if (loadingIntegration) return;
    try {
      loadingIntegration = true;
      if (isSendToInstalled) {
        await UninstallSendTo();
        showStatus('已成功移除 SendTo 右键发送快捷方式', 'success');
      } else {
        await InstallSendTo();
        showStatus('已成功创建 SendTo 右键发送快捷方式', 'success');
      }
      await updateIntegrationStatus();
    } catch (e: any) {
      showStatus(`操作失败: ${e.message || e}`, 'error');
    } finally {
      loadingIntegration = false;
    }
  }

  async function toggleDesktopShortcut() {
    if (loadingIntegration) return;
    try {
      loadingIntegration = true;
      if (isDesktopInstalled) {
        await UninstallDesktopShortcut();
        showStatus('已删除桌面快捷方式', 'success');
      } else {
        await InstallDesktopShortcut();
        showStatus('已成功创建桌面快捷方式', 'success');
      }
      await updateIntegrationStatus();
    } catch (e: any) {
      showStatus(`操作失败: ${e.message || e}`, 'error');
    } finally {
      loadingIntegration = false;
    }
  }

  async function toggleStartMenuShortcut() {
    if (loadingIntegration) return;
    try {
      loadingIntegration = true;
      if (isStartMenuInstalled) {
        await UninstallStartMenuShortcut();
        showStatus('已从开始菜单移除快捷方式', 'success');
      } else {
        await InstallStartMenuShortcut();
        showStatus('已成功添加至开始菜单', 'success');
      }
      await updateIntegrationStatus();
    } catch (e: any) {
      showStatus(`操作失败: ${e.message || e}`, 'error');
    } finally {
      loadingIntegration = false;
    }
  }

  async function toggleContextMenu() {
    if (loadingIntegration) return;
    try {
      loadingIntegration = true;
      if (isContextMenuInstalled) {
        await UninstallContextMenu();
        showStatus('已从系统卸载右键直接压缩菜单', 'success');
      } else {
        await InstallContextMenu();
        showStatus('已成功将“使用 Videopress 压缩”加入右键菜单！', 'success');
      }
      await updateIntegrationStatus();
    } catch (e: any) {
      showStatus(`操作失败: ${e.message || e}`, 'error');
    } finally {
      loadingIntegration = false;
    }
  }

  async function togglePathEnv() {
    if (loadingIntegration) return;
    try {
      loadingIntegration = true;
      if (isPathConfigured) {
        await RemoveFromPath();
        showStatus('已从系统环境变量 Path 中移除', 'success');
      } else {
        await AddToPath();
        showStatus('成功添加到系统环境变量 Path', 'success');
      }
      await updateIntegrationStatus();
    } catch (e: any) {
      showStatus(`操作失败: ${e.message || e}`, 'error');
    } finally {
      loadingIntegration = false;
    }
  }
</script>

<div class="settings-container glass-panel">
  <div class="settings-scroll-area">
    <div class="section-title">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="title-icon"><polygon points="6 2 18 2 18 6 6 6 6 2"></polygon><rect x="3" y="6" width="18" height="16" rx="2"></rect><line x1="10" y1="12" x2="14" y2="12"></line></svg>
      压缩核心设置
    </div>
    
    <div class="setting-row">
      <div class="label-group">
        <span class="label">压缩预设规格</span>
        <span class="desc">预设了不同的码率与分辨率配置</span>
      </div>
      <CustomSelect bind:value={preset} options={presetOptions} />
    </div>

    <div class="setting-row">
      <div class="label-group">
        <span class="label">最大并发任务数</span>
        <span class="desc">同时压缩的视频数量，建议根据核心数合理配置</span>
      </div>
      <CustomSelect bind:value={concurrency} options={concurrencyOptions} />
    </div>

    <div class="grid-toggles">
      <label class="toggle-card {hwAccel ? 'checked' : ''}">
        <div class="checkbox-container">
          <input type="checkbox" bind:checked={hwAccel} />
        </div>
        <div class="toggle-content">
          <span class="title">GPU 硬件加速</span>
          <span class="desc">
            {#if detectedGPU && detectedGPU !== 'libx264'}
              已启用显卡加速: <span class="gpu-active">{detectedGPU}</span>
            {:else}
              无显卡加速，自动 Fallback 至 CPU 编码
            {/if}
          </span>
        </div>
      </label>

      <label class="toggle-card {copyAudio ? 'checked' : ''}">
        <div class="checkbox-container">
          <input type="checkbox" bind:checked={copyAudio} />
        </div>
        <div class="toggle-content">
          <span class="title">音频流直通 (Copy Audio)</span>
          <span class="desc">直接复制原片音轨，节省音频重编码开销</span>
        </div>
      </label>

      <label class="toggle-card {forceMode ? 'checked' : ''}">
        <div class="checkbox-container">
          <input type="checkbox" bind:checked={forceMode} />
        </div>
        <div class="toggle-content">
          <span class="title">强制覆盖同名文件</span>
          <span class="desc">直接重写已有的同名输出文件，不生成序号副本</span>
        </div>
      </label>

      <label class="toggle-card {skipExisting ? 'checked' : ''}">
        <div class="checkbox-container">
          <input type="checkbox" bind:checked={skipExisting} />
        </div>
        <div class="toggle-content">
          <span class="title">增量跳过模式</span>
          <span class="desc">如检测到输出文件夹已存在同名压缩文件则跳过</span>
        </div>
      </label>
    </div>

    <div class="section-title margin-top">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="title-icon"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"></rect><rect x="9" y="9" width="6" height="6"></rect><line x1="9" y1="1" x2="9" y2="4"></line><line x1="15" y1="1" x2="15" y2="4"></line><line x1="9" y1="20" x2="9" y2="23"></line><line x1="15" y1="20" x2="15" y2="23"></line><line x1="20" y1="9" x2="23" y2="9"></line><line x1="20" y1="15" x2="23" y2="15"></line><line x1="1" y1="9" x2="4" y2="9"></line><line x1="1" y1="15" x2="4" y2="15"></line></svg>
      Windows 系统集成与快捷设置
    </div>
    <div class="sys-actions">
      <!-- 1. 右键直达菜单 -->
      <div class="action-card {loadingIntegration ? 'loading-shimmer' : ''}">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">右键直接视频压缩 (推荐)</span>
            {#if loadingIntegration}
              <span class="status-badge loading">检测中...</span>
            {:else}
              <span class="status-badge {isContextMenuInstalled ? 'success' : 'muted'}">
                {isContextMenuInstalled ? '已开启' : '已关闭'}
              </span>
            {/if}
          </div>
          <span class="desc">在资源管理器中直接右键点击任意视频文件，直接在菜单选择“使用 Videopress 压缩”。</span>
        </div>
        <div class="action-buttons">
          <button class="btn {isContextMenuInstalled ? 'btn-danger' : 'btn-primary'}" on:click={toggleContextMenu} disabled={loadingIntegration}>
            {#if loadingIntegration}检测中...{:else}{isContextMenuInstalled ? '卸载移除' : '一键开启'}{/if}
          </button>
        </div>
      </div>

      <!-- 2. 桌面快捷方式 -->
      <div class="action-card {loadingIntegration ? 'loading-shimmer' : ''}">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">桌面快捷方式</span>
            {#if loadingIntegration}
              <span class="status-badge loading">检测中...</span>
            {:else}
              <span class="status-badge {isDesktopInstalled ? 'success' : 'muted'}">
                {isDesktopInstalled ? '已创建' : '未创建'}
              </span>
            {/if}
          </div>
          <span class="desc">在 Windows 系统桌面上创建 Videopress 的快捷启动方式。</span>
        </div>
        <div class="action-buttons">
          <button class="btn {isDesktopInstalled ? 'btn-danger' : 'btn-primary'}" on:click={toggleDesktopShortcut} disabled={loadingIntegration}>
            {#if loadingIntegration}检测中...{:else}{isDesktopInstalled ? '删除图标' : '一键创建'}{/if}
          </button>
        </div>
      </div>

      <!-- 3. 开始菜单快捷方式 -->
      <div class="action-card {loadingIntegration ? 'loading-shimmer' : ''}">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">添加至开始菜单</span>
            {#if loadingIntegration}
              <span class="status-badge loading">检测中...</span>
            {:else}
              <span class="status-badge {isStartMenuInstalled ? 'success' : 'muted'}">
                {isStartMenuInstalled ? '已添加' : '未添加'}
              </span>
            {/if}
          </div>
          <span class="desc">在 Windows 开始菜单的程序列表中添加 Videopress，可在搜索框快速搜索唤醒。</span>
        </div>
        <div class="action-buttons">
          <button class="btn {isStartMenuInstalled ? 'btn-danger' : 'btn-primary'}" on:click={toggleStartMenuShortcut} disabled={loadingIntegration}>
            {#if loadingIntegration}检测中...{:else}{isStartMenuInstalled ? '取消固定' : '一键添加'}{/if}
          </button>
        </div>
      </div>

      <!-- 4. 右键 SendTo 发送到菜单 -->
      <div class="action-card {loadingIntegration ? 'loading-shimmer' : ''}">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">发送到快捷菜单 (SendTo)</span>
            {#if loadingIntegration}
              <span class="status-badge loading">检测中...</span>
            {:else}
              <span class="status-badge {isSendToInstalled ? 'success' : 'muted'}">
                {isSendToInstalled ? '已开启' : '已关闭'}
              </span>
            {/if}
          </div>
          <span class="desc">在资源管理器右键选中文件 -> 发送到 -> 快速压缩视频。</span>
        </div>
        <div class="action-buttons">
          <button class="btn {isSendToInstalled ? 'btn-danger' : 'btn-primary'}" on:click={toggleSendTo} disabled={loadingIntegration}>
            {#if loadingIntegration}检测中...{:else}{isSendToInstalled ? '卸载移除' : '一键开启'}{/if}
          </button>
        </div>
      </div>

      <!-- 5. 用户 Path 环境变量 -->
      <div class="action-card {loadingIntegration ? 'loading-shimmer' : ''}">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">配置用户 Path 环境变量</span>
            {#if loadingIntegration}
              <span class="status-badge loading">检测中...</span>
            {:else}
              <span class="status-badge {isPathConfigured ? 'success' : 'muted'}">
                {isPathConfigured ? '已配置' : '未配置'}
              </span>
            {/if}
          </div>
          <span class="desc">把当前程序所在文件夹加入 Path 环境变量，可在任意命令终端直接运行。</span>
        </div>
        <div class="action-buttons">
          <button class="btn {isPathConfigured ? 'btn-danger' : 'btn-primary'}" on:click={togglePathEnv} disabled={loadingIntegration}>
            {#if loadingIntegration}检测中...{:else}{isPathConfigured ? '移除路径' : '一键配置'}{/if}
          </button>
        </div>
      </div>

      <!-- 6. 调试与排查日志 -->
      <div class="section-title margin-top">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="title-icon"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
        调试与排查日志
      </div>

      <div class="action-card">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">环境检测与探测性能日志</span>
            <span class="status-badge {enableDebugLog ? 'success' : 'muted'}">
              {enableDebugLog ? '已启用' : '已禁用'}
            </span>
          </div>
          <span class="desc">开启后将记录 GPU 探测错误、各模块检测耗时及底层 FFMPEG 出错堆栈，关闭则停止写入日志。</span>
        </div>
        <div class="action-buttons">
          <label class="switch">
            <input type="checkbox" bind:checked={enableDebugLog} />
            <span class="slider"></span>
          </label>
        </div>
      </div>

      <div class="action-card">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">日志文件管理</span>
          </div>
          <span class="desc">直接在系统默认编辑器中打开当前日志，或清空积累的探测日志及硬件加速缓存。</span>
        </div>
        <div class="action-buttons">
          <button class="btn btn-secondary" on:click={handleOpenDebugLog} disabled={!enableDebugLog} style="opacity: {enableDebugLog ? 1 : 0.4};">
            打开日志文件
          </button>
          <button class="btn btn-danger" style="margin-left: 0.4rem;" on:click={handleClearDebugLog}>
            清空日志与缓存
          </button>
        </div>
      </div>
    </div>
  </div>

  {#if statusMessage}
    <div class="toast-message {statusType}">
      {#if statusType === 'success'}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="toast-icon"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
      {:else}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="toast-icon"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
      {/if}
      {statusMessage}
    </div>
  {/if}
</div>

<style>
  .settings-container {
    display: flex;
    flex-direction: column;
    height: 100%;          /* Fill workspace area */
    overflow: hidden;
    background: rgba(18, 18, 24, 0.4);
    border-radius: 12px;
  }

  .settings-scroll-area {
    flex: 1;
    overflow-y: auto;      /* Independent scrolling container */
    padding: 1.2rem;
    display: flex;
    flex-direction: column;
    gap: 0.9rem;
    padding-right: 6px;
  }

  /* Allow dropdown popover to overflow individual rows */
  .setting-row {
    position: relative;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1.5rem;
    border-bottom: 1px solid var(--border-color);
    padding-bottom: 0.8rem;
  }

  .section-title {
    font-size: 0.8rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-muted);
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin-bottom: 0.1rem;
  }

  .title-icon {
    width: 14px;
    height: 14px;
    color: var(--accent-purple);
  }

  .margin-top {
    margin-top: 0.6rem;
  }

  .label-group {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  .label {
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .desc {
    font-size: 0.72rem;
    color: var(--text-secondary);
  }

  /* Grid Toggles */
  .grid-toggles {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.8rem;
  }

  @media (max-width: 600px) {
    .grid-toggles {
      grid-template-columns: 1fr;
    }
  }

  .toggle-card {
    display: flex;
    gap: 0.8rem;
    padding: 0.8rem 0.9rem;
    background: rgba(255, 255, 255, 0.015);
    border: 1px solid var(--border-color);
    border-radius: 10px;
    cursor: pointer;
    align-items: flex-start;
    transition: all 0.2s ease;
  }

  .toggle-card:hover {
    background: rgba(255, 255, 255, 0.03);
    border-color: rgba(255, 255, 255, 0.1);
  }

  .toggle-card.checked {
    border-color: rgba(168, 85, 247, 0.25);
    background: rgba(168, 85, 247, 0.02);
  }

  .checkbox-container {
    display: flex;
    align-items: center;
    height: 18px;
  }

  .toggle-card input[type="checkbox"] {
    cursor: pointer;
    accent-color: var(--accent-purple);
    width: 14px;
    height: 14px;
  }

  .toggle-content {
    display: flex;
    flex-direction: column;
    gap: 0.08rem;
  }

  .toggle-content .title {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .toggle-content .desc {
    font-size: 0.68rem;
    line-height: 1.3;
  }

  .gpu-active {
    color: var(--accent-green);
    font-weight: 700;
  }

  /* System actions styling */
  .sys-actions {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .action-card {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1.5rem;
    background: rgba(255, 255, 255, 0.012);
    border: 1px solid var(--border-color);
    padding: 0.8rem 1rem;
    border-radius: 10px;
  }

  @media (max-width: 600px) {
    .action-card {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.8rem;
    }
    .action-buttons {
      width: 100%;
      justify-content: flex-end;
    }
  }

  .action-meta {
    display: flex;
    flex-direction: column;
    gap: 0.08rem;
    flex: 1;
  }

  .action-meta .title {
    font-size: 0.8rem;
    font-weight: 650;
    color: var(--text-primary);
  }

  .action-meta .desc {
    font-size: 0.68rem;
    color: var(--text-secondary);
  }

  .action-buttons {
    display: flex;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  /* Button styles */
  .btn {
    border: none;
    padding: 0.4rem 0.85rem;
    border-radius: 6px;
    font-size: 0.75rem;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: rgba(168, 85, 247, 0.15);
    color: var(--accent-purple);
    border: 1px solid rgba(168, 85, 247, 0.25);
  }

  .btn-primary:hover {
    background: var(--accent-purple);
    color: white;
    border-color: var(--accent-purple);
    box-shadow: 0 0 10px rgba(168, 85, 247, 0.25);
  }

  .btn-muted {
    background: rgba(255, 255, 255, 0.03);
    color: var(--text-secondary);
    border: 1px solid var(--border-color);
  }

  .btn-muted:hover {
    background: rgba(255, 255, 255, 0.08);
    color: var(--text-primary);
  }

  /* Toast notification */
  .toast-message {
    position: absolute;
    bottom: 1.5rem;
    right: 1.5rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: var(--bg-panel-solid);
    border: 1px solid var(--border-color);
    padding: 0.6rem 1rem;
    border-radius: 8px;
    box-shadow: var(--shadow-lg);
    font-size: 0.75rem;
    font-weight: 600;
    z-index: 99;
    animation: slide-in 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .toast-message.success {
    border-color: rgba(16, 185, 129, 0.3);
    color: var(--accent-green);
  }

  .toast-message.error {
    border-color: rgba(244, 63, 94, 0.3);
    color: var(--accent-red);
  }

  .toast-icon {
    width: 14px;
    height: 14px;
  }

  @keyframes slide-in {
    0% { transform: translateY(1rem); opacity: 0; }
    100% { transform: translateY(0); opacity: 1; }
  }

  .action-title-row {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .status-badge {
    font-size: 0.62rem;
    font-weight: 700;
    padding: 0.08rem 0.35rem;
    border-radius: 4px;
  }

  .status-badge.success {
    background: rgba(16, 185, 129, 0.12);
    color: var(--accent-green);
    border: 1px solid rgba(16, 185, 129, 0.2);
  }

  .status-badge.muted {
    background: rgba(255, 255, 255, 0.04);
    color: var(--text-muted);
    border: 1px solid var(--border-color);
  }

  .btn-danger {
    background: rgba(244, 63, 94, 0.08);
    color: var(--accent-red);
    border: 1px solid rgba(244, 63, 94, 0.2);
  }

  .btn-danger:hover {
    background: var(--accent-red);
    color: white;
    border-color: var(--accent-red);
    box-shadow: 0 0 10px rgba(244, 63, 94, 0.25);
  }

  .status-badge.loading {
    background: rgba(168, 85, 247, 0.08);
    color: var(--text-muted);
    border: 1px solid rgba(168, 85, 247, 0.15);
    animation: badge-pulse 1.5s infinite alternate;
  }

  button:disabled {
    opacity: 0.6;
    cursor: not-allowed !important;
    background: rgba(255, 255, 255, 0.02) !important;
    color: var(--text-muted) !important;
    border-color: var(--border-color) !important;
    box-shadow: none !important;
  }

  .loading-shimmer {
    position: relative;
    overflow: hidden;
  }

  .loading-shimmer::after {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: linear-gradient(
      90deg,
      rgba(255, 255, 255, 0) 0%,
      rgba(168, 85, 247, 0.03) 50%,
      rgba(255, 255, 255, 0) 100%
    );
    animation: shimmer-swipe 2s infinite linear;
  }

  @keyframes badge-pulse {
    0% { opacity: 0.6; }
    100% { opacity: 1; }
  }

  @keyframes shimmer-swipe {
    0% { transform: translateX(-100%); }
    100% { transform: translateX(100%); }
  }

  /* Custom toggle switch styling */
  .switch {
    position: relative;
    display: inline-block;
    width: 36px;
    height: 18px;
    flex-shrink: 0;
  }

  .switch input {
    opacity: 0;
    width: 0;
    height: 0;
  }

  .slider {
    position: absolute;
    cursor: pointer;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(255, 255, 255, 0.08);
    transition: .25s ease;
    border-radius: 18px;
    border: 1px solid rgba(255, 255, 255, 0.12);
  }

  .slider:before {
    position: absolute;
    content: "";
    height: 10px;
    width: 10px;
    left: 3px;
    bottom: 3px;
    background-color: var(--text-muted);
    transition: .25s ease;
    border-radius: 50%;
  }

  input:checked + .slider {
    background-color: rgba(168, 85, 247, 0.25);
    border-color: rgba(168, 85, 247, 0.45);
  }

  input:checked + .slider:before {
    transform: translateX(18px);
    background-color: var(--accent-purple);
  }
</style>

