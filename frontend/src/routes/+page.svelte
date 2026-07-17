<script lang="ts">
  import Meta from "$lib/Meta.svelte";
  import SearchBar from "$components/SearchBar.svelte";
  import TrendingGrid from "$components/TrendingGrid.svelte";
  import Wordmark from "$components/Wordmark.svelte";
  import GuideList from "$components/GuideList.svelte";
  import type { Guide } from "$components/GuideList.svelte";
  import UseItInYourApp from "$components/UseItInYourApp.svelte";

  let { data } = $props();

  const apps: Guide[] = [
    {
      id: "beets",
      name: "Beets",
      link: "https://beets.io",
      steps: [
        "Add <strong>lyrics</strong> to the <strong>plugins</strong> list in your configuration.",
        "Run <strong>beet lyrics</strong> to fetch synced lyrics for your library.",
      ],
    },
    {
      id: "navidrome",
      name: "Navidrome",
      link: "https://www.navidrome.org",
      steps: [
        'Install the <a href="https://github.com/J0R6IT0/navidrome-lyrics-plugin" target="_blank" rel="noopener" class="hover:text-ink underline">navidrome-lyrics-plugin</a> if you haven\'t already.',
        "Click your <strong>account icon</strong> in the top right and open <strong>Plugins</strong>.",
        "Click on the <strong>nd-lyrics</strong> plugin, then scroll down to <strong>Lyrics providers</strong>.",
        "Press <strong>+</strong>, click the new slot, and select <strong>lrcmux</strong> as the provider.",
        "Click <strong>Save</strong>.",
      ],
    },
  ];

  const compat: Guide[] = [
    {
      id: "youly",
      name: "YouLy+",
      link: "https://github.com/ibratabian17/YouLyPlus",
      url: "/compat/kpoe",
      steps: [
        "Click the YouLy+ icon in your browser toolbar and open <strong>More Settings</strong>.",
        "Under <strong>Sources</strong>, remove all other providers and add <strong>Custom KPoe Server</strong>.",
        "Set <strong>Custom KPoe Server URL</strong> to the endpoint URL above.",
      ],
    },
    {
      id: "lrcget",
      name: "LRCGET",
      link: "https://github.com/tranxuanthang/lrcget",
      url: "/compat/lrclib",
      steps: [
        "Open LRCGET and click the <strong>three dots</strong> menu, then open <strong>Settings</strong>.",
        "Under <strong>LRCLIB instance</strong>, replace the default URL with the endpoint URL above.",
        "Click <strong>Save</strong>.",
      ],
    },
  ];
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
      <Wordmark /> searches every public lyrics provider and returns the best available
      match. No API keys, no sign-up, no costs.
    </p>
  </section>

  <SearchBar />

  <TrendingGrid tracks={data.trending} />

  <hr class="m-0 border-rule" />

  <section id="apps" class="flex flex-col gap-10">
    <div class="flex flex-col gap-4">
      <div>
        <h2 class="text-2xl font-semibold text-ink mb-3">Apps</h2>
        <p class="text-muted leading-relaxed">
          A growing number of apps support <Wordmark /> natively. Enable the corresponding
          integration and you're set.
        </p>
      </div>
      <GuideList guides={apps} />
    </div>

    <div class="flex flex-col gap-4">
      <div>
        <h2 class="text-2xl font-semibold text-ink mb-3">Compatibility</h2>
        <p class="text-muted leading-relaxed">
          Some apps support compatible APIs that <Wordmark /> can speak. Point them
          at the right endpoint and they'll work without native support.
        </p>
      </div>
      <GuideList guides={compat} />
    </div>
  </section>

  <hr class="m-0 border-rule" />

  <UseItInYourApp />
</div>
