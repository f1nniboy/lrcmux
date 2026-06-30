<script lang="ts">
  import { getLyricsText, downloadURL, LyricsError } from "$lib/api";
  import Spinner from "$components/Spinner.svelte";
  import FormatBar from "$components/FormatBar.svelte";
  import ActionButton from "$components/ActionButton.svelte";
  import type {
    Track,
    LyricsFormat,
    LyricsResult,
    SyncLevel,
  } from "$lib/types";
  import { FORMATS, LEVEL_RANK } from "$lib/types";

  interface Props {
    track: Track;
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
      const text = await getLyricsText(track.artist, track.title, format, {
        level,
      });
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
    downloadURL(track.artist, track.title, activeFormat, {
      level: activeLevel,
    }),
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
    <ActionButton
      onclick={async () => {
        if (text) await navigator.clipboard.writeText(text);
      }}
      label="Copy"
    >
      {#snippet icon()}
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          xmlns="http://www.w3.org/2000/svg"
          aria-hidden="true"
        >
          <path
            d="M2 4a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v4h4a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H10a2 2 0 0 1-2-2v-4H4a2 2 0 0 1-2-2zm8 12v4h10V10h-4v4a2 2 0 0 1-2 2zm4-2V4H4v10z"
            fill="currentColor"
          />
        </svg>
      {/snippet}
    </ActionButton>

    <ActionButton href={dlURL} download label="Download">
      {#snippet icon()}
        <svg
          width="16"
          height="16"
          viewBox="0 0 16 16"
          xmlns="http://www.w3.org/2000/svg"
          aria-hidden="true"
        >
          <path
            fill-rule="evenodd"
            d="M14 9a1 1 0 0 1 1 1v3a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2v-3a1 1 0 0 1 2 0v3h10v-3a1 1 0 0 1 1-1M8 1a1 1 0 0 1 1 1v4.586l1.293-1.293a1 1 0 1 1 1.414 1.414L8 10.414 4.293 6.707a1 1 0 0 1 1.414-1.414L7 6.586V2a1 1 0 0 1 1-1"
            fill="currentColor"
          />
        </svg>
      {/snippet}
    </ActionButton>
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
