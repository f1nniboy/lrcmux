<script lang="ts">
  import type { Snippet } from "svelte";
  import type { LyricsFormat, SyncLevel } from "$lib/types";
  import { FORMATS, LEVELS, LEVEL_RANK } from "$lib/types";
  import TabBar from "$components/TabBar.svelte";

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
  const levelRank = $derived(maxLevel ? LEVEL_RANK[maxLevel] : 2);
  const formatMinLevel = $derived(
    FORMATS.find((f) => f.id === format)?.minLevel,
  );
  const disabledLevel = $derived((l: SyncLevel) => {
    if (LEVEL_RANK[l] > levelRank) return true;
    if (formatMinLevel && LEVEL_RANK[l] < LEVEL_RANK[formatMinLevel])
      return true;
    return false;
  });
  const disabledFormat = $derived((f: LyricsFormat) => {
    const def = FORMATS.find((x) => x.id === f);
    return def?.minLevel ? LEVEL_RANK[def.minLevel] > levelRank : false;
  });
</script>

<TabBar
  tabs={FORMATS.map((f) => ({
    id: f.id,
    label: f.id.toUpperCase(),
    disabled: disabledFormat(f.id),
  }))}
  active={format}
  onchange={(id) => onformat(id as LyricsFormat)}
/>

<div
  class="flex items-center gap-2 px-3 py-2 border-b border-rule overflow-x-auto"
>
  {#if !isRich}
    <span class="text-xs font-semibold text-ink">Sync</span>
    {#each LEVELS as l}
      {@const disabled = disabledLevel(l)}
      <button
        type="button"
        onclick={() => l !== level && onlevel(l)}
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
    <div class="ml-auto flex items-center gap-2">
      {@render children()}
    </div>
  {/if}
</div>
