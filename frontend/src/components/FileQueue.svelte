<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  export interface QueueItem {
    path: string;
    name: string;
    size: number;
    percent: number;
    status: 'waiting' | 'compressing' | 'success' | 'failed' | 'skipped';
    targetSize?: number;
    duration?: number; // in ms
    error?: string;
  }

  export let items: QueueItem[] = [];
  const dispatch = createEventDispatcher();

  function formatSize(bytes: number): string {
    if (!bytes || bytes <= 0) return '-';
    const mb = bytes / (1024 * 1024);
    return mb.toFixed(1) + ' MB';
  }

  function formatRatio(source: number, target?: number): string {
    if (!target || source <= 0) return '-';
    const saved = ((source - target) / source) * 100;
    return saved.toFixed(1) + '%';
  }

  function formatTime(ms?: number): string {
    if (!ms) return '-';
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  function handleRemove(index: number) {
    dispatch('remove', index);
  }
</script>

<div class="queue-container glass-panel">
  <div class="queue-header">
    <span class="title">任务队列 ({items.length})</span>
    {#if items.length > 0}
      <button class="clear-btn" on:click={() => dispatch('clear')}>清空</button>
    {/if}
  </div>

  {#if items.length === 0}
    <div class="empty-state">
      <p>暂无待压缩任务</p>
    </div>
  {:else}
    <div class="list-wrapper">
      {#each items as item, index}
        <div class="queue-item {item.status}">
          <div class="item-meta">
            <div class="item-title">
              <span class="dot dot-{item.status}"></span>
              <span class="file-name" title={item.path}>{item.name}</span>
            </div>
            {#if item.status === 'waiting'}
              <button class="remove-btn" on:click={() => handleRemove(index)}>
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="close-icon"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
              </button>
            {/if}
          </div>

          {#if item.status === 'compressing'}
            <div class="progress-container">
              <div class="progress-bar">
                <div class="progress-fill" style="width: {item.percent}%"></div>
              </div>
              <span class="percent-label">{item.percent.toFixed(0)}%</span>
            </div>
          {/if}

          {#if item.status === 'success' || item.status === 'skipped' || item.status === 'failed'}
            <div class="result-details">
              {#if item.status === 'success'}
                <div class="stat">
                  <span class="label">原始大小:</span>
                  <span class="val">{formatSize(item.size)}</span>
                </div>
                <div class="stat">
                  <span class="label">压缩后:</span>
                  <span class="val accent-green">{formatSize(item.targetSize || 0)}</span>
                </div>
                <div class="stat">
                  <span class="label">节省率:</span>
                  <span class="val accent-purple">{formatRatio(item.size, item.targetSize)}</span>
                </div>
                <div class="stat">
                  <span class="label">耗时:</span>
                  <span class="val">{formatTime(item.duration)}</span>
                </div>
              {:else}
                <div class="status-msg">
                  {#if item.status === 'skipped'}
                    <span class="skipped-txt">跳过：输出文件已存在</span>
                  {:else}
                    <span class="failed-txt" title={item.error}>失败：{item.error || '未知错误'}</span>
                  {/if}
                </div>
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .queue-container {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 350px);
    min-height: 250px;
    padding: 1rem;
    overflow: hidden;
  }

  .queue-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.75rem;
    border-bottom: 1px solid var(--border-color);
    padding-bottom: 0.5rem;
  }

  .title {
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .clear-btn {
    background: transparent;
    border: none;
    color: var(--accent-magenta);
    font-size: 0.8rem;
    cursor: pointer;
  }

  .clear-btn:hover {
    text-decoration: underline;
  }

  .empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
    font-size: 0.85rem;
  }

  .list-wrapper {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding-right: 2px;
  }

  .queue-item {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    transition: background 0.2s ease;
  }

  .queue-item:hover {
    background: rgba(255, 255, 255, 0.04);
  }

  .item-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .item-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    overflow: hidden;
    flex: 1;
  }

  .file-name {
    font-size: 0.85rem;
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .remove-btn {
    background: transparent;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    display: flex;
    align-items: center;
    padding: 2px;
    border-radius: 4px;
  }

  .remove-btn:hover {
    background: rgba(255, 255, 255, 0.05);
    color: var(--accent-red);
  }

  .close-icon {
    width: 14px;
    height: 14px;
  }

  /* Progress */
  .progress-container {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .progress-bar {
    flex: 1;
    height: 6px;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: var(--accent-purple-gradient);
    border-radius: 3px;
    transition: width 0.1s linear;
  }

  .percent-label {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--accent-purple);
    width: 32px;
    text-align: right;
  }

  /* Result details */
  .result-details {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.5rem;
    background: rgba(0, 0, 0, 0.15);
    padding: 0.4rem 0.6rem;
    border-radius: 6px;
    border: 1px solid rgba(255, 255, 255, 0.02);
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
  }

  .stat .label {
    font-size: 0.7rem;
    color: var(--text-muted);
  }

  .stat .val {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-secondary);
  }

  .accent-green { color: var(--accent-green) !important; }
  .accent-purple { color: var(--accent-purple) !important; }

  .status-msg {
    grid-column: span 4;
    font-size: 0.75rem;
  }

  .skipped-txt {
    color: var(--accent-yellow);
  }

  .failed-txt {
    color: var(--accent-red);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
  }
</style>
