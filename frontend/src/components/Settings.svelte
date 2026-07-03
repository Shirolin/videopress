<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    GetPresets, 
    DetectGPUEncoder, 
    InstallSendTo, 
    UninstallSendTo, 
    AddToPath, 
    RemoveFromPath 
  } from '../../wailsjs/go/main/App.js';

  export let preset: string = 'standard';
  export let concurrency: number = 1;
  export let hwAccel: boolean = false;
  export let copyAudio: boolean = false;
  export let forceMode: boolean = false;
  export let skipExisting: boolean = false;
  
  let presetsList: any[] = [];
  let detectedGPU: string = '';
  let statusMessage: string = '';
  let statusType: 'success' | 'info' | 'error' = 'info';

  onMount(async () => {
    try {
      presetsList = await GetPresets();
      detectedGPU = await DetectGPUEncoder();
      if (detectedGPU !== 'libx264') {
        hwAccel = true; // Auto-enable if GPU encoder found
      }
    } catch (e) {
      console.error("Mount error:", e);
    }
  });

  function showStatus(msg: string, type: 'success' | 'info' | 'error' = 'success') {
    statusMessage = msg;
    statusType = type;
    setTimeout(() => {
      statusMessage = '';
    }, 4000);
  }

  async function handleInstallSendTo() {
    try {
      const res = await InstallSendTo();
      showStatus(`右键菜单安装成功: ${res}`, 'success');
    } catch (e: any) {
      showStatus(`安装失败: ${e}`, 'error');
    }
  }

  async function handleUninstallSendTo() {
    try {
      await UninstallSendTo();
      showStatus('右键菜单已成功卸载', 'success');
    } catch (e: any) {
      showStatus(`卸载失败: ${e}`, 'error');
    }
  }

  async function handleAddToPath() {
    try {
      const added = await AddToPath();
      if (added) {
        showStatus('成功添加到系统环境变量 Path', 'success');
      } else {
        showStatus('系统环境变量 Path 中已存在该路径', 'info');
      }
    } catch (e: any) {
      showStatus(`配置失败: ${e}`, 'error');
    }
  }

  async function handleRemoveFromPath() {
    try {
      const removed = await RemoveFromPath();
      if (removed) {
        showStatus('已成功从系统环境变量 Path 中移除', 'success');
      } else {
        showStatus('系统环境变量 Path 中未找到该路径', 'info');
      }
    } catch (e: any) {
      showStatus(`移除失败: ${e}`, 'error');
    }
  }
</script>

