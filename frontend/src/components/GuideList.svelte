<script lang="ts" module>
  export interface Guide {
    id: string;
    name: string;
    link?: string;
    url?: string;
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
</script>

<div
  class="flex flex-col divide-y divide-rule border border-rule rounded-md overflow-hidden"
>
  {#each guides as guide (guide.id)}
    <div>
      <div
        class="flex items-center justify-between px-4 py-3.5 hover:bg-paper-2 transition-colors cursor-pointer"
        class:bg-paper={expanded !== guide.id}
        class:bg-paper-2={expanded === guide.id}
        onclick={() => toggle(guide.id)}
        onkeydown={(e) => e.key === "Enter" && toggle(guide.id)}
        role="button"
        tabindex="0"
      >
        {#if guide.link}
          <!-- eslint-disable svelte/no-navigation-without-resolve -->
          <a
            class="font-medium text-ink text-sm hover:text-cue transition-colors flex items-center gap-1.5"
            href={guide.link}
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
              stroke-width="2"
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
        <svg
          class="w-4 h-4 text-muted transition-transform"
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

      {#if expanded === guide.id}
        <div
          class="px-4 py-5 flex flex-col gap-5 bg-paper-2"
          transition:slide={{ duration: 150 }}
        >
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
          {:else}
            <p class="text-muted text-center">No additional setup required.</p>
          {/if}
        </div>
      {/if}
    </div>
  {/each}
</div>
