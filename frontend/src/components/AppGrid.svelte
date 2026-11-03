<script lang="ts">
  import ExternalLink from "./ExternalLink.svelte";
  import { ArrowUpRight, ChevronRight } from "@lucide/svelte";
  import { API_URL } from "$lib/env";
  import type { Guide } from "$lib/guides";
  import CopyButton from "./CopyButton.svelte";
  import Modal from "./Modal.svelte";

  let { guides }: { guides: Guide[] } = $props();

  // kept after closing, so the dialog fades out with its content
  let shown = $state<Guide>();
  let open = $state(false);
</script>

{#snippet row(guide: Guide)}
  <img
    class="size-9 shrink-0 rounded-lg object-cover"
    alt=""
    src={guide.icon}
  />
  <span class="min-w-0 flex-1 truncate font-medium">{guide.name}</span>
  <span class={guide.compat ? "tag" : "tag-cue"}
    >{guide.compat ? "Compatible" : "Native"}</span
  >
{/snippet}

<ul class="grid gap-2.5 sm:grid-cols-2">
  {#each guides as guide (guide.id)}
    <li>
      {#if guide.steps || guide.compat}
        <button
          class="flex w-full items-center gap-3 card-button py-2.5 pr-3 pl-2.5 text-left"
          onclick={() => {
            shown = guide;
            open = true;
          }}
          type="button"
        >
          {@render row(guide)}
          <ChevronRight class="shrink-0 text-muted-foreground" />
        </button>
      {:else}
        <ExternalLink
          class="flex items-center gap-3 card-button py-2.5 pr-3 pl-2.5"
          href={guide.url}
        >
          {@render row(guide)}
          <ArrowUpRight class="shrink-0 text-muted-foreground" />
        </ExternalLink>
      {/if}
    </li>
  {/each}
</ul>

{#if shown}
  {const guide = $derived(shown)}
  <Modal icon={guide.icon} bind:open>
    {#snippet title()}{guide.name}{/snippet}
    {#snippet description()}
      <ExternalLink
        class="inline-flex items-center gap-1 transition-colors hover:text-foreground"
        href={guide.url}
        >{new URL(guide.url).hostname}<ArrowUpRight size={14} /></ExternalLink
      >
    {/snippet}

    {#if guide.compat}
      <div
        class="flex items-center gap-2 rounded-lg border border-border bg-inset py-1 pr-1 pl-3"
      >
        <code class="min-w-0 flex-1 truncate font-mono text-[13px]"
          >{API_URL}{guide.compat}</code
        >
        <CopyButton content="{API_URL}{guide.compat}" />
      </div>
    {/if}

    {#if guide.steps}
      <ol
        class="flex list-decimal flex-col gap-2 pl-5 text-sm marker:text-muted-foreground"
      >
        {#each guide.steps as step, i (i)}
          <!-- eslint-disable-next-line svelte/no-at-html-tags -->
          <li class="pl-1">{@html step}</li>
        {/each}
      </ol>
    {/if}
  </Modal>
{/if}
