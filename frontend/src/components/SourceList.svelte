<script lang="ts">
  import ExternalLink from "./ExternalLink.svelte";
  import { ArrowUpRight } from "@lucide/svelte";
  import type { Source, SyncLevel } from "$lib/types";

  let { sources }: { sources: Source[] } = $props();

  const LEVEL: Record<SyncLevel, string> = {
    word: "Word synced",
    line: "Line synced",
    none: "Plain text",
  };
</script>

<ul class="grid gap-2.5 sm:grid-cols-2">
  {#each sources as source (source.id)}
    <li>
      <ExternalLink
        class="flex items-center gap-3 card-button py-2.5 pr-3 pl-3.5"
        href={source.url}
      >
        <span class="min-w-0 flex-1 truncate font-medium">{source.name}</span>
        <span class={source.level === "word" ? "tag-cue" : "tag"}
          >{LEVEL[source.level]}</span
        >
        <ArrowUpRight class="shrink-0 text-muted-foreground" />
      </ExternalLink>
    </li>
  {/each}
</ul>
