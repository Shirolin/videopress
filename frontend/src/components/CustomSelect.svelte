<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { slide, fade } from 'svelte/transition';

  interface SelectOption {
    value: any;
    label: string;
    desc?: string;
  }

  export let value: any;
  export let options: SelectOption[] = [];
  export let disabled = false;

  let isOpen = false;
  let containerEl: HTMLDivElement;

  $: selectedOption = options.find(opt => opt.value === value) || options[0];

  function toggleOpen() {
    if (disabled) return;
    isOpen = !isOpen;
  }

  function selectOption(val: any) {
    value = val;
    isOpen = false;
  }

  // Handle click outside to close the menu
  function handleClickOutside(event: MouseEvent) {
    if (containerEl && !containerEl.contains(event.target as Node)) {
      isOpen = false;
    }
  }

  onMount(() => {
    window.addEventListener('click', handleClickOutside);
  });

  onDestroy(() => {
    window.removeEventListener('click', handleClickOutside);
  });
</script>

<div class="custom-select-container" bind:this={containerEl}>
  <!-- Trigger Button -->
  <button 
    type="button"
    class="select-trigger glass-panel {isOpen ? 'focused' : ''} {disabled ? 'disabled' : ''}" 
    on:click={toggleOpen}
    {disabled}
  >
    <div class="trigger-label">
      <span class="selected-text">{selectedOption ? selectedOption.label : ''}</span>
      {#if selectedOption && selectedOption.desc}
        <span class="selected-desc">— {selectedOption.desc}</span>
      {/if}
    </div>
    <div class="arrow-icon {isOpen ? 'rotated' : ''}">
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <polyline points="6 9 12 15 18 9"></polyline>
      </svg>
    </div>
  </button>

  <!-- Options Menu -->
  {#if isOpen}
    <div class="options-menu glass-panel" transition:slide={{ duration: 180 }}>
      <div class="options-list">
        {#each options as option}
          <div 
            class="option-item {option.value === value ? 'selected' : ''}" 
            on:click={() => selectOption(option.value)}
          >
            <div class="option-text-group">
              <span class="option-label">{option.label}</span>
              {#if option.desc}
                <span class="option-desc">{option.desc}</span>
              {/if}
            </div>
            {#if option.value === value}
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" class="check-icon">
                <polyline points="20 6 9 17 4 12"></polyline>
              </svg>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .custom-select-container {
    position: relative;
    min-width: 260px;
    user-select: none;
  }

  /* Trigger styling */
  .select-trigger {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: rgba(18, 18, 24, 0.55) !important;
    border: 1px solid var(--border-color);
    padding: 0.5rem 0.9rem;
    border-radius: 8px;
    color: var(--text-primary);
    cursor: pointer;
    text-align: left;
    outline: none;
    font-size: 0.8rem;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .select-trigger:hover:not(.disabled) {
    border-color: rgba(255, 255, 255, 0.15);
    background: rgba(255, 255, 255, 0.02) !important;
  }

  .select-trigger.focused {
    border-color: var(--accent-purple);
    box-shadow: 0 0 12px rgba(168, 85, 247, 0.2);
  }

  .select-trigger.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .trigger-label {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    margin-right: 0.8rem;
  }

  .selected-text {
    font-weight: 700;
    color: var(--text-primary);
  }

  .selected-desc {
    font-size: 0.72rem;
    color: var(--text-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .arrow-icon {
    width: 12px;
    height: 12px;
    color: var(--text-secondary);
    transition: transform 0.2s;
    display: flex;
    align-items: center;
    flex-shrink: 0;
  }

  .arrow-icon.rotated {
    transform: rotate(180deg);
    color: var(--accent-purple);
  }

  /* Popover list styling */
  .options-menu {
    position: absolute;
    top: calc(100% + 5px);
    left: 0;
    width: 100%;
    z-index: 999;
    background: rgba(12, 12, 16, 0.92) !important;
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.5), 0 8px 10px -6px rgba(0, 0, 0, 0.5);
    overflow: hidden;
  }

  .options-list {
    max-height: 250px;
    overflow-y: auto;
    padding: 0.25rem;
  }

  .option-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s ease;
    gap: 1rem;
  }

  .option-item:hover {
    background: rgba(168, 85, 247, 0.1);
    color: var(--text-primary);
  }

  .option-item.selected {
    background: rgba(168, 85, 247, 0.2);
    border-left: 2.5px solid var(--accent-purple);
    border-top-left-radius: 2px;
    border-bottom-left-radius: 2px;
  }

  .option-text-group {
    display: flex;
    flex-direction: column;
    gap: 0.08rem;
    overflow: hidden;
  }

  .option-label {
    font-size: 0.78rem;
    font-weight: 650;
    color: var(--text-primary);
  }

  .option-desc {
    font-size: 0.68rem;
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .check-icon {
    width: 12px;
    height: 12px;
    color: var(--accent-purple);
    flex-shrink: 0;
  }
</style>
