<script lang="ts" module>
  import type { LyricsResult } from "$lib/types";

  type Outcome =
    | { ok: true; result: LyricsResult }
    | { ok: false; code: number; message: string; retryAfter?: number };

  async function fetchLyrics(url: string): Promise<Outcome> {
    try {
      const res = await fetch(url);
      if (res.ok) return { ok: true, result: await res.json() };
      const body = await res.json().catch(() => null);
      return {
        ok: false,
        code: res.status,
        message: body?.detail ?? `Request failed with status ${res.status}`,
        retryAfter: Number(res.headers.get("Retry-After")) || undefined,
      };
    } catch (e) {
      return { ok: false, code: 0, message: (e as Error).message };
    }
  }
</script>

<script lang="ts">
  import { browser } from "$app/environment";
  import ExternalLink from "$components/ExternalLink.svelte";
  import { resolve } from "$app/paths";
  import { toSlug } from "$lib/slug";
  import { API_URL } from "$lib/env";
  import { fmtDuration } from "$lib/utils";
  import {
    ArrowUpRight,
    AudioLines,
    Clock,
    Disc3,
    Radio,
  } from "@lucide/svelte";
  import Meta from "$lib/Meta.svelte";
  import Cover from "$components/Cover.svelte";
  import DownloadMenu from "$components/DownloadMenu.svelte";
  import Lyrics from "$components/Lyrics.svelte";
  import Spinner from "$components/Spinner.svelte";

  let { data } = $props();

  const pending = $derived(
    browser
      ? fetchLyrics(
          `${API_URL}/get?${new URLSearchParams({ artist: data.artist, title: data.title })}`,
        )
      : undefined,
  );
</script>

<Meta title="{data.title} by {data.artist}" />

{#if pending}
  {#await pending}
    {@render loading()}
  {:then outcome}
    {#if outcome.ok}
      {@render song(outcome.result)}
    {:else}
      {@render failure(outcome)}
    {/if}
  {/await}
{:else}
  {@render loading()}
{/if}

{#snippet song(lyrics: LyricsResult)}
  {const track = lyrics.track}
  <div class="page gap-10 pt-10 pb-24 sm:pt-14">
    <header class="flex items-center gap-5 sm:items-start sm:gap-6">
      <Cover
        class="size-28 sm:size-44"
        src={track.cover.big ?? track.cover.medium}
      />
      <div class="flex min-w-0 flex-col gap-4">
        <div class="flex flex-col gap-1">
          <h1
            class="text-2xl/tight font-semibold tracking-tight text-balance sm:text-3xl/tight"
          >
            {track.title}
          </h1>
          <a
            class="self-start link"
            href={resolve("/s/[artist]", { artist: toSlug(track.artist) })}
            >{track.artist}</a
          >
        </div>
        <ul class="flex flex-wrap gap-x-5 gap-y-1.5 text-sm">
          {#if track.album}
            <li class="flex min-w-0 items-center gap-2">
              <Disc3 class="shrink-0 text-muted-foreground" /><span
                class="truncate">{track.album}</span
              >
            </li>
          {/if}
          <li class="flex items-center gap-2">
            <Clock class="shrink-0 text-muted-foreground" /><span
              class="tabular-nums">{fmtDuration(track.duration)}</span
            >
          </li>
          {#if lyrics.meta.source}
            <li class="flex items-center gap-2">
              <Radio class="shrink-0 text-muted-foreground" />
              {#if lyrics.meta.source.url}
                <ExternalLink class="link" href={lyrics.meta.source.url}
                  >{lyrics.meta.source.name}</ExternalLink
                >
              {:else}
                {lyrics.meta.source.name}
              {/if}
            </li>
          {/if}
        </ul>
        {#if !lyrics.meta.instrumental}
          <DownloadMenu
            artist={data.artist}
            level={lyrics.meta.level}
            title={data.title}
          />
        {/if}
      </div>
    </header>

    <section class="surface px-4 py-5 sm:p-8">
      {#if lyrics.meta.instrumental}
        <div class="flex flex-col items-center gap-3 py-10 text-center">
          <span
            class="grid size-12 place-items-center rounded-full bg-muted text-muted-foreground"
            ><AudioLines size={22} /></span
          >
          <p class="text-sm text-muted-foreground">This track has no lyrics.</p>
        </div>
      {:else}
        <Lyrics lines={lyrics.lines} synced={lyrics.meta.level !== "none"} />
      {/if}
    </section>
  </div>
{/snippet}

{#snippet loading()}
  <div class="page flex-1 items-center justify-center gap-4 py-24 text-center">
    <span class="text-muted-foreground"><Spinner size={32} /></span>
    <p class="text-sm text-muted-foreground">
      <span class="font-medium text-foreground">{data.title}</span> by
      <span class="font-medium text-foreground">{data.artist}</span>
    </p>
  </div>
{/snippet}

{#snippet failure(f: Extract<Outcome, { ok: false }>)}
  <div class="page flex-1 items-center justify-center gap-6 py-24 text-center">
    <div class="flex flex-col gap-1.5">
      <h1 class="section-title">
        {f.code === 404
          ? "We couldn't find the lyrics for this song."
          : f.code === 429
            ? "Slow down a bit."
            : "Something went wrong."}
      </h1>
      <p class="text-muted-foreground">
        {#if f.code === 404}
          Know a source that has it?
        {:else if f.code === 429}
          Try again in <strong>{f.retryAfter ?? "a few"}</strong>
          {f.retryAfter === 1 ? "second" : "seconds"}.
        {:else}
          {f.message}
        {/if}
      </p>
    </div>
    {#if f.code === 404}
      <ExternalLink
        class="button-cue"
        href="https://github.com/f1nniboy/lrcmux/issues/new?template=04-new-provider.yml&example={encodeURIComponent(
          `${data.artist} - ${data.title}`,
        )}"
      >
        Request a provider<ArrowUpRight />
      </ExternalLink>
    {/if}
  </div>
{/snippet}
