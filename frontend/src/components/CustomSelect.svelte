<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { slide } from 'svelte/transition';

  interface SelectOption {
    value: any;
    label: string;
    desc?: string;
  }

  let {
    value = $bindable(),
    options = [],
    disabled = false,
    onchange
  }: {
    value?: any;
    options?: SelectOption[];
    disabled?: boolean;
    onchange?: (value: any) => void;
  } = $props();

  let isOpen = $state(false);
  let containerEl: HTMLDivElement;

  let selectedOption = $derived(options.find(opt => opt.value === value) || options[0]);

  function toggleOpen() {
    if (disabled) return;
    isOpen = !isOpen;
  }

  function selectOption(val: any) {
    value = val;
    isOpen = false;
    onchange?.(val);
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
    class="select-trigger hem-panel {isOpen ? 'focused' : ''} {disabled ? 'disabled' : ''}" 
    onclick={toggleOpen}
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
    <div class="options-menu hem-panel" transition:slide={{ duration: 180 }}>
      <div class="options-list">
        {#each options as option}
          <div 
            class="option-item {option.value === value ? 'selected' : ''}" 
            role="option"
            aria-selected={option.value === value}
            tabindex="0"
            onclick={() => selectOption(option.value)}
            onkeydown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                selectOption(option.value);
              }
            }}
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
    min-width: 200px;
    user-select: none;
  }

  /* Trigger styling */
  .select-trigger {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    background: var(--desk-inset) !important;
    border: 1px solid var(--rule);
    padding: 0.5rem 0.9rem;
    border-radius: var(--radius);
    color: var(--ink);
    cursor: pointer;
    text-align: left;
    outline: none;
    font-size: 0.8rem;
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .select-trigger:hover:not(.disabled) {
    border-color: var(--rule-strong);
    background: var(--desk-hem) !important;
  }

  .select-trigger.focused {
    border-color: var(--tungsten);
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
    flex-shrink: 0;
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
    color: var(--tungsten);
  }

  .options-menu {
    position: absolute;
    top: calc(100% + 5px);
    right: 0;
    width: 100%;
    z-index: 999;
    background: var(--desk-hem) !important;
    border: 1px solid var(--rule-strong);
    border-radius: var(--radius);
    box-shadow: var(--shadow-md);
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
    background: var(--desk-inset);
    color: var(--ink);
  }

  .option-item.selected {
    background: var(--desk-inset);
    border-left: 2px solid var(--tungsten);
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
    color: var(--tungsten);
    flex-shrink: 0;
  }
</style>
