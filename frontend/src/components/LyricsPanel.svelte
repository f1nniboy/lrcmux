<script lang="ts">
  import type { DeezerTrack, LyricsResult } from "$lib/types";
  import DownloadPanel from "./DownloadPanel.svelte";

  let { track, result }: { track: DeezerTrack; result: LyricsResult } =
    $props();
</script>

<section class="space-y-6">
  <header class="flex items-start gap-4">
    {#if track.album.cover_medium}
      <img
        src={track.album.cover_medium}
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
      <p class="text-muted mt-1">{track.artist.name}</p>
    </div>
  </header>

  <DownloadPanel {track} {result} />
</section>
