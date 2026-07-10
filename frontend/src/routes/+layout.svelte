<script lang="ts">
  import "../app.css";
  import Header from "$components/Header.svelte";
  import { navigating, page } from "$app/state";
  import { fade } from "svelte/transition";

  let { children } = $props();
</script>

<svelte:head>
  <script async src="/x/script.js"></script>
  <script>
    window.plausible =
      window.plausible ||
      function () {
        (plausible.q = plausible.q || []).push(arguments);
      };
    plausible.init =
      plausible.init ||
      function (i) {
        plausible.o = i || {};
      };
    plausible.init({ endpoint: "/x/event" });
  </script>
</svelte:head>

{#if navigating.to}
  <div
    out:fade={{ duration: 200 }}
    class="nav-bar fixed inset-x-0 top-0 z-50 h-0.5 origin-left bg-cue"
  ></div>
{/if}

<div class="min-h-screen flex flex-col">
  {#if page.url.pathname !== "/"}
    <Header />
  {/if}
  <main class="flex-1 flex flex-col items-center">
    {@render children()}
  </main>
</div>

<style>
  .nav-bar {
    animation: nav-bar-fill 6s ease-out forwards;
  }

  @keyframes nav-bar-fill {
    from {
      transform: scaleX(0);
    }
    to {
      transform: scaleX(0.85);
    }
  }
</style>
