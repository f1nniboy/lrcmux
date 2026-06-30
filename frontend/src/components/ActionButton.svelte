<script lang="ts">
  import type { Snippet } from "svelte";

  interface Props {
    icon: Snippet;
    label: string;
    onclick?: () => void | Promise<void>;
    href?: string;
    download?: boolean | string;
  }

  let {
    icon,
    label,
    onclick = undefined,
    href = undefined,
    download = undefined,
  } = $props();

  let done = $state(false);
  let timer: ReturnType<typeof setTimeout> | undefined;

  async function trigger() {
    try {
      await onclick?.();
      done = true;
      clearTimeout(timer);
      timer = setTimeout(() => (done = false), 1500);
    } catch {
      /* ignore */
    }
  }
</script>

{#snippet body()}
  <span class:invisible={done} class="flex items-center gap-1.5">
    {@render icon()}
    {label}
  </span>
  {#if done}
    <svg
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="3"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"
      class="absolute inset-0 m-auto"
    >
      <polyline points="20 6 9 17 4 12" />
    </svg>
  {/if}
{/snippet}

{#if href != null}
  <a
    {href}
    {download}
    onclick={trigger}
    class="relative inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium shadow-sm no-underline text-ink cursor-pointer"
    class:bg-rule={!done}
    class:bg-cue={done}
    class:text-paper={done}
  >
    {@render body()}
  </a>
{:else}
  <button
    onclick={trigger}
    class="relative inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium shadow-sm no-underline text-ink cursor-pointer"
    class:bg-rule={!done}
    class:bg-cue={done}
    class:text-paper={done}
  >
    {@render body()}
  </button>
{/if}
