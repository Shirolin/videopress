<script lang="ts">
  import { onMount } from 'svelte';
  import CustomSelect from './CustomSelect.svelte';
  import { t, locale } from '../i18n.ts';
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
    SetDebugMode,
    SetLanguage
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

  // 当界面语言发生改变时，通知 Go 后端热重写注册表并重新加载预设信息
  $: if ($locale) {
    SetLanguage($locale).catch(console.error);
    GetPresets().then(list => {
      presetsList = list;
    }).catch(console.error);
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

  $: presetLabels = {
    small: $t('status.preset_small'),
    standard: $t('status.preset_standard'),
    quality: $t('status.preset_quality')
  };

  // Map presets list to CustomSelect format
  $: presetOptions = presetsList.map(p => ({
    value: p.name,
    label: presetLabels[p.name] || p.name,
    desc: p.description
  }));

  // Concurrency list for CustomSelect based on actual CPU cores
  const maxCores = typeof navigator !== 'undefined' ? (navigator.hardwareConcurrency || 4) : 4;
  let concurrencyOptions: any[] = [];
  $: {
    concurrencyOptions = [
      { value: 1, label: $t('settings.concurrency_1'), desc: $t('settings.concurrency_1_desc') }
    ];
    if (maxCores >= 2) {
      concurrencyOptions.push({ value: 2, label: $t('settings.concurrency_n', {num: 2}), desc: $t('settings.concurrency_n_desc', {num: 2}) });
    }
    if (maxCores >= 4) {
      concurrencyOptions.push({ value: 4, label: $t('settings.concurrency_n', {num: 4}), desc: $t('settings.concurrency_n_desc', {num: 4}) });
    }
    if (maxCores >= 8) {
      concurrencyOptions.push({ value: 8, label: $t('settings.concurrency_n', {num: 8}), desc: $t('settings.concurrency_n_desc', {num: 8}) });
    }
    if (maxCores > 8) {
      concurrencyOptions.push({ value: maxCores, label: $t('settings.concurrency_n', {num: maxCores}), desc: $t('settings.concurrency_all_desc') });
    }
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
      showStatus($t('toast.log_opened'), "success");
    } catch (e: any) {
      showStatus($t('toast.log_open_failed', {err: e.message}), "error");
    }
  }

  async function handleClearDebugLog() {
    try {
      await ClearDebugLogs();
      showStatus($t('toast.log_cleared'), "success");
      // 重新触发一次后台探测，重新填充缓存
      DetectGPUEncoder().then(gpu => {
        detectedGPU = gpu;
      });
    } catch (e: any) {
      showStatus($t('toast.log_clear_failed', {err: e.message}), "error");
    }
  }

  async function toggleSendTo() {
    if (loadingIntegration) return;
    try {
      loadingIntegration = true;
      if (isSendToInstalled) {
        await UninstallSendTo();
        showStatus($t('toast.sendto_removed'), 'success');
      } else {
        await InstallSendTo();
        showStatus($t('toast.sendto_added'), 'success');
      }
      await updateIntegrationStatus();
    } catch (e: any) {
      showStatus($t('toast.action_failed', {err: e.message || e}), 'error');
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
        showStatus($t('toast.desktop_removed'), 'success');
      } else {
        await InstallDesktopShortcut();
        showStatus($t('toast.desktop_added'), 'success');
      }
      await updateIntegrationStatus();
    } catch (e: any) {
      showStatus($t('toast.action_failed', {err: e.message || e}), 'error');
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
        showStatus($t('toast.startmenu_removed'), 'success');
      } else {
        await InstallStartMenuShortcut();
        showStatus($t('toast.startmenu_added'), 'success');
      }
      await updateIntegrationStatus();
    } catch (e: any) {
      showStatus($t('toast.action_failed', {err: e.message || e}), 'error');
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
        showStatus($t('toast.contextmenu_removed'), 'success');
      } else {
        await InstallContextMenu();
        showStatus($t('toast.contextmenu_added'), 'success');
      }
      await updateIntegrationStatus();
    } catch (e: any) {
      showStatus($t('toast.action_failed', {err: e.message || e}), 'error');
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
        showStatus($t('toast.path_removed'), 'success');
      } else {
        await AddToPath();
        showStatus($t('toast.path_added'), 'success');
      }
      await updateIntegrationStatus();
    } catch (e: any) {
      showStatus($t('toast.action_failed', {err: e.message || e}), 'error');
    } finally {
      loadingIntegration = false;
    }
  }
</script>

