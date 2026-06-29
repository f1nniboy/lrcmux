<script lang="ts">
  import { goto } from "$app/navigation";
  import { getTrending } from "$lib/deezer";
  import Spinner from "$components/Spinner.svelte";
  import TrackItem from "./TrackItem.svelte";
  import type { DeezerTrack } from "$lib/types";

  const COUNT = 12;
  let tracks = $state<DeezerTrack[]>([]);
  let loading = $state(true);

  $effect(() => {
    void load();
  });

  async function load() {
    loading = true;
    try {
      tracks = await getTrending(COUNT);
    } catch {
      tracks = [];
    } finally {
      loading = false;
    }
  }

  function pick(t: DeezerTrack) {
    goto(
      `/s/${encodeURIComponent(t.artist.name)}/${encodeURIComponent(t.title)}`,
    );
  }

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
        {#each tracks as t (t.id)}
          <li>
            <TrackItem track={t} variant="tile" onclick={() => pick(t)} />
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</section>
