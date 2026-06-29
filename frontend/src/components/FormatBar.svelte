<script lang="ts">
  import type { Snippet } from "svelte";
  import type { LyricsFormat, SyncLevel } from "$lib/types";
  import { FORMATS, LEVELS } from "$lib/types";

  interface Props {
    format: LyricsFormat;
    level: SyncLevel;
    onformat: (f: LyricsFormat) => void;
    onlevel: (l: SyncLevel) => void;
    maxLevel?: SyncLevel;
    children?: Snippet;
  }
  let { format, level, onformat, onlevel, maxLevel, children }: Props =
    $props();

  const isRich = $derived(FORMATS.find((f) => f.id === format)?.rich ?? false);
  const rank = { word: 2, line: 1, none: 0 } as const;
  const levelRank = $derived(maxLevel ? rank[maxLevel] : 2);
  const formatMinLevel = $derived(
    FORMATS.find((f) => f.id === format)?.minLevel,
  );
  const disabledLevel = $derived((l: SyncLevel) => {
    if (rank[l] > levelRank) return true;
    if (formatMinLevel && rank[l] < rank[formatMinLevel]) return true;
    return false;
  });
  const disabledFormat = $derived((f: LyricsFormat) => {
    const def = FORMATS.find((x) => x.id === f);
    return def?.minLevel ? rank[def.minLevel] > levelRank : false;
  });
</script>

<header
  class="flex items-center gap-1 border-b border-rule px-2 py-1.5 overflow-x-auto"
>
  {#each FORMATS as f (f.id)}
    {@const disabled = disabledFormat(f.id)}
    <button
      type="button"
      onclick={() => onformat(f.id)}
      {disabled}
      class="px-3 py-1.5 text-sm font-medium rounded transition-colors
                {disabled
        ? 'text-muted opacity-35 cursor-not-allowed'
        : format === f.id
          ? 'bg-ink text-paper cursor-pointer'
          : 'text-muted hover:text-ink hover:bg-paper cursor-pointer'}"
    >
      {f.id.toUpperCase()}
    </button>
  {/each}
</header>

<div class="flex items-center gap-2 px-3 py-2 border-b border-rule">
  {#if !isRich}
    <span class="text-xs font-semibold text-ink">Sync</span>
    {#each LEVELS as l}
      {@const disabled = disabledLevel(l)}
      <button
        type="button"
        onclick={() => onlevel(l)}
        {disabled}
        class="px-2 py-1 text-xs font-medium rounded transition-colors
                    {disabled
          ? 'text-muted opacity-35 cursor-not-allowed'
          : level === l
            ? 'bg-ink/10 text-ink cursor-pointer'
            : 'text-muted hover:text-ink hover:bg-paper cursor-pointer'}"
      >
        {l.charAt(0).toUpperCase() + l.slice(1)}
      </button>
    {/each}
  {/if}
  {#if children}
    {@render children()}
  {/if}
</div>
