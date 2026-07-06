<script lang="ts">
  import { marked, Renderer } from "marked";
  import TabBar from "$components/TabBar.svelte";
  import { API_URL } from "$lib/env";

  const renderer = new Renderer();
  renderer.link = ({ href, text }) =>
    `<a href="${href}" target="_blank" rel="noopener" class="text-ink underline">${text}</a>`;

  interface Guide {
    id: string;
    name: string;
    url: string;
    steps: string[];
    screenshot?: string;
  }

  const guides: Guide[] = [
    {
      id: "youly",
      name: "YouLy+",
      url: "/compat/kpoe",
      steps: [
        "Click the YouLy+ icon in your browser toolbar and open **More Settings**.",
        "Under **Sources**, remove all other providers and add **Custom KPoe Server**.",
        "Set **Custom KPoe Server URL** to the endpoint URL above.",
      ],
    },
    {
      id: "lrcget",
      name: "LRCGET",
      url: "/compat/lrclib",
      steps: [
        "Open LRCGET and click the **three dots** menu, then open **Settings**.",
        "Under **LRCLIB instance**, replace the default URL with the endpoint URL above.",
        "Click **Save**.",
      ],
    },
  ];

  let activeId = $state(guides[0].id);
  const guide = $derived(guides.find((g) => g.id === activeId)!);
</script>

<div class="border border-rule rounded-md overflow-hidden bg-paper-2">
  <TabBar
    tabs={guides.map((g) => ({ id: g.id, label: g.name }))}
    active={activeId}
    onchange={(id) => (activeId = id)}
  />

  <div class="p-5 sm:p-6 flex flex-col gap-6">
    <div>
      <p class="text-xs text-muted uppercase tracking-wide mb-1.5">Endpoint</p>
      <code
        class="block font-mono text-sm text-ink bg-paper border border-rule rounded-md px-4 py-2.5"
      >
        {API_URL}{guide.url}
      </code>
    </div>

    <ol class="flex flex-col gap-3">
      {#each guide.steps as step, i}
        <li class="flex gap-3.5">
          <span
            class="shrink-0 w-5 h-5 rounded-full bg-ink text-paper text-xs font-semibold flex items-center justify-center mt-0.5"
          >
            {i + 1}
          </span>
          <span class="text-muted text-sm leading-relaxed"
            >{@html marked.parseInline(step, { renderer })}</span
          >
        </li>
      {/each}
    </ol>

    {#if guide.screenshot}
      <img
        src={guide.screenshot}
        alt="{guide.name} configuration screenshot"
        class="rounded-md border border-rule w-full"
      />
    {/if}
  </div>
</div>
