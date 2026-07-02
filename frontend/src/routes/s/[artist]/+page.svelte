<script lang="ts">
  import { getArtistTopTracks, deezerToAPITrack } from "$lib/deezer";
  import type { Track } from "$lib/types";
  import Meta from "$lib/Meta.svelte";
  import TrackItem from "$components/TrackItem.svelte";
  import Spinner from "$components/Spinner.svelte";
  import ErrorAlert from "$components/ErrorAlert.svelte";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();

  let tracks = $state<Track[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  $effect(() => {
    const artistId = data.artist.id;
    loading = true;
    tracks = [];
    error = null;

    void getArtistTopTracks({}, artistId)
      .then((deezerTracks) => {
        tracks = deezerTracks.map((dt) => ({
          ...deezerToAPITrack(dt),
          isrc: dt.isrc || String(dt.id),
        }));
      })
      .catch((e: Error) => {
        error = e.message;
      })
      .finally(() => {
        loading = false;
      });
  });
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

    {#if loading}
      <div class="flex justify-center py-12 text-muted">
        <Spinner size={48} />
      </div>
    {:else if error}
      <ErrorAlert message={error} />
    {:else if tracks.length}
      <div class="divide-y divide-rule -mx-3">
        {#each tracks as track (track.isrc)}
          <TrackItem {track} variant="row" />
        {/each}
      </div>
    {:else}
      <p class="text-muted text-sm">No top tracks available.</p>
    {/if}
  </div>
</div>
