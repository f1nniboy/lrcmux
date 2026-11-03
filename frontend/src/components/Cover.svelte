<script lang="ts">
  import type { ClassValue } from "svelte/elements";

  interface Props {
    src?: string;
    round?: boolean;
    lazy?: boolean;
    class?: ClassValue;
  }
  let { src, round = false, lazy = false, class: cls }: Props = $props();

  const classes = $derived([
    "shrink-0 bg-muted object-cover outline -outline-offset-1 outline-foreground/8",
    round ? "rounded-full" : "rounded-xl",
    cls,
  ]);
</script>

{#if src}
  <img class={classes} alt="" loading={lazy ? "lazy" : undefined} {src} />
{:else}
  <div class={classes}></div>
{/if}
