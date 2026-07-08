<script lang="ts">
  import Meta from "$lib/Meta.svelte";
  import Link from "$lib/Link.svelte";
  import SearchBar from "$components/SearchBar.svelte";
  import TrendingGrid from "$components/TrendingGrid.svelte";
  import UseItInYourApp from "$components/UseItInYourApp.svelte";
  import CompatGuides from "$components/CompatGuides.svelte";
  import Wordmark from "$components/Wordmark.svelte";
</script>

<Meta
  description="Free synchronized lyrics for any song. LRC, word-by-word, and plain text, powered by multiple providers. Free lyrics API, no key required."
  og={{ image: "/logo.png" }}
/>

<div
  class="flex flex-col gap-12 sm:gap-16 w-full max-w-3xl min-w-0 px-5 sm:px-8 pt-20 sm:pt-28 pb-16"
>
  <section>
    <h1
      class="text-5xl sm:text-6xl lg:text-7xl font-bold tracking-tight leading-none text-ink mb-6 text-balance"
    >
      Synced lyrics for <span class="text-cue">any song</span>.
    </h1>
    <p class="text-ink/80 text-lg sm:text-xl leading-relaxed max-w-2xl">
      <Wordmark /> is an API that aggregates every public lyrics provider and returns
      the best available match. No API keys, no sign-up, no costs.
    </p>
  </section>

  <SearchBar />

  <TrendingGrid />

  <hr class="m-0 border-rule" />

  <section>
    <h2 class="text-2xl font-semibold text-ink mb-3">Drop-in usage</h2>
    <p class="text-muted leading-relaxed mb-6">
      Various apps allow you to set custom endpoints for fetching lyrics. Point
      them at <Wordmark />
      and get better coverage easily, no code required.
    </p>
    <CompatGuides />
  </section>

  <hr class="m-0 border-rule" />

  <section>
    <h2 class="text-2xl font-semibold text-ink mb-3">Wrong lyrics?</h2>
    <p class="text-muted leading-relaxed">
      <Wordmark /> is purely an aggregator with no lyric database of its own. If a
      song has wrong, missing, or badly synchronized lyrics, the fix needs to happen
      at the source:
      <Link href="https://lrclib.net">LRCLIB</Link>,
      <Link href="https://www.musixmatch.com">Musixmatch</Link>, or
      <Link href="https://genius.com">Genius</Link>. Once a correction is live
      upstream, it will appear in <Wordmark /> after some time.
    </p>
  </section>

  <hr class="m-0 border-rule" />

  <UseItInYourApp />
</div>
