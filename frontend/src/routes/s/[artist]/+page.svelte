<script lang="ts">
  import Meta from "$lib/Meta.svelte";
  import TrackItem from "$components/TrackItem.svelte";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();
</script>

<Meta
  title={data.artist.name}
  description="View lyrics for all tracks by {data.artist.name}."
  og={{ image: data.artist.picture_medium }}
/>

<div class="w-full max-w-3xl min-w-0 px-5 sm:px-8 pt-12 pb-16">
  <div class="flex flex-col gap-8">
    <header class="flex items-center gap-5">
      {#if data.artist.picture_medium}
        <img
          src={data.artist.picture_medium}
          alt=""
          class="w-20 h-20 sm:w-24 sm:h-24 rounded-full object-cover shadow-md shrink-0"
        />
      {:else}
        <div
          class="w-20 h-20 sm:w-24 sm:h-24 rounded-full bg-paper-2 shrink-0"
        ></div>
      {/if}
      <h1 class="text-3xl sm:text-4xl font-bold text-ink leading-tight">
        {data.artist.name}
      </h1>
    </header>
    <div class="divide-y divide-rule -mx-3">
      {#each data.tracks as track (track.isrc)}
        <TrackItem {track} variant="row" />
      {/each}
    </div>
  </div>
</div>
