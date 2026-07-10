<script lang="ts">
  import Meta from "$lib/Meta.svelte";
  import TabBar from "$components/TabBar.svelte";
  import { API_URL } from "$lib/env";
  import Wordmark from "$components/Wordmark.svelte";

  interface Guide {
    id: string;
    name: string;
    url?: string;
    steps: string[];
  }

  const guides: Guide[] = [
    {
      id: "youly",
      name: "YouLy+",
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
      url: "/compat/lrclib",
      steps: [
        "Open LRCGET and click the <strong>three dots</strong> menu, then open <strong>Settings</strong>.",
        "Under <strong>LRCLIB instance</strong>, replace the default URL with the endpoint URL above.",
        "Click <strong>Save</strong>.",
      ],
    },
    {
      id: "navidrome",
      name: "Navidrome",
      steps: [
        'Install the <a href="https://github.com/J0R6IT0/navidrome-lyrics-plugin" target="_blank" rel="noopener" class="text-ink underline">navidrome-lyrics-plugin</a> if you haven\'t already.',
        "Click your <strong>account icon</strong> in the top right and open <strong>Plugins</strong>.",
        "Click on the <strong>nd-lyrics</strong> plugin, then scroll down to <strong>Lyrics providers</strong>.",
        "Press <strong>+</strong>, click the new slot, and select <strong>lrcmux</strong> as the provider.",
        "Click <strong>Save</strong>.",
      ],
    },
  ];

  let activeId = $state(guides[0].id);
  const guide = $derived(guides.find((g) => g.id === activeId)!);
</script>

<Meta
  description="Point your music app at lrcmux.dev for better lyrics coverage, no code required."
  og={{ image: "/logo.png" }}
  title="Apps"
/>

<div
  class="flex flex-col gap-10 w-full max-w-3xl min-w-0 px-5 sm:px-8 pt-12 pb-16"
>
  <div>
    <h1 class="text-3xl font-bold text-ink mb-2">Apps</h1>
    <p class="text-muted leading-relaxed">
      Various apps let you set a custom endpoint for fetching lyrics. Point them
      at <Wordmark /> and get better coverage, no code required.
    </p>
  </div>

  <div class="border border-rule rounded-md overflow-hidden bg-paper-2">
    <TabBar
      active={activeId}
      onchange={(id) => (activeId = id)}
      tabs={guides.map((g) => ({ id: g.id, label: g.name }))}
    />

    <div class="p-5 sm:p-6 flex flex-col gap-6">
      {#if guide.url}
        <div>
          <p class="text-xs text-muted uppercase tracking-wide mb-1.5">
            Endpoint
          </p>
          <code
            class="block font-mono text-sm text-ink bg-paper border border-rule rounded-md px-4 py-2.5"
          >
            {API_URL}{guide.url}
          </code>
        </div>
      {/if}

      <ol class="flex flex-col gap-3">
        {#each guide.steps as step, i (i)}
          <li class="flex gap-3.5">
            <span
              class="shrink-0 w-5 h-5 rounded-full bg-ink text-paper text-xs font-semibold flex items-center justify-center mt-0.5"
            >
              {i + 1}
            </span>
            <!-- eslint-disable-next-line svelte/no-at-html-tags -->
            <span class="text-muted text-sm leading-relaxed">{@html step}</span>
          </li>
        {/each}
      </ol>
    </div>
  </div>
</div>
