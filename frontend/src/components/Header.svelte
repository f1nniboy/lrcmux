<script lang="ts">
  import { page } from "$app/state";
  import Wordmark from "./Wordmark.svelte";
  import SearchBar from "./SearchBar.svelte";

  const isHome = $derived(page.url.pathname === "/");
</script>

<header class="border-b border-rule">
  <div
    class="mx-auto w-full max-w-3xl flex items-center justify-between gap-4 px-5 sm:px-8 h-14"
  >
    <a href="/" class="text-ink hover:opacity-90 transition-opacity shrink-0">
      <Wordmark class="text-xl tracking-tight" />
    </a>

    {#if !isHome}
      <div class="flex-1 max-w-md min-w-0">
        <SearchBar slim={true} />
      </div>
    {/if}

    <nav class="flex items-center gap-6 text-sm shrink-0">
      <a href="/docs" class="text-muted hover:text-ink transition-colors">API</a
      >
      <a
        href="https://github.com/f1nniboy/lrcmux"
        target="_blank"
        rel="noopener"
        class="text-muted hover:text-ink transition-colors">Source</a
      >
    </nav>
  </div>
</header>
