<script lang="ts">
  import Meta from "$lib/Meta.svelte";
  import Cover from "$components/Cover.svelte";
  import TrackRow from "$components/TrackRow.svelte";

  let { data } = $props();
</script>

<Meta
  description="View lyrics for all tracks by {data.artist.name}."
  og={{ image: data.artist.picture_medium }}
  title={data.artist.name}
/>

<div class="page gap-10 pt-10 pb-24 sm:pt-14">
  <header class="flex items-center gap-5 sm:gap-6">
    <Cover
      class="size-28 sm:size-44"
      round
      src={data.artist.picture_big ?? data.artist.picture_medium}
    />
    <h1
      class="text-3xl/tight font-semibold tracking-tight text-balance sm:text-4xl/tight"
    >
      {data.artist.name}
    </h1>
  </header>

  <section>
    {#if data.tracks.length > 0}
      <ul class="-mx-2.5">
        {#each data.tracks as track (track.isrc)}
          <li>
            <TrackRow artist={track.artist !== data.artist.name} {track} />
          </li>
        {/each}
      </ul>
    {:else}
      <p class="text-muted-foreground">No top tracks available.</p>
    {/if}
  </section>
</div>
