<script lang="ts">
  import type { Track } from "$lib/types";

  interface Props {
    track: Track;
    variant?: "row" | "tile";
    onclick?: () => void;
  }
  let { track, variant = "row", onclick }: Props = $props();

  const href = $derived(
    `/s/${encodeURIComponent(track.artist)}/${encodeURIComponent(track.title)}`,
  );

  function fmtDuration(s: number): string {
    const m = Math.floor(s / 60);
    const sec = (s % 60).toString().padStart(2, "0");
    return `${m}:${sec}`;
  }
</script>

{#if variant === "tile"}
  <a {href} {onclick} class="block w-full text-left group cursor-pointer">
    {#if track.cover.medium}
      <img
        src={track.cover.medium}
        alt=""
        loading="lazy"
        class="w-full aspect-square object-cover rounded-md bg-paper-2 shadow-sm group-hover:shadow-lg group-hover:-translate-y-0.5 transition-all duration-200"
      />
    {:else}
      <div class="w-full aspect-square rounded-md bg-paper-2"></div>
    {/if}
    <p
      class="mt-2 text-sm font-medium text-ink truncate group-hover:text-cue transition-colors"
    >
      {track.title}
    </p>
    <p class="text-xs text-muted truncate">{track.artist}</p>
  </a>
{:else}
  <a
    {href}
    {onclick}
    class="w-full flex items-center gap-3 px-3 py-2.5 text-left hover:bg-cue/10 transition-colors cursor-pointer"
  >
    {#if track.cover.small}
      <img
        src={track.cover.small}
        alt=""
        class="w-10 h-10 rounded object-cover bg-paper-2"
        loading="lazy"
      />
    {:else}
      <div class="w-10 h-10 rounded bg-paper-2"></div>
    {/if}
    <div class="flex-1 min-w-0">
      <p class="text-sm font-medium truncate text-ink">{track.title}</p>
      <p class="text-xs text-muted truncate">{track.artist}</p>
    </div>
    <span class="font-mono text-xs text-muted tabular-nums"
      >{fmtDuration(track.duration)}</span
    >
  </a>
{/if}
