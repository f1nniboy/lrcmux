<script lang="ts">
  import { makeDebouncedSearch, deezerToAPITrack } from "$lib/deezer";
  import type { Track } from "$lib/types";
  import Spinner from "$components/Spinner.svelte";
  import TrackItem from "./TrackItem.svelte";

  interface Props {
    slim?: boolean;
  }
  let { slim = false }: Props = $props();

  let query = $state("");
  let tracks = $state<Track[]>([]);
  let loading = $state(false);
  let error = $state<string | null>(null);

  const search = makeDebouncedSearch(300);

  $effect(() => {
    const q = query.trim();
    if (q.length < 2) {
      tracks = [];
      loading = false;
      error = null;
      search.abort();
      return;
    }
    loading = true;
    error = null;
    search
      .run(q, { limit: 8 })
      .then((r) => {
        tracks = r.map(deezerToAPITrack);
        loading = false;
      })
      .catch((e: Error) => {
        loading = false;
        error = e.message ?? "Search failed";
      });
  });
</script>

<div class="w-full relative">
  <label
    class="flex items-center {slim
      ? 'gap-2 px-4 py-2.5'
      : 'gap-3 px-5 py-4'} border border-rule rounded-lg focus-within:border-cue transition-colors bg-paper-2/40"
  >
    <svg
      class="{slim ? 'w-4 h-4' : 'w-5 h-5'} text-muted shrink-0"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      aria-hidden="true"
    >
      <circle cx="11" cy="11" r="7" />
      <path d="m20 20-3.5-3.5" stroke-linecap="round" />
    </svg>
    <span class="sr-only">Search for a song</span>
    <input
      bind:value={query}
      type="search"
      autocomplete="off"
      spellcheck="false"
      placeholder={slim ? "search..." : "search by artist, song, or ISRC"}
      class="flex-1 bg-transparent outline-none text-ink {slim
        ? 'text-sm min-w-0'
        : 'text-base sm:text-lg'} placeholder:text-muted/70"
    />
    {#if loading}
      <span class="text-muted shrink-0 flex items-center"
        ><Spinner size={18} /></span
      >
    {/if}
  </label>

  {#if error}
    <p
      class="absolute top-full left-0 right-0 z-50 mt-1 text-sm text-peak bg-paper border border-rule rounded-lg px-4 py-3 shadow-lg"
    >
      {error}
    </p>
  {/if}

  {#if tracks.length}
    <ul
      class="absolute top-full left-0 right-0 z-50 mt-1 border border-rule rounded-lg divide-y divide-rule overflow-hidden bg-paper shadow-lg"
    >
      {#each tracks as track (track.isrc)}
        <li>
          <TrackItem
            {track}
            variant="row"
            onclick={() => {
              query = "";
              tracks = [];
            }}
          />
        </li>
      {/each}
    </ul>
  {:else if query.trim().length >= 2 && !loading && !error}
    <p
      class="absolute top-full left-0 right-0 z-50 mt-1 text-sm text-muted bg-paper border border-rule rounded-lg px-4 py-3 shadow-lg"
    >
      No matches.
    </p>
  {/if}
</div>
