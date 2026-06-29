<script lang="ts">
  import { getLyricsText, downloadURL, LyricsError } from "$lib/api";
  import Spinner from "$components/Spinner.svelte";
  import FormatBar from "$components/FormatBar.svelte";
  import type {
    DeezerTrack,
    LyricsFormat,
    LyricsResult,
    SyncLevel,
  } from "$lib/types";
  import { FORMATS, LEVEL_RANK } from "$lib/types";

  interface Props {
    track: DeezerTrack;
    result: LyricsResult;
  }
  let { track, result }: Props = $props();

  let activeFormat = $state<LyricsFormat>("txt");
  let activeLevel = $state<SyncLevel>("word");
  let cache = $state<Record<string, string>>({});

  $effect(() => {
    activeLevel = result.meta.level;
  });
  let loading = $state(false);
  let err = $state<string | null>(null);

  async function loadContent(format: LyricsFormat, level: SyncLevel) {
    const key = `${format}:${level}`;
    if (cache[key] != null) return;
    loading = true;
    err = null;
    try {
      const text = await getLyricsText(track, format, { level });
      cache = { ...cache, [key]: text };
    } catch (e) {
      err = e instanceof LyricsError ? e.message : (e as Error).message;
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    if (activeFormat === "txt") return;
    void loadContent(activeFormat, activeLevel);
  });

  const text = $derived.by(() => {
    // so we don't do a useless request for .txt format
    if (activeFormat === "txt" && result) {
      return result.lines.map((l) => l.text).join("\n");
    }

    const raw = cache[`${activeFormat}:${activeLevel}`];
    if (!raw) return "";
    if (activeFormat === "json") {
      try {
        return JSON.stringify(JSON.parse(raw), null, 2);
      } catch {
        return raw;
      }
    }
    return raw;
  });
  const dlURL = $derived(
    downloadURL(track, activeFormat, { level: activeLevel }),
  );
  const isRich = $derived(
    FORMATS.find((f) => f.id === activeFormat)?.rich ?? false,
  );
</script>

<section class="border border-rule rounded-md overflow-hidden bg-paper-2/40">
  <FormatBar
    format={activeFormat}
    level={activeLevel}
    maxLevel={result.meta.level}
    onformat={(format) => {
      activeFormat = format;
      err = null;

      // clamp sync level to min level of selected format
      const f = FORMATS.find((x) => x.id === format);
      if (f?.minLevel && LEVEL_RANK[activeLevel] < LEVEL_RANK[f.minLevel]) {
        activeLevel = f.minLevel;
      }
    }}
    onlevel={(level) => {
      activeLevel = level;
      err = null;
    }}
  >
    <a
      href={dlURL}
      class="ml-auto inline-flex items-center gap-1.5 bg-cue text-ink dark:text-paper px-3 py-1.5 rounded-md text-xs font-medium hover:brightness-110 transition-all shadow-sm no-underline"
    >
      Download .{activeFormat} <span aria-hidden="true">↓</span>
    </a>
  </FormatBar>

  <div class="relative">
    <pre
      class="p-5 h-112 overflow-auto text-ink"
      class:font-mono={!isRich}
      class:font-sans={isRich}
      class:text-xs={!isRich}
      class:text-base={isRich}
      class:leading-relaxed={!isRich}
      class:leading-9={isRich}
      class:whitespace-pre={!isRich}
      class:whitespace-pre-wrap={isRich}>{text ||
        (loading ? "" : err ? "" : "empty...")}</pre>

    {#if loading}
      <div
        class="absolute inset-0 flex items-center justify-center text-muted bg-paper-2/70"
      >
        <Spinner size={36} />
      </div>
    {/if}

    {#if err && !loading}
      <div class="absolute inset-0 flex items-center justify-center p-5">
        <p class="text-sm text-peak text-center max-w-md">{err}</p>
      </div>
    {/if}
  </div>
</section>
