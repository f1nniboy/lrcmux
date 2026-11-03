<script lang="ts">
  import { Search } from "@lucide/svelte";
  import { goto } from "$app/navigation";
  import { debounce } from "$lib/utils";
  import { resolve } from "$app/paths";
  import { toSlug } from "$lib/slug";
  import type { Track } from "$lib/types";
  import Spinner from "./Spinner.svelte";
  import TrackRow from "./TrackRow.svelte";

  let { large = false }: { large?: boolean } = $props();

  const id = $props.id();

  let root = $state<HTMLDivElement>();
  let query = $state("");
  let tracks = $state<Track[]>([]);
  // the highlighted result, none until the mouse or the arrow keys pick one
  let active = $state(-1);
  let loading = $state(false);
  let error = $state<string>();
  let open = $state(false);

  const search = debounce(async (signal: AbortSignal, q: string) => {
    const res = await fetch(
      `/api/search?${new URLSearchParams({ q, limit: "5" })}`,
      { signal },
    );
    if (!res.ok) throw new Error("Search failed");
    return res.json() as Promise<Track[]>;
  }, 300);

  const showResults = $derived(
    open &&
      query.trim().length >= 2 &&
      (!!error || tracks.length > 0 || !loading),
  );
  const listed = $derived(showResults && !error && tracks.length > 0);
  const status = $derived.by(() => {
    if (!showResults) return "";
    if (error) return error;
    if (!tracks.length) return "No matches.";
    return tracks.length === 1 ? "1 result" : `${tracks.length} results`;
  });

  $effect(() => {
    const q = query.trim();
    if (q.length < 2) {
      tracks = [];
      loading = false;
      error = undefined;
      search.abort();
      return;
    }
    loading = true;
    error = undefined;
    search.run(q).then(
      (r) => {
        tracks = r;
        active = -1;
        loading = false;
      },
      (e: Error) => {
        if (e.name === "AbortError") return;
        loading = false;
        error = e.message;
      },
    );
  });

  function close() {
    open = false;
    query = "";
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === "Escape") return close();
    if (!listed) return;
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
      e.preventDefault();
      const n = tracks.length;
      if (e.key === "ArrowDown") active = (active + 1) % n;
      else active = active <= 0 ? n - 1 : active - 1;
    } else if (e.key === "Enter" && active >= 0) {
      e.preventDefault();
      const track = tracks[active];
      close();
      void goto(
        resolve("/s/[artist]/[title]", {
          artist: toSlug(track.artist),
          title: toSlug(track.title),
        }),
      );
    }
  }
</script>

<div
  bind:this={root}
  class="relative w-full"
  onfocusout={(e) => {
    if (!root?.contains(e.relatedTarget as Node)) open = false;
  }}
>
  <label
    class={[
      "flex items-center rounded-xl border border-border-strong bg-card transition-colors focus-within:border-cue-text",
      large ? "h-16 gap-4 px-5" : "h-9 gap-2 rounded-lg px-3",
    ]}
  >
    <Search class="shrink-0 text-muted-foreground" size={large ? 22 : 15} />
    <span class="sr-only">Search for a track...</span>
    <input
      class={[
        "min-w-0 flex-1 bg-transparent outline-none placeholder:text-muted-foreground",
        large ? "text-lg sm:text-xl" : "text-sm",
      ]}
      aria-activedescendant={listed && active >= 0
        ? `${id}-${active}`
        : undefined}
      aria-autocomplete="list"
      aria-controls={listed ? `${id}-list` : undefined}
      aria-expanded={listed}
      autocomplete="off"
      onfocus={() => (open = true)}
      oninput={() => (open = true)}
      {onkeydown}
      placeholder="Search for a track..."
      role="combobox"
      spellcheck="false"
      type="search"
      bind:value={query}
    />
    {#if loading}
      <span class="flex shrink-0 text-muted-foreground" aria-hidden="true"
        ><Spinner size={16} /></span
      >
    {/if}
  </label>

  <p class="sr-only" role="status">{status}</p>

  {#if showResults}
    <!-- keeps focus in the input, so a click on a result lands before the popup closes -->
    <div
      class="absolute inset-x-0 top-full z-40 mt-1.5 overflow-hidden popover p-1"
      onmousedown={(e) => e.preventDefault()}
      role="presentation"
    >
      {#if error}
        <p class="px-3 py-2.5 text-sm text-danger">{error}</p>
      {:else if tracks.length}
        <ul id="{id}-list" onpointerleave={() => (active = -1)} role="listbox">
          {#each tracks as track, i (track.isrc)}
            <li id="{id}-{i}" aria-selected={i === active} role="option">
              <TrackRow
                active={i === active}
                onclick={close}
                onpointerenter={() => (active = i)}
                tabindex={-1}
                {track}
              />
            </li>
          {/each}
        </ul>
      {:else}
        <p class="px-3 py-2.5 text-sm text-muted-foreground">No matches.</p>
      {/if}
    </div>
  {/if}
</div>
