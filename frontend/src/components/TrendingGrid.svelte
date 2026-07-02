<script module lang="ts">
  import type { Track } from "$lib/types";
  let cache: Track[] = [];
</script>

<script lang="ts">
  import Spinner from "$components/Spinner.svelte";
  import TrackItem from "./TrackItem.svelte";
  import { getTrending, deezerToAPITrack } from "$lib/deezer";

  const COUNT = 12;
  let tracks = $state<Track[]>(cache ?? []);
  let loading = $state(cache.length === 0);

  $effect(() => {
    if (cache.length > 0) return;

    void (async () => {
      try {
        // Deezer chart endpoint doesn't return ISRC, so we have to use the ID as a stable unique key
        tracks = cache = (await getTrending({}, COUNT)).map((dt) => ({
          ...deezerToAPITrack(dt),
          isrc: String(dt.id),
        }));
      } catch {
        tracks = [];
      } finally {
        loading = false;
      }
    })();
  });

  const skeletons = Array.from({ length: COUNT });
</script>

<section>
  <div class="flex items-baseline justify-between mb-5">
    <h2 class="text-lg font-semibold text-ink">Trending now</h2>
    <span class="font-mono font-medium text-muted">Deezer</span>
  </div>

  <div class="relative">
    {#if loading || tracks.length === 0}
      <ul
        class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4"
        class:opacity-30={loading}
        aria-hidden="true"
      >
        {#each skeletons as _, i (i)}
          <li>
            <div class="aspect-square rounded-md bg-paper-2"></div>
            <p class="mt-2 text-sm font-medium text-transparent select-none">
              &nbsp;
            </p>
            <p class="text-xs text-transparent select-none">&nbsp;</p>
          </li>
        {/each}
      </ul>
      {#if loading}
        <div
          class="absolute inset-0 flex items-center justify-center pointer-events-none text-muted"
        >
          <Spinner size={56} />
        </div>
      {/if}
    {:else}
      <ul class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
        {#each tracks as t (t.isrc)}
          <li>
            <TrackItem track={t} variant="tile" />
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</section>
