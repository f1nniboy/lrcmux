<script lang="ts" module>
  export interface Guide {
    id: string;
    name: string;
    desc: string;
    icon: string;
    url?: string;
    compat?: string;
    steps?: string[];
  }
</script>

<script lang="ts">
  import { slide } from "svelte/transition";
  import { API_URL } from "$lib/env";

  let { guides }: { guides: Guide[] } = $props();

  let expanded = $state<string | null>(null);

  function toggle(id: string) {
    expanded = expanded === id ? null : id;
  }

  function isExpandable(guide: Guide) {
    return (guide.steps && guide.steps.length > 0) || !!guide.compat;
  }
</script>

<div
  class="flex flex-col divide-y divide-rule border border-rule rounded-md overflow-hidden"
>
  {#snippet content(guide: Guide)}
    <img
      class="shrink-0 w-8 h-8 rounded-md object-cover"
      alt=""
      src={guide.icon}
    />

    <div class="flex-1 min-w-0">
      {#if guide.url}
        <!-- eslint-disable svelte/no-navigation-without-resolve -->
        <a
          class="font-medium text-ink text-sm hover:text-cue transition-colors inline-flex items-center gap-1.5"
          href={guide.url}
          onclick={(e) => e.stopPropagation()}
          rel="noopener noreferrer"
          target="_blank"
        >
          <!-- eslint-enable svelte/no-navigation-without-resolve -->
          {guide.name}
          <svg
            class="w-3 h-3 opacity-50"
            fill="none"
            stroke="currentColor"
            stroke-width="4"
            viewBox="0 0 24 24"
          >
            <path
              d="M7 17L17 7M17 7H7M17 7v10"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </a>
      {:else}
        <span class="font-medium text-ink text-sm">{guide.name}</span>
      {/if}
      <p class="text-muted text-xs mt-0.5 leading-relaxed">
        {guide.desc}
      </p>
    </div>
  {/snippet}

  {#each guides as guide (guide.id)}
    <div>
      {#if isExpandable(guide)}
        <div
          class="flex items-center gap-3 px-4 py-3.5 transition-colors cursor-pointer hover:bg-paper-2"
          class:bg-paper={expanded !== guide.id}
          class:bg-paper-2={expanded === guide.id}
          onclick={() => toggle(guide.id)}
          onkeydown={(e) => e.key === "Enter" && toggle(guide.id)}
          role="button"
          tabindex="0"
        >
          {@render content(guide)}
          <svg
            class="w-4 h-4 text-muted transition-transform shrink-0"
            class:rotate-180={expanded === guide.id}
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            viewBox="0 0 24 24"
          >
            <path
              d="M19 9l-7 7-7-7"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </div>
      {:else}
        <div class="flex items-center gap-3 px-4 py-3.5 bg-paper">
          {@render content(guide)}
        </div>
      {/if}

      {#if expanded === guide.id}
        <div
          class="px-4 py-5 flex flex-col gap-5 bg-paper-2"
          transition:slide={{ duration: 150 }}
        >
          {#if guide.compat}
            <div>
              <p class="text-xs text-muted uppercase tracking-wide mb-1.5">
                Endpoint
              </p>
              <code
                class="block font-mono text-sm text-ink bg-paper border border-rule rounded-md px-4 py-2.5"
              >
                {API_URL}{guide.compat}
              </code>
            </div>
          {/if}

          {#if guide.steps && guide.steps.length > 0}
            <ol class="flex flex-col gap-3">
              {#each guide.steps as step, i (i)}
                <li class="flex gap-3.5">
                  <span
                    class="shrink-0 w-5 h-5 rounded-full bg-ink text-paper text-xs font-semibold flex items-center justify-center mt-0.5"
                  >
                    {i + 1}
                  </span>
                  <span class="text-muted text-sm leading-relaxed">
                    <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                    {@html step}
                  </span>
                </li>
              {/each}
            </ol>
          {/if}
        </div>
      {/if}
    </div>
  {/each}
</div>
