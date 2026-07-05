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
  const textClass = $derived(
    isRich
      ? "font-sans text-base leading-9 whitespace-pre-wrap"
      : "font-mono text-xs leading-relaxed whitespace-pre",
  );
</script>

<section
  class="flex-1 flex flex-col border border-rule rounded-md overflow-hidden bg-paper-2"
>
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
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <rect x="9" y="9" width="13" height="13" rx="2" />
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
      {/snippet}
    </ActionButton>

    <ActionButton href={dlURL} download label="Download">
      {#snippet icon()}
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <path d="M12 15V3" />
          <path d="m8 11 4 4 4-4" />
          <path d="M3 17v2a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-2" />
        </svg>
      {/snippet}
    </ActionButton>
  </FormatBar>

  <div class="relative">
    <pre
      aria-hidden="true"
      class="invisible p-5 w-full text-ink {textClass}">{text}</pre>
    <textarea
      readonly
      spellcheck="false"
      class="absolute inset-0 p-5 w-full h-full overflow-hidden text-ink bg-transparent resize-none outline-none {textClass}"
      >{text}</textarea
    >

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
