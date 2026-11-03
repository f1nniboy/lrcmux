<script lang="ts">
  import type { ClassValue } from "svelte/elements";

  let { class: cls }: { class?: ClassValue } = $props();
</script>

<span class={["font-bold tracking-tight", cls]}
  >lrc<span class="text-cue">mux</span></span
>
