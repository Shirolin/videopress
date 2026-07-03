<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { SelectFiles } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();
  let isDragOver = false;

  async function handleClick() {
    try {
      const files = await SelectFiles();
      if (files && files.length > 0) {
        dispatch('select', files);
      }
    } catch (err) {
      console.error("Select files error:", err);
    }
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    isDragOver = true;
  }

  function handleDragLeave() {
    isDragOver = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragOver = false;
    
    // In Wails, dragging files onto the window is handled by HTML5 drag and drop
    // We can extract files from e.dataTransfer
    if (e.dataTransfer && e.dataTransfer.files) {
      const files: string[] = [];
      for (let i = 0; i < e.dataTransfer.files.length; i++) {
        // Under Wails/webview2, HTML5 Drag & Drop file.path gives the absolute path of the file
        const file = e.dataTransfer.files[i];
        if (file.path) {
          files.push(file.path);
        } else {
          // Fallback if path is empty (though webview2 should provide it)
          files.push(file.name);
        }
      }
      if (files.length > 0) {
        dispatch('select', files);
      }
    }
  }
</script>

<div 
  class="drop-zone glass-panel neon-hover {isDragOver ? 'drag-over' : ''}"
  on:click={handleClick}
  on:dragover={handleDragOver}
  on:dragleave={handleDragLeave}
  on:drop={handleDrop}
>
  <div class="icon-container">
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="upload-icon">
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <polyline points="17 8 12 3 7 8" />
      <line x1="12" y1="3" x2="12" y2="15" />
    </svg>
  </div>
  <h3>拖拽视频文件到此处</h3>
  <p>支持多选，或者点击选择文件</p>
  <p class="file-types">支持 MP4, MOV, MKV, AVI, WEBM 等</p>
</div>

<style>
  .drop-zone {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2.5rem;
    border-style: dashed;
    border-width: 2px;
    cursor: pointer;
    text-align: center;
    transition: all 0.25s ease;
    height: 180px;
    margin-bottom: 1.5rem;
  }

  .drag-over {
    border-color: var(--accent-purple);
    background: rgba(168, 85, 247, 0.05);
    box-shadow: 0 0 20px rgba(168, 85, 247, 0.15);
  }

  .icon-container {
    margin-bottom: 1rem;
    color: var(--text-secondary);
    transition: transform 0.2s ease;
  }

  .drop-zone:hover .icon-container {
    transform: translateY(-4px);
    color: var(--accent-purple);
  }

  .upload-icon {
    width: 48px;
    height: 48px;
  }

  h3 {
    font-size: 1.1rem;
    font-weight: 600;
    margin-bottom: 0.25rem;
    color: var(--text-primary);
  }

  p {
    font-size: 0.85rem;
    color: var(--text-secondary);
  }

  .file-types {
    font-size: 0.75rem;
    color: var(--text-muted);
    margin-top: 0.5rem;
  }
</style>
