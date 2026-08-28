<script lang="ts">
  import { onMount } from 'svelte';
  import { t } from '../i18n.js';
  import {
    WindowMinimise,
    WindowToggleMaximise,
    WindowIsMaximised,
    Quit,
  } from '../../wailsjs/runtime/runtime.js';

  let { maximised = $bindable(false) } = $props();

  async function syncMaximised() {
    maximised = await WindowIsMaximised();
  }

  onMount(() => {
    syncMaximised();
  });

  function handleMinimise() {
    WindowMinimise();
  }

  async function handleToggleMaximise() {
    WindowToggleMaximise();
    await syncMaximised();
  }

  function handleClose() {
    Quit();
  }
</script>

<div class="window-controls" style="--wails-draggable:no-drag">
  <button
    type="button"
    class="win-btn"
    aria-label={$t('window.minimize')}
    onclick={handleMinimise}
  >
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 12 12" aria-hidden="true">
      <rect x="1.5" y="5.5" width="9" height="1" fill="currentColor" />
    </svg>
  </button>
  <button
    type="button"
    class="win-btn"
    aria-label={maximised ? $t('window.restore') : $t('window.maximize')}
    onclick={handleToggleMaximise}
  >
    {#if maximised}
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 12 12" aria-hidden="true">
        <rect x="3" y="1.5" width="7.5" height="7.5" fill="none" stroke="currentColor" stroke-width="1" />
        <path d="M1.5 4.5v6h6" fill="none" stroke="currentColor" stroke-width="1" />
      </svg>
    {:else}
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 12 12" aria-hidden="true">
        <rect x="1.5" y="1.5" width="9" height="9" fill="none" stroke="currentColor" stroke-width="1" />
      </svg>
    {/if}
  </button>
  <button
    type="button"
    class="win-btn win-btn-close"
    aria-label={$t('window.close')}
    onclick={handleClose}
  >
    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 12 12" aria-hidden="true">
      <path d="M2.5 2.5l7 7M9.5 2.5l-7 7" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" />
    </svg>
  </button>
</div>

<style>
  .window-controls {
    display: flex;
    align-items: stretch;
    height: 100%;
    flex-shrink: 0;
  }

  .win-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 100%;
    padding: 0;
    border: none;
    border-left: 1px solid var(--rule);
    background: transparent;
    color: var(--ink-dim);
    cursor: pointer;
    transition: background 0.12s ease, color 0.12s ease;
  }

  .win-btn svg {
    width: 12px;
    height: 12px;
  }

  .win-btn:hover {
    background: var(--desk-inset);
    color: var(--tungsten);
  }

  .win-btn-close:hover {
    background: var(--danger);
    color: var(--ink);
  }
</style>
