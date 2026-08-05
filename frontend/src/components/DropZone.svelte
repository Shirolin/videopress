<script lang="ts">
  import { SelectFiles } from '../../wailsjs/go/main/App.js';
  import { t } from '../i18n.js';

  let {
    compact = false,
    disabled = false,
    onselect
  }: {
    compact?: boolean;
    disabled?: boolean;
    onselect?: (files: string[]) => void;
  } = $props();

  let isDragOver = $state(false);

  async function handleClick() {
    if (disabled) return;
    try {
      const files = await SelectFiles();
      if (files && files.length > 0) {
        onselect?.(files);
      }
    } catch (err) {
      console.error("Select files error:", err);
    }
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    if (disabled) return;
    isDragOver = true;
  }

  function handleDragLeave() {
    isDragOver = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragOver = false;
    if (disabled) return;
    
    if (e.dataTransfer && e.dataTransfer.files) {
      const files: string[] = [];
      for (let i = 0; i < e.dataTransfer.files.length; i++) {
        const file = e.dataTransfer.files[i];
        // Wails 拖拽注入的绝对路径位于 File.path 非标准字段上
        const filePath = (file as any).path;
        if (filePath) {
          files.push(filePath);
        } else {
          files.push(file.name);
        }
      }
      if (files.length > 0) {
        onselect?.(files);
      }
    }
  }
</script>

<div 
  class="drop-zone glass-panel neon-hover {compact ? 'compact' : ''} {isDragOver ? 'drag-over' : ''} {disabled ? 'disabled' : ''}"
  role="button"
  tabindex={disabled ? -1 : 0}
  aria-disabled={disabled}
  onclick={handleClick}
  onkeydown={(e) => {
    if (disabled) return;
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick();
    }
  }}
  ondragover={handleDragOver}
  ondragleave={handleDragLeave}
  ondrop={handleDrop}
>
  {#if disabled}
    <div class="compact-content">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="upload-icon-small" style="color: var(--text-muted)">
        <circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line>
      </svg>
      <span class="compact-text" style="color: var(--text-muted)">{$t('dropzone.disabled')}</span>
    </div>
  {:else if compact}
    <div class="compact-content">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="upload-icon-small">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
        <polyline points="17 8 12 3 7 8" />
        <line x1="12" y1="3" x2="12" y2="15" />
      </svg>
      <span class="compact-text">{$t('dropzone.add_more')}</span>
    </div>
  {:else}
    <div class="icon-container">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="upload-icon">
        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
        <polyline points="17 8 12 3 7 8" />
        <line x1="12" y1="3" x2="12" y2="15" />
      </svg>
    </div>
    <h3>{$t('dropzone.title')}</h3>
    <p>{$t('dropzone.desc')}</p>
    <p class="file-types">{$t('dropzone.formats')}</p>
  {/if}
</div>

<style>
  .drop-zone {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2.2rem;
    border-style: dashed;
    border-width: 1.5px;
    border-color: rgba(255, 255, 255, 0.15);
    cursor: pointer;
    text-align: center;
    transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
    height: 180px;
    background: rgba(255, 255, 255, 0.015);
    border-radius: 12px;
  }

  .drop-zone.compact {
    height: 52px;
    padding: 0;
    border-width: 1px;
    background: rgba(168, 85, 247, 0.03);
    border-color: rgba(168, 85, 247, 0.2);
  }

  .drop-zone.compact:hover {
    border-color: rgba(168, 85, 247, 0.45);
    background: rgba(168, 85, 247, 0.06);
    box-shadow: 0 0 12px rgba(168, 85, 247, 0.1);
  }

  .drag-over {
    border-color: var(--accent-purple) !important;
    background: rgba(168, 85, 247, 0.08) !important;
    box-shadow: 0 0 20px rgba(168, 85, 247, 0.2) !important;
  }

  .compact-content {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .compact-text {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--accent-purple);
  }

  .upload-icon-small {
    width: 15px;
    height: 15px;
    color: var(--accent-purple);
  }

  .icon-container {
    margin-bottom: 0.8rem;
    color: var(--text-secondary);
    transition: transform 0.25s ease;
  }

  .drop-zone:hover:not(.compact) .icon-container {
    transform: translateY(-4px);
    color: var(--accent-purple);
  }

  .upload-icon {
    width: 44px;
    height: 44px;
  }

  h3 {
    font-size: 1.05rem;
    font-weight: 600;
    margin-bottom: 0.25rem;
    color: var(--text-primary);
    letter-spacing: -0.01em;
  }

  p {
    font-size: 0.8rem;
    color: var(--text-secondary);
  }

  .file-types {
    font-size: 0.72rem;
    color: var(--text-muted);
    margin-top: 0.4rem;
  }

  .drop-zone.disabled {
    opacity: 0.55;
    cursor: not-allowed;
    border-style: solid;
    border-color: rgba(255, 255, 255, 0.08) !important;
    background: rgba(255, 255, 255, 0.005) !important;
    box-shadow: none !important;
  }
  .drop-zone.disabled .compact-text {
    color: var(--text-muted) !important;
  }
  .drop-zone.disabled .upload-icon-small,
  .drop-zone.disabled .upload-icon {
    color: var(--text-muted) !important;
  }
</style>

