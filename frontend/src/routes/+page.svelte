<script lang="ts">
  import ExternalLink from "$components/ExternalLink.svelte";
  import { ArrowRight } from "@lucide/svelte";
  import { resolve } from "$app/paths";
  import Meta from "$lib/Meta.svelte";
  import { guides } from "$lib/guides";
  import SearchBar from "$components/SearchBar.svelte";
  import TrackRow from "$components/TrackRow.svelte";
  import AppGrid from "$components/AppGrid.svelte";
  import SourceList from "$components/SourceList.svelte";
  import { siGithub, siMatrix } from "simple-icons";
  import BrandIcon from "$components/BrandIcon.svelte";
  import Wordmark from "$components/Wordmark.svelte";

  let { data } = $props();
</script>

<Meta
  description="Free synchronized lyrics for any song. LRC, word-by-word, and plain text, powered by multiple providers. Free lyrics API, no key required."
  index
  og={{ image: "/logo.png" }}
/>

<div class="page gap-16 pt-16 pb-24 sm:gap-20 sm:pt-24">
  <section class="flex flex-col gap-6">
    <h1
      class="text-5xl/[0.95] font-bold tracking-tight text-balance sm:text-7xl/[0.95]"
    >
      Synced lyrics<br />for <span class="text-cue">any song</span>.
    </h1>
    <p class="max-w-2xl text-lg text-pretty text-muted-foreground sm:text-xl">
      <Wordmark class="text-foreground" /> is a free lyrics API that searches every
      public provider and returns the best available match. No keys, no sign-up, and
      no costs.
    </p>
    <a class="mt-2 button-cue self-start button-lg" href={resolve("/docs")}
      >Read the docs<ArrowRight /></a
    >
    <div class="mt-6 flex flex-col gap-2.5">
      <p class="text-sm text-muted-foreground">Or try it out</p>
      <SearchBar large />
      {#if data.chart.length}
        <ul class="-mx-2.5 grid grid-cols-2 sm:grid-cols-4">
          {#each data.chart as track (track.isrc)}
            <li>
              <TrackRow duration={false} {track} />
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </section>

  {#if data.sources.length}
    <section class="flex flex-col gap-6">
      <div class="flex flex-col gap-2">
        <h2 class="section-title">Sources</h2>
        <p class="text-muted-foreground">
          Each request asks these providers and gives you the most high-quality
          result.
        </p>
      </div>

      <SourceList sources={data.sources} />
    </section>
  {/if}

  <section class="flex flex-col gap-6">
    <div class="flex flex-col gap-2">
      <h2 class="section-title">Works with</h2>
      <p class="text-muted-foreground">Pick an app to see how to set it up.</p>
    </div>
    <AppGrid {guides} />
  </section>

  <section class="flex flex-col gap-6 surface p-6 sm:p-8">
    <div class="flex flex-col gap-2">
      <h2 class="section-title">Have questions?</h2>
      <p class="text-muted-foreground">
        Ask in the chat, or open an issue if something's broken.
      </p>
    </div>
    <div class="flex flex-col gap-2 sm:flex-row">
      <ExternalLink
        class="button button-lg"
        href="https://matrix.to/#/#lrcmux:oss.zone"
        ><BrandIcon icon={siMatrix} />Join the chat</ExternalLink
      >
      <ExternalLink
        class="button button-lg"
        href="https://github.com/f1nniboy/lrcmux/issues"
        ><BrandIcon icon={siGithub} />Open an issue</ExternalLink
      >
    </div>
  </section>
</div>
