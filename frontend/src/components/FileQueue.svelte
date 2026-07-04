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
    isWarning?: boolean;
  }

  export let items: QueueItem[] = [];
  export let isCompressing = false;
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
    <span class="title">任务队列 <span class="counter">{items.length}</span></span>
    {#if items.length > 0}
      <button class="clear-btn" on:click={() => dispatch('clear')} disabled={isCompressing}>清空队列</button>
    {/if}
  </div>

  {#if items.length === 0}
    <div class="empty-state">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="empty-icon">
        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
      </svg>
      <p>暂无待压缩的视频任务</p>
    </div>
  {:else}
    <div class="list-wrapper">
      {#each items as item, index}
        <div class="queue-item status-{item.status}">
          <div class="item-meta">
            <div class="item-title">
              <span class="dot dot-{item.status}"></span>
              <span class="file-name" title={item.path}>{item.name}</span>
            </div>
            {#if item.status === 'waiting' && !isCompressing}
              <button class="remove-btn" on:click={() => handleRemove(index)} title="移除任务">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="close-icon"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
              </button>
            {/if}
          </div>

          <!-- 常驻进度条 -->
          <div class="progress-container">
            <div class="progress-bar">
              <div class="progress-fill status-{item.status}" style="width: {item.status === 'success' || item.status === 'skipped' ? 100 : (item.status === 'failed' ? 100 : item.percent)}%"></div>
            </div>
            <span class="percent-label status-{item.status}">
              {#if item.status === 'waiting'}
                等待中
              {:else if item.status === 'compressing'}
                {item.percent.toFixed(0)}%
              {:else if item.status === 'success'}
                100% (完成)
              {:else if item.status === 'skipped'}
                已跳过
              {:else if item.status === 'failed'}
                失败
              {/if}
            </span>
          </div>
          
          {#if item.isWarning && item.status === 'waiting'}
            <div class="result-details warning-box">
              <div class="status-msg">
                <span class="warning-txt" title={item.error}>
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="inline-icon"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                  {item.error}
                </span>
              </div>
            </div>
          {/if}

          {#if item.status === 'success' || item.status === 'skipped' || item.status === 'failed'}
            <div class="result-details">
              {#if item.status === 'success'}
                <div class="stat">
                  <span class="label">原始大小</span>
                  <span class="val">{formatSize(item.size)}</span>
                </div>
                <div class="stat">
                  <span class="label">压缩后</span>
                  <span class="val text-green">{formatSize(item.targetSize || 0)}</span>
                </div>
                <div class="stat">
                  <span class="label">节省率</span>
                  <span class="val ratio-badge">{formatRatio(item.size, item.targetSize)}</span>
                </div>
                <div class="stat">
                  <span class="label">耗时</span>
                  <span class="val">{formatTime(item.duration)}</span>
                </div>
              {:else}
                <div class="status-msg">
                  {#if item.status === 'skipped'}
                    <span class="skipped-txt">
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="inline-icon"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                      已跳过：输出目录已存在同名文件
                    </span>
                  {:else}
                    <span class="failed-txt" title={item.error}>
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="inline-icon"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                      失败原因：{item.error || '未知错误'}
                    </span>
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
    flex: 1;               /* Let container fill the remaining space */
    min-height: 0;         /* Crucial for flex box nesting scroll to work */
    padding: 1.2rem;
    overflow: hidden;
    background: rgba(18, 18, 24, 0.4);
    border-radius: 12px;
  }

  .queue-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.8rem;
    border-bottom: 1px solid var(--border-color);
    padding-bottom: 0.6rem;
    flex-shrink: 0;
  }

  .title {
    font-size: 0.85rem;
    font-weight: 700;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .counter {
    font-size: 0.72rem;
    padding: 0.1rem 0.4rem;
    background: rgba(255, 255, 255, 0.06);
    border-radius: 20px;
    color: var(--text-secondary);
  }

  .clear-btn {
    background: transparent;
    border: none;
    color: var(--accent-magenta);
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.2s;
  }

  .clear-btn:hover {
    opacity: 0.85;
    text-decoration: underline;
  }

  .clear-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
    text-decoration: none !important;
  }

  .empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
    gap: 0.6rem;
  }

  .empty-icon {
    width: 32px;
    height: 32px;
    opacity: 0.4;
  }

  .empty-state p {
    font-size: 0.8rem;
  }

  .list-wrapper {
    flex: 1;
    overflow-y: auto;      /* Enable local vertical scrolling */
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding-right: 4px;
  }

  .queue-item {
    background: rgba(255, 255, 255, 0.015);
    border: 1px solid var(--border-color);
    border-left: 3.5px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 0.8rem 0.9rem;
    display: flex;
    flex-direction: column;
    gap: 0.55rem;
    transition: all 0.2s ease;
  }

  .queue-item:hover {
    background: rgba(255, 255, 255, 0.03);
    border-color: rgba(255, 255, 255, 0.08);
  }

  .queue-item.status-compressing {
    border-left-color: var(--accent-purple);
    background: rgba(168, 85, 247, 0.02);
  }

  .queue-item.status-success {
    border-left-color: var(--accent-green);
    background: rgba(16, 185, 129, 0.015);
  }

  .queue-item.status-skipped {
    border-left-color: var(--accent-yellow);
    background: rgba(245, 158, 11, 0.015);
  }

  .queue-item.status-failed {
    border-left-color: var(--accent-red);
    background: rgba(244, 63, 94, 0.015);
  }

  .item-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .item-title {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    overflow: hidden;
    flex: 1;
  }

  .file-name {
    font-size: 0.82rem;
    font-weight: 550;
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
    padding: 3px;
    border-radius: 4px;
    transition: background 0.15s, color 0.15s;
  }

  .remove-btn:hover {
    background: rgba(244, 63, 94, 0.1);
    color: var(--accent-red);
  }

  .close-icon {
    width: 13px;
    height: 13px;
  }

  /* Progress Bar styling */
  .progress-container {
    display: flex;
    align-items: center;
    gap: 0.8rem;
    margin-top: 0.1rem;
  }

  .progress-bar {
    flex: 1;
    height: 7px;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 10px;
    overflow: hidden;
    position: relative;
  }

  .progress-fill {
    height: 100%;
    transition: width 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .progress-fill.status-waiting {
    background: rgba(255, 255, 255, 0.08);
  }

  .progress-fill.status-compressing {
    background: linear-gradient(90deg, #a855f7, #d946ef, #a855f7);
    background-size: 200px 100%;
    animation: shimmer-move 1.5s linear infinite;
  }

  .progress-fill.status-success {
    background: var(--accent-green);
    box-shadow: 0 0 8px rgba(16, 185, 129, 0.3);
  }

  .progress-fill.status-skipped {
    background: var(--accent-yellow);
  }

  .progress-fill.status-failed {
    background: var(--accent-red);
  }

  .percent-label {
    font-size: 0.72rem;
    font-weight: 700;
    color: var(--text-muted);
    width: 65px;
    text-align: right;
    white-space: nowrap;
  }

  .percent-label.status-compressing {
    color: var(--accent-purple);
  }

  .percent-label.status-success {
    color: var(--accent-green);
  }

  .percent-label.status-skipped {
    color: var(--accent-yellow);
  }

  .percent-label.status-failed {
    color: var(--accent-red);
  }

  /* Result details grid */
  .result-details {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.5rem;
    background: rgba(0, 0, 0, 0.12);
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    border: 1px solid rgba(255, 255, 255, 0.02);
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 0.08rem;
  }

  .stat .label {
    font-size: 0.65rem;
    font-weight: 500;
    color: var(--text-muted);
  }

  .stat .val {
    font-size: 0.78rem;
    font-weight: 650;
    color: var(--text-secondary);
  }

  .text-green { color: var(--accent-green) !important; }
  
  .ratio-badge {
    color: var(--accent-magenta) !important;
    font-weight: 700 !important;
  }

  .status-msg {
    grid-column: span 4;
    font-size: 0.75rem;
    display: flex;
    align-items: center;
  }

  .inline-icon {
    width: 12px;
    height: 12px;
    display: inline-block;
    margin-right: 4px;
    vertical-align: text-bottom;
  }

  .skipped-txt {
    color: var(--accent-yellow);
    font-weight: 550;
  }

  .failed-txt {
    color: var(--accent-red);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
    font-weight: 550;
  }

  .warning-box {
    background: rgba(245, 158, 11, 0.05) !important;
    border: 1px solid rgba(245, 158, 11, 0.15) !important;
  }

  .warning-txt {
    color: var(--accent-yellow);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
    font-weight: 550;
  }
</style>
