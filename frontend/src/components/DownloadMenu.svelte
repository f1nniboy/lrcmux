<script lang="ts" module>
  import type { LyricsFormat, SyncLevel } from "$lib/types";

  interface Format {
    label: string;
    format: LyricsFormat;
    level: SyncLevel;
    minLevel: SyncLevel;
  }

  const RANK: Record<SyncLevel, number> = { none: 0, line: 1, word: 2 };

  const FORMATS: Format[] = [
    { label: "LRC", format: "lrc", level: "line", minLevel: "line" },
    { label: "Enhanced LRC", format: "lrc", level: "word", minLevel: "word" },
    { label: "SRT", format: "srt", level: "line", minLevel: "line" },
    { label: "WebVTT", format: "vtt", level: "line", minLevel: "line" },
    { label: "Plain text", format: "txt", level: "none", minLevel: "none" },
    { label: "JSON", format: "json", level: "word", minLevel: "none" },
  ];
</script>

<script lang="ts">
  import { DropdownMenu } from "bits-ui";
  import { ChevronDown, Download } from "@lucide/svelte";
  import { API_URL } from "$lib/env";

  interface Props {
    // the page's own query, so the download hits the cache
    artist: string;
    title: string;
    level: SyncLevel;
  }
  let { artist, title, level }: Props = $props();

  const formats = $derived(
    FORMATS.filter((f) => RANK[f.minLevel] <= RANK[level]),
  );

  function href(f: Format) {
    const p = new URLSearchParams({
      artist,
      title,
      format: f.format,
      level: f.level,
      download: "true",
    });
    return `${API_URL}/get?${p}`;
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class="button-cue self-start">
    <Download />Download<ChevronDown class="-mr-1 opacity-70" />
  </DropdownMenu.Trigger>
  <DropdownMenu.Portal>
    <DropdownMenu.Content
      class="z-40 min-w-48 pop popover p-1"
      align="start"
      sideOffset={6}
    >
      {#each formats as f (f.label)}
        <DropdownMenu.Item class="menu-item">
          {#snippet child({ props })}
            <!-- eslint-disable svelte/no-navigation-without-resolve -->
            <a {...props} href={href(f)}>
              {f.label}<span class="font-mono text-xs text-muted-foreground"
                >.{f.format}</span
              >
            </a>
            <!-- eslint-enable svelte/no-navigation-without-resolve -->
          {/snippet}
        </DropdownMenu.Item>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>
