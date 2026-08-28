<script module lang="ts">
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
</script>

<script lang="ts">
  import { t } from '../i18n.js';

  let {
    items = [],
    isCompressing = false,
    onremove,
    onclear
  }: {
    items?: QueueItem[];
    isCompressing?: boolean;
    onremove?: (index: number) => void;
    onclear?: () => void;
  } = $props();

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
</script>

<div class="queue-container hem-panel">
  <div class="queue-header">
    <span class="title">{$t('queue.title')} <span class="counter">{items.length}</span></span>
    {#if items.length > 0}
      <button class="clear-btn" onclick={() => onclear?.()} disabled={isCompressing}>{$t('queue.clear')}</button>
    {/if}
  </div>

  {#if items.length === 0}
    <div class="empty-state">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="empty-icon">
        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
      </svg>
      <p>{$t('queue.empty')}</p>
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
              <button class="remove-btn" onclick={() => onremove?.(index)} title={$t('queue.remove_task')}>
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
                {$t('queue.status_waiting')}
              {:else if item.status === 'compressing'}
                {item.percent.toFixed(0)}%
              {:else if item.status === 'success'}
                {$t('queue.percent_done')}
              {:else if item.status === 'skipped'}
                {$t('queue.status_skipped')}
              {:else if item.status === 'failed'}
                {$t('queue.compress_failed_simple')}
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
                  <span class="label">{$t('queue.origin_size')}</span>
                  <span class="val">{formatSize(item.size)}</span>
                </div>
                <div class="stat">
                  <span class="label">{$t('queue.target_size')}</span>
                  <span class="val text-green">{formatSize(item.targetSize || 0)}</span>
                </div>
                <div class="stat">
                  <span class="label">{$t('queue.save_ratio')}</span>
                  <span class="val ratio-badge">{formatRatio(item.size, item.targetSize)}</span>
                </div>
                <div class="stat">
                  <span class="label">{$t('queue.duration')}</span>
                  <span class="val">{formatTime(item.duration)}</span>
                </div>
              {:else}
                <div class="status-msg">
                  {#if item.status === 'skipped'}
                    <span class="skipped-txt">
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="inline-icon"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                      {$t('queue.skip_tip')}
                    </span>
                  {:else}
                    <span class="failed-txt" title={item.error}>
                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="inline-icon"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                      {$t('queue.fail_reason', {err: item.error || 'Unknown error'})}
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
    flex: 1;
    min-height: 0;
    padding: 1.2rem;
    overflow: hidden;
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
    color: var(--ink);
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .counter {
    font-size: 0.72rem;
    padding: 0.1rem 0.4rem;
    background: var(--desk-inset);
    border-radius: 3px;
    color: var(--ink-dim);
    border: 1px solid var(--rule);
    font-family: var(--font-mono);
  }

  .clear-btn {
    background: transparent;
    border: none;
    color: var(--ink-dim);
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    transition: color 0.15s;
  }

  .clear-btn:hover:not(:disabled) {
    color: var(--danger);
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
    background: var(--desk-inset);
    border: 1px solid var(--rule);
    border-left: 3px solid var(--rule-strong);
    border-radius: var(--radius);
    padding: 0.8rem 0.9rem;
    display: flex;
    flex-direction: column;
    gap: 0.55rem;
    transition: background 0.15s ease, border-color 0.15s ease;
  }

  .queue-item:hover {
    background: var(--desk-hem);
    border-color: var(--rule-strong);
  }

  .queue-item.status-compressing {
    border-left-color: var(--tungsten);
  }

  .queue-item.status-success {
    border-left-color: var(--ok);
  }

  .queue-item.status-skipped {
    border-left-color: var(--warn);
  }

  .queue-item.status-failed {
    border-left-color: var(--danger);
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
    color: var(--ink);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .remove-btn {
    background: transparent;
    border: none;
    color: var(--ink-faint);
    cursor: pointer;
    display: flex;
    align-items: center;
    padding: 3px;
    border-radius: 3px;
    transition: background 0.15s, color 0.15s;
  }

  .remove-btn:hover {
    background: var(--desk-hem);
    color: var(--danger);
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
    height: 6px;
    background: var(--desk);
    border-radius: 3px;
    overflow: hidden;
    position: relative;
    border: 1px solid var(--rule);
  }

  .progress-fill {
    height: 100%;
    transition: width 200ms cubic-bezier(0.16, 1, 0.3, 1);
  }

  .progress-fill.status-waiting {
    background: var(--rule-strong);
  }

  .progress-fill.status-compressing {
    background: var(--tungsten);
  }

  .progress-fill.status-success {
    background: var(--tungsten);
  }

  .progress-fill.status-skipped {
    background: var(--warn);
  }

  .progress-fill.status-failed {
    background: var(--danger);
  }

  .percent-label {
    font-size: 0.72rem;
    font-weight: 700;
    color: var(--ink-faint);
    width: 65px;
    text-align: right;
    white-space: nowrap;
    font-family: var(--font-mono);
  }

  .percent-label.status-compressing {
    color: var(--tungsten);
  }

  .percent-label.status-success {
    color: var(--ok);
  }

  .percent-label.status-skipped {
    color: var(--warn);
  }

  .percent-label.status-failed {
    color: var(--danger);
  }

  /* Result details grid */
  .result-details {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.5rem;
    background: var(--desk);
    padding: 0.5rem 0.75rem;
    border-radius: var(--radius);
    border: 1px solid var(--rule);
  }

  .stat .label {
    font-size: 0.65rem;
    font-weight: 500;
    color: var(--ink-faint);
  }

  .stat .val {
    font-size: 0.78rem;
    font-weight: 650;
    color: var(--ink-dim);
    font-family: var(--font-mono);
  }

  .text-green { color: var(--ok) !important; }
  
  .ratio-badge {
    color: var(--ink-dim) !important;
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
    color: var(--warn);
    font-weight: 550;
  }

  .failed-txt {
    color: var(--danger);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
    font-weight: 550;
  }

  .warning-box {
    background: var(--desk) !important;
    border: 1px solid var(--warn) !important;
  }

  .warning-txt {
    color: var(--warn);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: block;
    font-weight: 550;
  }
</style>
