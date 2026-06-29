<script lang="ts">
  import { page } from "$app/state";
  import LyricsPanel from "$components/LyricsPanel.svelte";
  import Spinner from "$components/Spinner.svelte";
  import UseItInYourApp from "$components/UseItInYourApp.svelte";
  import { getLyricsJSON, LyricsError } from "$lib/api";
  import { searchDeezer } from "$lib/deezer";
  import type { DeezerTrack, LyricsResult } from "$lib/types";

  let artist = $derived(decodeURIComponent(page.params.artist ?? ""));
  let title = $derived(decodeURIComponent(page.params.title ?? ""));

  let track = $state<DeezerTrack | null>(null);
  let result = $state<LyricsResult | null>(null);
  let loading = $state(true);
  let error = $state<{ status?: number; message: string } | null>(null);

  function syntheticTrack(): DeezerTrack {
    return {
      id: 0,
      title,
      title_short: title,
      duration: 0,
      artist: { id: 0, name: artist },
      album: { id: 0, title: "" },
    } as DeezerTrack;
  }

  $effect(() => {
    void load();
  });

  async function load() {
    loading = true;
    error = null;
    track = null;
    result = null;
    try {
      const tracks = await searchDeezer(`${artist} ${title}`, {
        limit: 5,
      });
      track =
        tracks.find(
          (t) =>
            t.artist.name.toLowerCase() === artist.toLowerCase() &&
            t.title.toLowerCase() === title.toLowerCase(),
        ) ??
        tracks[0] ??
        syntheticTrack();
    } catch {
      track = syntheticTrack();
    }

    try {
      const r = await getLyricsJSON(track!);
      result = r;
    } catch (e) {
      if (e instanceof LyricsError) {
        error = { status: e.status, message: e.message };
      } else {
        error = { message: (e as Error).message };
      }
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>{title} by {artist} - {page.url.hostname}</title>
</svelte:head>

<div class="w-full max-w-3xl min-w-0 px-5 sm:px-8 pt-12 pb-16">
  {#if loading}
    <div class="flex flex-col items-center justify-center py-32 gap-5">
      <span class="text-muted"><Spinner size={48} /></span>
      <p class="text-muted text-sm text-center">
        <span class="text-ink font-medium">{title}</span> by
        <span class="text-ink font-medium">{artist}</span>
      </p>
    </div>
  {:else if error}
    <div class="flex flex-col items-center justify-center py-16">
      <h1 class="text-2xl sm:text-3xl font-semibold text-ink mb-2 text-center">
        {error?.status === 404 ? "No lyrics found." : "Something went wrong."}
      </h1>
      <p class="text-muted text-center">{error.message}</p>
    </div>
  {/if}

  {#if track && result}
    <div class="flex flex-col gap-12">
      <div>
        <LyricsPanel {track} {result} />

        {#if result.meta.source}
          <p class="mt-3 text-xs text-muted text-center">
            &copy; {result.meta.source.name}
          </p>
        {/if}
      </div>

      <hr class="m-0 border-rule" />

      <UseItInYourApp />
    </div>
  {/if}
</div>
