<script lang="ts">
  import { resolve } from "$app/paths";
  import { page } from "$app/state";
  import Wordmark from "./Wordmark.svelte";
  import SearchBar from "./SearchBar.svelte";

  const isHome = $derived(page.url.pathname === "/");
</script>

<header class="sticky top-0 z-40 border-b border-rule bg-paper">
  <div
    class="mx-auto w-full max-w-3xl flex items-center justify-between gap-4 px-5 sm:px-8 h-14"
  >
    <a
      class="text-ink hover:opacity-90 transition-opacity shrink-0"
      href={resolve("/")}
    >
      <Wordmark class="text-xl tracking-tight" />
    </a>

    {#if !isHome}
      <div class="flex-1 max-w-md min-w-0">
        <SearchBar slim={true} />
      </div>
    {/if}

    <nav
      class="items-center gap-6 text-sm shrink-0"
      class:flex={isHome}
      class:hidden={!isHome}
      class:sm:flex={!isHome}
    >
      <a
        class="text-muted hover:text-ink transition-colors"
        href={resolve("/docs")}>API docs</a
      >
      <a
        class="text-muted hover:text-ink transition-colors"
        href="https://github.com/f1nniboy/lrcmux"
        rel="noopener noreferrer"
        target="_blank">Source</a
      >
      <a
        class="text-muted hover:text-ink transition-colors"
        href="https://matrix.to/#/#lrcmux:oss.zone"
        rel="noopener noreferrer"
        target="_blank">Matrix</a
      >
    </nav>
  </div>
</header>
