<script lang="ts">
  import type { ClassValue } from "svelte/elements";
  import { Check, CircleAlert, Copy } from "@lucide/svelte";

  let { content, class: cls }: { content: string; class?: ClassValue } =
    $props();

  let outcome = $state<"copied" | "blocked">();
  let reset: ReturnType<typeof setTimeout> | undefined;

  const label = $derived(
    outcome === "copied"
      ? "Copied"
      : outcome === "blocked"
        ? "The browser blocked copying"
        : "Copy",
  );

  function copy() {
    navigator.clipboard.writeText(content).then(
      () => show("copied"),
      () => show("blocked"),
    );
  }

  function show(result: "copied" | "blocked") {
    outcome = result;
    clearTimeout(reset);
    reset = setTimeout(() => (outcome = undefined), 1500);
  }
</script>

<button
  class={["icon-button", cls]}
  aria-label={label}
  onclick={copy}
  title={outcome ? label : undefined}
  type="button"
>
  {#if outcome === "copied"}
    <Check />
  {:else if outcome === "blocked"}
    <CircleAlert class="text-danger" />
  {:else}
    <Copy />
  {/if}
</button>
