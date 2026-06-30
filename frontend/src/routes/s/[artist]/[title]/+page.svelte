<script lang="ts">
  import { page } from "$app/state";
  import LyricsPanel from "$components/LyricsPanel.svelte";
  import Spinner from "$components/Spinner.svelte";
  import UseItInYourApp from "$components/UseItInYourApp.svelte";
  import { getLyricsJSON, LyricsError } from "$lib/api";
  import type { LyricsResult } from "$lib/types";

  let artist = $derived(decodeURIComponent(page.params.artist ?? ""));
  let title = $derived(decodeURIComponent(page.params.title ?? ""));

  let result = $state<LyricsResult | null>(null);
  let loading = $state(true);
  let error = $state<{
    status?: number;
    message: string;
    retryAfter?: number;
  } | null>(null);

  function formatRetryTime(seconds: number): string {
    return new Date(Date.now() + seconds * 1000).toLocaleTimeString([], {
      hour: "numeric",
      minute: "2-digit",
    });
  }

  $effect(() => {
    load();
  });

  async function load() {
    loading = true;
    error = null;
    result = null;

    try {
      result = await getLyricsJSON(artist, title);
    } catch (e) {
      if (e instanceof LyricsError) {
        error = {
          status: e.status,
          message: e.message,
          retryAfter: e.retryAfter,
        };
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
        {#if error?.status === 404}
          No lyrics found.
        {:else if error?.status === 429}
          Slow down a bit.
        {:else}
          Something went wrong.
        {/if}
      </h1>
      <p class="text-muted text-center">
        {#if error?.status === 429 && error.retryAfter}
          Try again at <strong class="text-ink"
            >{formatRetryTime(error.retryAfter)}</strong
          >.
        {:else}
          {error.message}
        {/if}
      </p>
    </div>
  {/if}

  {#if result}
    <div class="flex flex-col gap-12">
      <div>
        <LyricsPanel track={result.track} {result} />

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