<div class="settings-container glass-panel">
  <div class="settings-scroll-area">
    <div class="section-title">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="title-icon"><polygon points="6 2 18 2 18 6 6 6 6 2"></polygon><rect x="3" y="6" width="18" height="16" rx="2"></rect><line x1="10" y1="12" x2="14" y2="12"></line></svg>
      {$t('settings.title.general')}
    </div>
    
    <div class="setting-row">
      <div class="label-group">
        <span class="label">{$t('settings.preset')}</span>
        <span class="desc">{$t('settings.preset_desc')}</span>
      </div>
      <CustomSelect bind:value={preset} options={presetOptions} />
    </div>

    <div class="setting-row">
      <div class="label-group">
        <span class="label">{$t('settings.concurrency')}</span>
        <span class="desc">{$t('settings.concurrency_desc')}</span>
      </div>
      <CustomSelect bind:value={concurrency} options={concurrencyOptions} />
    </div>

    <div class="grid-toggles">
      <label class="toggle-card {hwAccel ? 'checked' : ''}">
        <div class="checkbox-container">
          <input type="checkbox" bind:checked={hwAccel} />
        </div>
        <div class="toggle-content">
          <span class="title">{$t('settings.hw_accel')}</span>
          <span class="desc">
            {#if detectedGPU && detectedGPU !== 'libx264'}
              {$t('settings.hw_accel_active', {gpu: detectedGPU})}
            {:else}
              {$t('settings.hw_accel_inactive')}
            {/if}
          </span>
        </div>
      </label>

      <label class="toggle-card {copyAudio ? 'checked' : ''}">
        <div class="checkbox-container">
          <input type="checkbox" bind:checked={copyAudio} />
        </div>
        <div class="toggle-content">
          <span class="title">{$t('settings.copy_audio')}</span>
          <span class="desc">{$t('settings.copy_audio_desc')}</span>
        </div>
      </label>

      <label class="toggle-card {forceMode ? 'checked' : ''}">
        <div class="checkbox-container">
          <input type="checkbox" bind:checked={forceMode} />
        </div>
        <div class="toggle-content">
          <span class="title">{$t('settings.force_mode')}</span>
          <span class="desc">{$t('settings.force_mode_desc')}</span>
        </div>
      </label>

      <label class="toggle-card {skipExisting ? 'checked' : ''}">
        <div class="checkbox-container">
          <input type="checkbox" bind:checked={skipExisting} />
        </div>
        <div class="toggle-content">
          <span class="title">{$t('settings.skip_existing')}</span>
          <span class="desc">{$t('settings.skip_existing_desc')}</span>
        </div>
      </label>
    </div>

    <div class="section-title margin-top">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="title-icon"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"></rect><rect x="9" y="9" width="6" height="6"></rect><line x1="9" y1="1" x2="9" y2="4"></line><line x1="15" y1="1" x2="15" y2="4"></line><line x1="9" y1="20" x2="9" y2="23"></line><line x1="15" y1="20" x2="15" y2="23"></line><line x1="20" y1="9" x2="23" y2="9"></line><line x1="20" y1="15" x2="23" y2="15"></line><line x1="1" y1="9" x2="4" y2="9"></line><line x1="1" y1="15" x2="4" y2="15"></line></svg>
      {$t('settings.title.sys')}
    </div>
    <div class="sys-actions">
      <!-- 1. 右键直达菜单 -->
      <div class="action-card {loadingIntegration ? 'loading-shimmer' : ''}">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">{$t('settings.contextmenu')}</span>
            {#if loadingIntegration}
              <span class="status-badge loading">{$t('status.checking')}</span>
            {:else}
              <span class="status-badge {isContextMenuInstalled ? 'success' : 'muted'}">
                {isContextMenuInstalled ? $t('status.enabled') : $t('status.disabled')}
              </span>
            {/if}
          </div>
          <span class="desc">{$t('settings.contextmenu_desc')}</span>
        </div>
        <div class="action-buttons">
          <button class="btn {isContextMenuInstalled ? 'btn-danger' : 'btn-primary'}" on:click={toggleContextMenu} disabled={loadingIntegration}>
            {#if loadingIntegration}{$t('status.checking')}{:else}{isContextMenuInstalled ? $t('btn.disable') : $t('btn.enable')}{/if}
          </button>
        </div>
      </div>

      <!-- 2. 桌面快捷方式 -->
      <div class="action-card {loadingIntegration ? 'loading-shimmer' : ''}">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">{$t('settings.desktop')}</span>
            {#if loadingIntegration}
              <span class="status-badge loading">{$t('status.checking')}</span>
            {:else}
              <span class="status-badge {isDesktopInstalled ? 'success' : 'muted'}">
                {isDesktopInstalled ? $t('status.created') : $t('status.not_created')}
              </span>
            {/if}
          </div>
          <span class="desc">{$t('settings.desktop_desc')}</span>
        </div>
        <div class="action-buttons">
          <button class="btn {isDesktopInstalled ? 'btn-danger' : 'btn-primary'}" on:click={toggleDesktopShortcut} disabled={loadingIntegration}>
            {#if loadingIntegration}{$t('status.checking')}{:else}{isDesktopInstalled ? $t('btn.delete_icon') : $t('btn.create')}{/if}
          </button>
        </div>
      </div>

      <!-- 3. 开始菜单快捷方式 -->
      <div class="action-card {loadingIntegration ? 'loading-shimmer' : ''}">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">{$t('settings.startmenu')}</span>
            {#if loadingIntegration}
              <span class="status-badge loading">{$t('status.checking')}</span>
            {:else}
              <span class="status-badge {isStartMenuInstalled ? 'success' : 'muted'}">
                {isStartMenuInstalled ? $t('status.added') : $t('status.not_added')}
              </span>
            {/if}
          </div>
          <span class="desc">{$t('settings.startmenu_desc')}</span>
        </div>
        <div class="action-buttons">
          <button class="btn {isStartMenuInstalled ? 'btn-danger' : 'btn-primary'}" on:click={toggleStartMenuShortcut} disabled={loadingIntegration}>
            {#if loadingIntegration}{$t('status.checking')}{:else}{isStartMenuInstalled ? $t('btn.remove') : $t('btn.add')}{/if}
          </button>
        </div>
      </div>

      <!-- 4. 右键 SendTo 发送到菜单 -->
      <div class="action-card {loadingIntegration ? 'loading-shimmer' : ''}">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">{$t('settings.sendto')}</span>
            {#if loadingIntegration}
              <span class="status-badge loading">{$t('status.checking')}</span>
            {:else}
              <span class="status-badge {isSendToInstalled ? 'success' : 'muted'}">
                {isSendToInstalled ? $t('status.enabled') : $t('status.disabled')}
              </span>
            {/if}
          </div>
          <span class="desc">{$t('settings.sendto_desc')}</span>
        </div>
        <div class="action-buttons">
          <button class="btn {isSendToInstalled ? 'btn-danger' : 'btn-primary'}" on:click={toggleSendTo} disabled={loadingIntegration}>
            {#if loadingIntegration}{$t('status.checking')}{:else}{isSendToInstalled ? $t('btn.disable') : $t('btn.enable')}{/if}
          </button>
        </div>
      </div>

      <!-- 5. 用户 Path 环境变量 -->
      <div class="action-card {loadingIntegration ? 'loading-shimmer' : ''}">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">{$t('settings.path')}</span>
            {#if loadingIntegration}
              <span class="status-badge loading">{$t('status.checking')}</span>
            {:else}
              <span class="status-badge {isPathConfigured ? 'success' : 'muted'}">
                {isPathConfigured ? $t('status.configured') : $t('status.not_configured')}
              </span>
            {/if}
          </div>
          <span class="desc">{$t('settings.path_desc')}</span>
        </div>
        <div class="action-buttons">
          <button class="btn {isPathConfigured ? 'btn-danger' : 'btn-primary'}" on:click={togglePathEnv} disabled={loadingIntegration}>
            {#if loadingIntegration}{$t('status.checking')}{:else}{isPathConfigured ? $t('btn.remove_path') : $t('btn.configure')}{/if}
          </button>
        </div>
      </div>

      <!-- 界面语言设置 -->
      <div class="section-title margin-top">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="title-icon"><circle cx="12" cy="12" r="10"></circle><line x1="2" y1="12" x2="22" y2="12"></line><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"></path></svg>
        {$t('settings.title.lang')}
      </div>
      <div class="action-card">
        <div class="action-meta">
          <span class="desc">{$t('settings.lang.desc')}</span>
        </div>
        <div class="action-buttons">
          <div class="segmented-control" style="width: 200px;">
            <button class="segment-btn {$locale === 'zh' ? 'active' : ''}" on:click={() => $locale = 'zh'}>简体中文</button>
            <button class="segment-btn {$locale === 'en' ? 'active' : ''}" on:click={() => $locale = 'en'}>English</button>
          </div>
        </div>
      </div>

      <!-- 6. 调试与排查日志 -->
      <div class="section-title margin-top">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="title-icon"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
        {$t('settings.title.debug')}
      </div>

      <div class="action-card">
        <div class="action-meta">
          <div class="action-title-row">
            <span class="title">{$t('settings.debug.toggle')}</span>
            <span class="status-badge {enableDebugLog ? 'success' : 'muted'}">
              {enableDebugLog ? $t('status.enabled_simple') : $t('status.disabled_simple')}
            </span>
          </div>
          <span class="desc">{$t('settings.debug.desc')}</span>
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
            <span class="title">{$t('settings.debug.manage')}</span>
          </div>
          <span class="desc">{$t('settings.debug.file_desc')}</span>
        </div>
        <div class="action-buttons">
          <button class="btn btn-secondary" on:click={handleOpenDebugLog} disabled={!enableDebugLog} style="opacity: {enableDebugLog ? 1 : 0.4};">
            {$t('btn.open_log')}
          </button>
          <button class="btn btn-danger" style="margin-left: 0.4rem;" on:click={handleClearDebugLog}>
            {$t('btn.clear_log')}
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

