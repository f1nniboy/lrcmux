<script lang="ts">
  import Meta from "$lib/Meta.svelte";
  import Link from "$lib/Link.svelte";
  import SearchBar from "$components/SearchBar.svelte";
  import TrendingGrid from "$components/TrendingGrid.svelte";
  import UseItInYourApp from "$components/UseItInYourApp.svelte";
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
      <Wordmark /> searches every public lyrics provider in parallel and returns the
      best available match. No API keys, no sign-up, no costs.
    </p>
  </section>

  <SearchBar />

  <TrendingGrid />

  <hr class="m-0 border-rule" />

  <section>
    <h2 class="text-2xl sm:text-3xl font-semibold text-ink mb-8">
      Why <Wordmark />?
    </h2>
    <div class="grid sm:grid-cols-3 gap-8 sm:gap-10">
      <div>
        <h3 class="text-lg font-semibold text-ink mb-1.5">Better coverage</h3>
        <p class="text-muted text-base sm:text-sm leading-relaxed">
          No single provider has everything. By querying them all in parallel,
          we catch what others miss.
        </p>
      </div>
      <div>
        <h3 class="text-lg font-semibold text-ink mb-1.5">Free and public</h3>
        <p class="text-muted text-base sm:text-sm leading-relaxed">
          Use this public instance or host it yourself, without the need for any
          provider API keys.
        </p>
      </div>
      <div>
        <h3 class="text-lg font-semibold text-ink mb-1.5">
          Drop-in compatible
        </h3>
        <p class="text-muted text-base sm:text-sm leading-relaxed">
          Existing LRCLIB and LyricsPlus clients work by pointing them at the
          matching compatibility endpoint, so you won't miss out on improved
          lyrics.
        </p>
      </div>
    </div>
  </section>

  <hr class="m-0 border-rule" />

  <section>
    <h2 class="text-2xl sm:text-3xl font-semibold text-ink mb-8">Use cases</h2>
    <div class="space-y-10">
      <div>
        <h3 class="text-lg font-semibold text-ink mb-2">
          Drop-in upgrade for existing apps
        </h3>
        <p class="text-muted leading-relaxed">
          Apps with a configurable LRCLIB or KPOE base URL work with <Wordmark
          /> without code changes. Point any LRCLIB-compatible client at
          <code class="text-ink font-mono text-sm"
            >https://api.lrcmux.dev/compat/lrclib</code
          >, or set
          <Link href="https://github.com/ibratabian17/YouLyPlus">YouLy+</Link>
          and other LyricsPlus clients to
          <code class="text-ink font-mono text-sm"
            >https://api.lrcmux.dev/compat/kpoe</code
          >.
        </p>
      </div>

      <div>
        <h3 class="text-lg font-semibold text-ink mb-2">
          Music library tagging
        </h3>
        <p class="text-muted leading-relaxed">
          Tools like <Link href="https://beets.io">beets</Link>
          and
          <Link href="https://picard.musicbrainz.org">MusicBrainz Picard</Link>
          tag your audio files with ISRCs. You can use those ISRCs to fetch synced
          LRC sidecar files from <Wordmark />
          and place them next to your audio, which players pick up automatically.
        </p>
      </div>

      <div>
        <h3 class="text-lg font-semibold text-ink mb-2">
          Media center integration
        </h3>
        <p class="text-muted leading-relaxed">
          Kodi, Plex, and Jellyfin can fetch synced lyrics via plugins. Point
          them at our compatibility endpoints for better coverage than any
          single provider.
        </p>
      </div>

      <div>
        <h3 class="text-lg font-semibold text-ink mb-2">
          Building your own app
        </h3>
        <p class="text-muted leading-relaxed">
          The <code class="text-ink font-mono text-sm">json</code>
          format returns structured line and word timestamps suitable for building
          karaoke-style displays, lyric overlays, or anything your heart desires.
        </p>
      </div>
    </div>
  </section>

  <hr class="m-0 border-rule" />

  <section class="border border-rule rounded-md p-5 sm:p-6 bg-paper-2">
    <h2 class="text-lg font-semibold text-ink mb-2">Wrong lyrics?</h2>
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
