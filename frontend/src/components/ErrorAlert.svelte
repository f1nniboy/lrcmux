<script lang="ts">
  import type { Snippet } from "svelte";
  interface Props {
    message?: string;
    heading?: string;
    children?: Snippet;
  }
  let {
    message,
    heading = "Something went wrong.",
    children,
  }: Props = $props();
</script>

<div class="flex flex-col items-center justify-center py-16">
  <h1 class="text-2xl sm:text-3xl font-semibold text-ink mb-2 text-center">
    {heading}
  </h1>
  <p class="text-muted text-center">
    {#if children}
      {@render children()}
    {:else}
      {message}
    {/if}
  </p>
</div>
