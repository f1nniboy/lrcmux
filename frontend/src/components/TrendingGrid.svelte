<script lang="ts">
  import type { Track } from "$lib/types";
  import TrackItem from "./TrackItem.svelte";

  let { tracks }: { tracks: Track[] } = $props();
</script>

{#if tracks.length > 0}
  <section>
    <div class="flex items-baseline justify-between mb-5">
      <h2 class="text-lg font-semibold text-ink">Trending now</h2>
      <span class="font-mono font-medium text-muted">Deezer</span>
    </div>
    <ul class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
      {#each tracks as t (t.isrc)}
        <li>
          <TrackItem track={t} variant="tile" />
        </li>
      {/each}
    </ul>
  </section>
{/if}
