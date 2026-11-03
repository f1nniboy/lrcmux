<script lang="ts">
  import type { HTMLAnchorAttributes } from "svelte/elements";
  import type { Track } from "$lib/types";
  import { resolve } from "$app/paths";
  import { toSlug } from "$lib/slug";
  import { fmtDuration } from "$lib/utils";
  import Cover from "./Cover.svelte";

  interface Props extends HTMLAnchorAttributes {
    track: Track;
    // an artist's own page already names the artist
    artist?: boolean;
    duration?: boolean;
    active?: boolean;
  }
  let {
    track,
    artist = true,
    duration = true,
    active,
    ...rest
  }: Props = $props();
</script>

<a
  {...rest}
  class={[
    "flex items-center gap-3 rounded-lg px-2.5 py-2 transition-colors",
    active ? "bg-cue/15" : "hover:bg-cue/15",
  ]}
  href={resolve("/s/[artist]/[title]", {
    artist: toSlug(track.artist),
    title: toSlug(track.title),
  })}
>
  <Cover class="size-10" lazy src={track.cover.small} />
  <span class="flex min-w-0 flex-1 flex-col">
    <span class="truncate text-sm font-medium">{track.title}</span>
    {#if artist}
      <span class="truncate text-sm text-muted-foreground">{track.artist}</span>
    {/if}
  </span>
  {#if duration}
    <span class="text-sm text-muted-foreground tabular-nums"
      >{fmtDuration(track.duration)}</span
    >
  {/if}
</a>
