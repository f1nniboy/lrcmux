<script lang="ts">
  import { debounce } from "$lib/utils";
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

  const search = debounce(async (signal: AbortSignal, q: string) => {
    const p = new URLSearchParams({ q, limit: "5" });
    const res = await fetch(`/api/search?${p}`, { signal });
    if (!res.ok) throw new Error("Search failed");
    return res.json() as Promise<Track[]>;
  }, 300);

  const showDropdown = $derived(
    query.trim().length >= 2 && (!!error || tracks.length > 0 || !loading),
  );

  const dropdownClass = $derived(
    slim
      ? "fixed inset-x-0 top-14 border-y border-rule bg-paper shadow-lg z-50 overflow-hidden sm:absolute sm:inset-x-0 sm:top-full sm:mt-1 sm:border sm:rounded-lg"
      : "absolute top-full left-0 right-0 z-50 mt-1 border border-rule rounded-lg bg-paper shadow-lg overflow-hidden",
  );

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
      .run(q)
      .then((r) => {
        tracks = r;
        loading = false;
      })
      .catch((e: Error) => {
        if (e.name === "AbortError") return;
        loading = false;
        error = e.message ?? "Search failed";
      });
  });
</script>

<div class="w-full relative">
  <label
    class="flex items-center {slim
      ? 'gap-2 px-4 py-2.5'
      : 'gap-3 px-5 py-4'} border border-rule rounded-lg focus-within:border-cue transition-colors bg-paper-2"
  >
    <svg
      class="{slim ? 'w-4 h-4' : 'w-5 h-5'} text-muted shrink-0"
      aria-hidden="true"
      fill="none"
      stroke="currentColor"
      stroke-width="2.5"
      viewBox="0 0 24 24"
    >
      <circle cx="11" cy="11" r="7" />
      <path d="m20 20-3.5-3.5" stroke-linecap="round" />
    </svg>
    <span class="sr-only">Search for a track</span>
    <input
      class="flex-1 bg-transparent outline-none text-ink {slim
        ? 'text-sm min-w-0'
        : 'text-base sm:text-lg'} placeholder:text-muted/70"
      autocomplete="off"
      placeholder={slim ? "Search..." : "Search for a track..."}
      spellcheck="false"
      type="search"
      bind:value={query}
    />
    {#if loading}
      <span class="text-muted shrink-0 flex items-center"
        ><Spinner size={18} /></span
      >
    {/if}
  </label>

  {#if showDropdown}
    <div class={dropdownClass}>
      {#if error}
        <p class="px-4 py-3 text-sm text-peak">{error}</p>
      {:else if tracks.length}
        <ul class="divide-y divide-rule">
          {#each tracks as track (track.isrc)}
            <li>
              <TrackItem
                onclick={() => {
                  query = "";
                  tracks = [];
                }}
                {track}
              />
            </li>
          {/each}
        </ul>
      {:else if !loading}
        <p class="px-4 py-3 text-sm text-muted">No matches.</p>
      {/if}
    </div>
  {/if}
</div>
