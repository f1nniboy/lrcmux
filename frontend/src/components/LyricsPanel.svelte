<script lang="ts">
  import type { LyricsResult, Track } from "$lib/types";
  import { toSlug } from "$lib/slug";
  import DownloadPanel from "./DownloadPanel.svelte";

  let {
    track,
    result,
  }: {
    track: Track;
    result: LyricsResult;
  } = $props();

  let artistHref = $derived(`/s/${toSlug(track.artist)}`);
</script>

<section class="flex-1 flex flex-col space-y-6">
  <header class="flex items-start gap-4">
    {#if track.cover.medium}
      <img
        src={track.cover.medium}
        alt=""
        class="w-20 h-20 sm:w-24 sm:h-24 object-cover rounded-md shadow-md shrink-0"
      />
    {:else}
      <div
        class="w-20 h-20 sm:w-24 sm:h-24 rounded-md bg-paper-2 shrink-0"
      ></div>
    {/if}
    <div class="flex-1 min-w-0 pt-1">
      <h2 class="text-2xl sm:text-3xl font-semibold text-ink leading-tight">
        {track.title}
      </h2>
      <a href={artistHref} class="text-muted mt-1 hover:underline no-underline"
        >{track.artist}</a
      >
    </div>
  </header>

  <DownloadPanel {track} {result} />
</section>
