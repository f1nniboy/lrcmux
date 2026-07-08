<script lang="ts">
  import type { PageData } from "./+page.server";
  import { toSlug } from "$lib/slug";
  import Meta from "$lib/Meta.svelte";
  import UseItInYourApp from "$components/UseItInYourApp.svelte";

  let { data }: { data: PageData } = $props();
  const track = $derived(data.lyrics.track);

  const description = $derived.by(() => {
    const level = data.lyrics.meta.level;
    const qualifier = level === "none" ? "" : `${level}-synced `;
    return `View the full ${qualifier}lyrics for ${track.title} by ${track.artist}.`;
  });
</script>

<Meta
  title="{data.title} by {data.artist}"
  {description}
  og={{ type: "music.song", image: track.cover.medium }}
/>

<div
  class="flex flex-col gap-12 sm:gap-16 w-full max-w-3xl min-w-0 px-5 sm:px-8 pt-12 pb-16 flex-1"
>
  <div class="flex flex-col gap-8">
    <header class="flex items-start gap-4">
      {#if track.cover.medium}
        <img
          src={track.cover.medium}
          alt=""
          class="w-20 h-20 sm:w-24 sm:h-24 object-cover rounded-md shadow-md shrink-0"
        />
      {:else}
        <div
          class="w-20 h-20 sm:w-24 sm:h-24 rounded-md bg-paper-2 shrink-0"
        ></div>
      {/if}
      <div class="flex-1 min-w-0 pt-1">
        <h2 class="text-2xl sm:text-3xl font-semibold text-ink leading-tight">
          {track.title}
        </h2>
        <a
          href={`/s/${toSlug(track.artist)}`}
          class="text-muted mt-1 hover:underline no-underline"
        >
          {track.artist}
        </a>
      </div>
    </header>

    <div class="text-ink text-base leading-9 whitespace-pre-wrap">
      {data.lyrics.lines.map((l) => l.text).join("\n")}
    </div>

    {#if data.lyrics.meta.source}
      <p class="text-xs text-muted text-center">
        &copy; {data.lyrics.meta.source.name}
      </p>
    {/if}
  </div>

  <hr class="m-0 border-rule" />

  <UseItInYourApp />
</div>
