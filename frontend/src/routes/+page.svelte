<script lang="ts">
  import { resolve } from "$app/paths";
  import Meta from "$lib/Meta.svelte";
  import SearchBar from "$components/SearchBar.svelte";
  import Wordmark from "$components/Wordmark.svelte";
</script>

<Meta
  description="Free synchronized lyrics for any song. LRC, word-by-word, and plain text, powered by multiple providers. Free lyrics API, no key required."
  og={{ image: "/logo.png" }}
/>

<div
  class="flex flex-col flex-1 items-center justify-center w-full min-w-0 px-5 sm:px-8 gap-7"
>
  <div class="flex flex-col items-center gap-2">
    <h1 class="text-6xl sm:text-7xl font-bold tracking-tight text-ink">
      <Wordmark />
    </h1>
    <p class="text-muted">A lyrics aggregator API</p>
  </div>
  <div class="w-full max-w-lg">
    <SearchBar />
  </div>
  <nav class="flex items-center gap-5 text-sm text-muted">
    <a class="hover:text-ink transition-colors" href={resolve("/apps")}>Apps</a>
    <a class="hover:text-ink transition-colors" href={resolve("/docs")}
      >API docs</a
    >
    <a
      class="hover:text-ink transition-colors"
      href="https://github.com/f1nniboy/lrcmux"
      rel="noopener noreferrer"
      target="_blank">Source</a
    >
    <a
      class="hover:text-ink transition-colors"
      href="https://matrix.to/#/#lrcmux:oss.zone"
      rel="noopener noreferrer"
      target="_blank">Matrix</a
    >
  </nav>
</div>
