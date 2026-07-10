<script lang="ts">
  import type { LyricsResult } from "$lib/types";
  import type { PageData } from "./+page.server";
  import { toSlug } from "$lib/slug";
  import { API_URL } from "$lib/env";
  import Meta from "$lib/Meta.svelte";
  import Spinner from "$components/Spinner.svelte";
  import UseItInYourApp from "$components/UseItInYourApp.svelte";

  let { data }: { data: PageData } = $props();

  type FetchState =
    | { status: "loading" }
    | { status: "ok"; result: LyricsResult }
    | { status: "error"; code: number; message: string; retryAfter?: number };

  let fetchState = $state<FetchState>({ status: "loading" });

  const lyrics = $derived<LyricsResult | null>(
    data.lyrics ?? (fetchState.status === "ok" ? fetchState.result : null),
  );

  const track = $derived(lyrics?.track);

  const description = $derived(
    lyrics
      ? `View the full lyrics for ${lyrics.track.title} by ${lyrics.track.artist}.`
      : undefined,
  );

  $effect(() => {
    if (data.lyrics) return;

    fetchState = { status: "loading" };
    const controller = new AbortController();
    const p = new URLSearchParams({ artist: data.artist, title: data.title });

    fetch(`${API_URL}/get?${p}`, { signal: controller.signal })
      .then(async (res) => {
        if (res.ok) {
          fetchState = { status: "ok", result: await res.json() };
          return;
        }
        const body = await res.json().catch(() => null);
        fetchState = {
          status: "error",
          code: res.status,
          message: body?.detail ?? `Request failed with status ${res.status}`,
          retryAfter:
            parseInt(res.headers.get("Retry-After") ?? "0", 10) || undefined,
        };
      })
      .catch((e: Error) => {
        if (e.name === "AbortError") return;
        fetchState = { status: "error", code: 0, message: e.message };
      });

    return () => controller.abort();
  });
</script>

<Meta
  title="{data.title} by {data.artist}"
  {description}
  og={data.lyrics
    ? { type: "music.song", image: data.lyrics.track.cover.medium }
    : undefined}
/>

<div
  class="flex flex-col gap-12 sm:gap-16 w-full max-w-3xl min-w-0 px-5 sm:px-8 pt-12 pb-16 flex-1"
>
  {#if lyrics && track}
    <div class="flex flex-col gap-8">
      <header class="flex items-start gap-4">
        {#if track.cover.medium}
          <img
            src={track.cover.medium}
            alt=""
            class="w-20 h-20 sm:w-24 sm:h-24 object-cover rounded-md shadow-md shrink-0"
          />
        {:else}
          <div
            class="w-20 h-20 sm:w-24 sm:h-24 rounded-md bg-paper-2 shrink-0"
          ></div>
        {/if}
        <div class="flex-1 min-w-0 pt-1">
          <h1 class="text-2xl sm:text-3xl font-semibold text-ink leading-tight">
            {track.title}
          </h1>
          <a
            href={`/s/${toSlug(track.artist)}`}
            class="text-muted mt-1 hover:underline no-underline"
          >
            {track.artist}
          </a>
        </div>
      </header>

      <div>
        <div
          class="border border-rule rounded-md bg-paper-2 p-5 text-ink text-base leading-9 whitespace-pre-wrap"
        >
          {lyrics.lines.map((l) => l.text).join("\n")}
        </div>
        {#if lyrics.meta.source}
          <p class="mt-3 text-xs text-muted text-center">
            &copy; {lyrics.meta.source.name}
          </p>
        {/if}
      </div>
    </div>

    <hr class="m-0 border-rule" />

    <UseItInYourApp />
  {:else if fetchState.status === "loading"}
    <div class="flex-1 flex flex-col items-center justify-center gap-5">
      <span class="text-muted"><Spinner size={48} /></span>
      <p class="text-muted text-sm text-center">
        <span class="text-ink font-medium">{data.title}</span> by
        <span class="text-ink font-medium">{data.artist}</span>
      </p>
    </div>
  {:else if fetchState.status === "error"}
    <div class="flex-1 flex flex-col items-center justify-center gap-5">
      <div class="flex flex-col items-center gap-3">
        <p class="text-2xl font-semibold text-ink">
          {fetchState.code === 404
            ? "We couldn't find the lyrics for this song."
            : fetchState.code === 429
              ? "Slow down a bit."
              : "Something went wrong."}
        </p>
        <p class="text-muted text-center">
          {#if fetchState.code === 404}
            Know a source that has it?
          {:else if fetchState.code === 429}
            Try again in <strong>{fetchState.retryAfter ?? "a few"}</strong>
            {fetchState.retryAfter === 1 ? "second" : "seconds"}.
          {:else}
            {fetchState.message}
          {/if}
        </p>
      </div>
      {#if fetchState.code === 404}
        <a
          href={`https://github.com/f1nniboy/lrcmux/issues/new?template=04-new-provider.yml&example=${encodeURIComponent(`${data.artist} - ${data.title}`)}`}
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-2 bg-ink text-paper px-3 py-2 rounded-md font-medium hover:bg-cue transition-colors"
        >
          Request a provider
        </a>
      {/if}
    </div>
  {/if}
</div>
