<script lang="ts">
  import type { Line } from "$lib/types";
  import { fmtDuration } from "$lib/utils";

  let { lines, synced }: { lines: Line[]; synced: boolean } = $props();

  // the timestamp column fits m:ss, or mm:ss once a song passes ten minutes
  const long = $derived(lines.some((l) => (l.start ?? 0) >= 600_000));
</script>

<!-- the timestamps are generated content, so copying the lyrics leaves them out, and the
     empty alt text keeps screen readers from reading them -->
<ol class="flex flex-col gap-3 text-lg/snug sm:text-xl/snug">
  {#each lines as line, i (i)}
    {#if line.text.trim()}
      <li
        class={[
          synced &&
            "grid grid-cols-[auto_minmax(0,1fr)] gap-3 before:pt-1.5 before:text-right before:font-mono before:text-xs before:text-muted-foreground before:tabular-nums before:content-[attr(data-time)/'']",
          synced && (long ? "before:w-[5ch]" : "before:w-[4ch]"),
        ]}
        data-time={synced
          ? fmtDuration(Math.floor((line.start ?? 0) / 1000))
          : undefined}
      >
        <span>{line.text}</span>
      </li>
    {:else}
      <li class="h-4" aria-hidden="true"></li>
    {/if}
  {/each}
</ol>
