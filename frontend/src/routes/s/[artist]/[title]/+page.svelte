<script lang="ts">
  import type { LyricsResult } from "$lib/types";
  import { retryAfterSeconds } from "$lib/api";
  import { API_URL } from "$lib/env";
  import Meta from "$lib/Meta.svelte";
  import ErrorAlert from "$components/ErrorAlert.svelte";
  import LyricsPanel from "$components/LyricsPanel.svelte";
  import Spinner from "$components/Spinner.svelte";
  import UseItInYourApp from "$components/UseItInYourApp.svelte";
  import type { PageData } from "./+page.server";

  let { data }: { data: PageData } = $props();

  let ogImage = $derived(data.lyrics?.track?.cover?.medium);
  let description = $derived.by(() => {
    if (!data.lyrics) return undefined;

    const { title, artist } = data.lyrics.track;
    const level = data.lyrics.meta.level;
    const qualifier = level === "none" ? "" : `${level}-synced `;

    return `View the full ${qualifier}lyrics for ${title} by ${artist}.`;
  });

  type FetchState =
    | { status: "loading" }
    | { status: "ok"; result: LyricsResult }
    | { status: "error"; code: number; message: string; retryAfter?: number };

  let fetchState: FetchState = $state({ status: "loading" });

  $effect(() => {
    const artist = data.artist;
    const title = data.title;

    if (data.lyrics) return;
    fetchState = { status: "loading" };

    const controller = new AbortController();
    const p = new URLSearchParams({ artist, title, format: "json" });

    fetch(`${API_URL}/get?${p}`, { signal: controller.signal })
      .then(async (res) => {
        if (res.ok) {
          fetchState = { status: "ok", result: await res.json() };
          return;
        }

        const body = await res.text().catch(() => "");
        let message = body;

        try {
          const json = JSON.parse(body);
          if (json.detail) message = String(json.detail).slice(0, 500);
        } catch {}
        fetchState = {
          status: "error",
          code: res.status,
          message: message || `Request failed (${res.status})`,
          retryAfter: retryAfterSeconds(res),
        };
      })
      .catch((e) => {
        if ((e as Error).name === "AbortError") return;
        fetchState = {
          status: "error",
          code: 0,
          message: (e as Error).message,
        };
      });

    return () => controller.abort();
  });

  function formatRetryTime(seconds: number): string {
    return new Date(Date.now() + seconds * 1000).toLocaleTimeString([], {
      hour: "numeric",
      minute: "2-digit",
    });
  }
</script>

<Meta
  title="{data.title} by {data.artist}"
  {description}
  og={{ type: "music.song", image: ogImage }}
/>

<div class="w-full max-w-3xl min-w-0 px-5 sm:px-8 pt-12 pb-16">
  {#if data.lyrics}
    <div class="flex flex-col gap-12">
      <div>
        <LyricsPanel track={data.lyrics.track} result={data.lyrics} />
        {#if data.lyrics.meta.source}
          <p class="mt-3 text-xs text-muted text-center">
            &copy; {data.lyrics.meta.source.name}
          </p>
        {/if}
      </div>
      <hr class="m-0 border-rule" />
      <UseItInYourApp />
    </div>
  {:else if fetchState.status === "loading"}
    <div class="flex flex-col items-center justify-center py-32 gap-5">
      <span class="text-muted"><Spinner size={48} /></span>
      <p class="text-muted text-sm text-center">
        <span class="text-ink font-medium">{data.title}</span> by
        <span class="text-ink font-medium">{data.artist}</span>
      </p>
    </div>
  {:else if fetchState.status === "error"}
    {#if fetchState.code === 404}
      <ErrorAlert heading="No lyrics found." message={fetchState.message} />
    {:else if fetchState.code === 429}
      <ErrorAlert
        heading="Slow down a bit."
        message={fetchState.retryAfter
          ? `Try again at ${formatRetryTime(fetchState.retryAfter)}.`
          : fetchState.message}
      />
    {:else}
      <ErrorAlert message={fetchState.message} />
    {/if}
  {:else}
    <div class="flex flex-col gap-12">
      <div>
        <LyricsPanel
          track={fetchState.result.track}
          result={fetchState.result}
        />
        {#if fetchState.result.meta.source}
          <p class="mt-3 text-xs text-muted text-center">
            &copy; {fetchState.result.meta.source.name}
          </p>
        {/if}
      </div>
      <hr class="m-0 border-rule" />
      <UseItInYourApp />
    </div>
  {/if}
</div>