<div class="settings-container glass-panel">
  <div class="section-title">压缩设置</div>
  
  <div class="setting-row">
    <div class="label-group">
      <span class="label">压缩规格</span>
      <span class="desc">选择视频输出的分辨率和码率预设</span>
    </div>
    <select bind:value={preset}>
      {#each presetsList as p}
        <option value={p.name}>{p.name} - {p.description}</option>
      {/each}
    </select>
  </div>

  <div class="setting-row">
    <div class="label-group">
      <span class="label">最大并发任务数</span>
      <span class="desc">同时压缩的视频数量，建议根据CPU核心数设置</span>
    </div>
    <select bind:value={concurrency}>
      <option value={1}>1 (单线程，完美进度条)</option>
      <option value={2}>2</option>
      <option value={4}>4 (多任务并发)</option>
      <option value={8}>8</option>
    </select>
  </div>

  <div class="grid-toggles">
    <label class="toggle-card">
      <input type="checkbox" bind:checked={hwAccel} />
      <div class="toggle-content">
        <span class="title">GPU 硬件加速</span>
        <span class="desc">
          {#if detectedGPU && detectedGPU !== 'libx264'}
            已检测到可用 GPU 编码: <span class="gpu-active">{detectedGPU}</span>
          {:else}
            未检测到 GPU 编码，使用 CPU (libx264)
          {/if}
        </span>
      </div>
    </label>

    <label class="toggle-card">
      <input type="checkbox" bind:checked={copyAudio} />
      <div class="toggle-content">
        <span class="title">音频流直通</span>
        <span class="desc">不重新编码音频，保留原音质并提升速度</span>
      </div>
    </label>

    <label class="toggle-card">
      <input type="checkbox" bind:checked={forceMode} />
      <div class="toggle-content">
        <span class="title">强制覆盖同名</span>
        <span class="desc">覆盖已存在的同名压缩文件，不产生递增副本</span>
      </div>
    </label>

    <label class="toggle-card">
      <input type="checkbox" bind:checked={skipExisting} />
      <div class="toggle-content">
        <span class="title">增量跳过模式</span>
        <span class="desc">如果输出目录已存在同名压缩文件，则直接跳过</span>
      </div>
    </label>
  </div>

  <div class="section-title margin-top">Windows 系统集成</div>
  <div class="sys-actions">
    <div class="action-card">
      <div class="action-meta">
        <span class="title">右键“发送到”菜单</span>
        <span class="desc">一键将系统右键发送到绑定至本软件，右键点击视频即可压缩。</span>
      </div>
      <div class="action-buttons">
        <button class="btn btn-secondary" on:click={handleInstallSendTo}>安装</button>
        <button class="btn btn-muted" on:click={handleUninstallSendTo}>卸载</button>
      </div>
    </div>

    <div class="action-card">
      <div class="action-meta">
        <span class="title">环境变量 Path 绑定</span>
        <span class="desc">自动将程序所在路径添加至系统的 Path 中，方便从任何终端调用。</span>
      </div>
      <div class="action-buttons">
        <button class="btn btn-secondary" on:click={handleAddToPath}>配置</button>
        <button class="btn btn-muted" on:click={handleRemoveFromPath}>移除</button>
      </div>
    </div>
  </div>

  {#if statusMessage}
    <div class="toast-message {statusType}">
      {statusMessage}
    </div>
  {/if}
</div>

<style>
  .settings-container {
    padding: 1.2rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    height: 100%;
    overflow-y: auto;
  }

  .section-title {
    font-size: 0.85rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    border-left: 2px solid var(--accent-purple);
    padding-left: 0.5rem;
    margin-bottom: 0.25rem;
  }

  .margin-top {
    margin-top: 0.75rem;
  }

  .setting-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1.5rem;
    border-bottom: 1px solid var(--border-color);
    padding-bottom: 0.75rem;
  }

  .label-group {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }

  .label {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .desc {
    font-size: 0.75rem;
    color: var(--text-secondary);
  }

  select {
    background: var(--bg-primary);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    padding: 0.4rem 0.6rem;
    border-radius: 6px;
    outline: none;
    font-size: 0.8rem;
    min-width: 160px;
    cursor: pointer;
  }

  select:focus {
    border-color: var(--border-focus);
  }

  /* Grid Toggles */
  .grid-toggles {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }

  .toggle-card {
    display: flex;
    gap: 0.75rem;
    padding: 0.75rem;
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    cursor: pointer;
    align-items: flex-start;
    transition: all 0.2s ease;
  }

  .toggle-card:hover {
    background: rgba(255, 255, 255, 0.04);
    border-color: rgba(255, 255, 255, 0.15);
  }

  .toggle-card input[type="checkbox"] {
    margin-top: 0.2rem;
    cursor: pointer;
    accent-color: var(--accent-purple);
  }

  .toggle-content {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  .toggle-content .title {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .gpu-active {
    color: var(--accent-green);
    font-weight: 600;
  }

  /* Sys Actions */
  .sys-actions {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .action-card {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem;
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    gap: 1.5rem;
  }

  .action-meta {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
    flex: 1;
  }

  .action-meta .title {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .action-buttons {
    display: flex;
    gap: 0.5rem;
  }

  .btn {
    border: none;
    outline: none;
    font-size: 0.75rem;
    font-weight: 600;
    padding: 0.4rem 0.8rem;
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.2s ease;
  }

  .btn-secondary {
    background: var(--accent-purple);
    color: white;
  }

  .btn-secondary:hover {
    background: #9333ea;
  }

  .btn-muted {
    background: rgba(255, 255, 255, 0.05);
    color: var(--text-secondary);
    border: 1px solid var(--border-color);
  }

  .btn-muted:hover {
    background: rgba(255, 255, 255, 0.08);
  }

  /* Toast Message */
  .toast-message {
    position: fixed;
    bottom: 1.5rem;
    right: 1.5rem;
    padding: 0.6rem 1rem;
    border-radius: 8px;
    font-size: 0.8rem;
    font-weight: 600;
    box-shadow: var(--shadow-lg);
    border: 1px solid rgba(255, 255, 255, 0.05);
    z-index: 1000;
    animation: slideUp 0.3s ease-out;
  }

  .toast-message.success {
    background: #064e3b;
    color: #34d399;
  }

  .toast-message.info {
    background: #1e3a8a;
    color: #60a5fa;
  }

  .toast-message.error {
    background: #7f1d1d;
    color: #f87171;
  }

  @keyframes slideUp {
    from {
      transform: translateY(20px);
      opacity: 0;
    }
    to {
      transform: translateY(0);
      opacity: 1;
    }
  }
</style>
