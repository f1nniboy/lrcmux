<script lang="ts">
  import "../app.css";
  import Header from "$components/Header.svelte";

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

<div class="min-h-screen flex flex-col">
  <Header />
  <main class="flex-1 flex flex-col items-center">
    {@render children()}
  </main>
</div>
