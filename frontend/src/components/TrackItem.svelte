<script lang="ts">
  import type { Track } from "$lib/types";
  import { resolve } from "$app/paths";
  import { toSlug } from "$lib/slug";

  interface Props {
    track: Track;
    onclick?: () => void;
  }
  let { track, onclick }: Props = $props();

  function fmtDuration(s: number): string {
    const m = Math.floor(s / 60);
    const sec = (s % 60).toString().padStart(2, "0");
    return `${m}:${sec}`;
  }
</script>

<a
  class="w-full flex items-center gap-3 px-3 py-2.5 text-left hover:bg-cue/10 transition-colors cursor-pointer"
  href={resolve("/s/[artist]/[title]", {
    artist: toSlug(track.artist),
    title: toSlug(track.title),
  })}
  {onclick}
>
  {#if track.cover.small}
    <img
      class="w-10 h-10 rounded object-cover bg-paper-2"
      alt=""
      loading="lazy"
      src={track.cover.small}
    />
  {:else}
    <div class="w-10 h-10 rounded bg-paper-2"></div>
  {/if}
  <div class="flex-1 min-w-0">
    <p class="text-sm font-medium truncate text-ink">{track.title}</p>
    <p class="text-xs text-muted truncate">{track.artist}</p>
  </div>
  <span class="text-xs text-muted tabular-nums"
    >{fmtDuration(track.duration)}</span
  >
</a>
