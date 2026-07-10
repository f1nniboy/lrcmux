<script lang="ts">
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
    <a href="/apps" class="hover:text-ink transition-colors">Apps</a>
    <a href="/docs" class="hover:text-ink transition-colors">API docs</a>
    <a
      href="https://github.com/f1nniboy/lrcmux"
      target="_blank"
      rel="noopener"
      class="hover:text-ink transition-colors">Source</a
    >
    <a
      href="https://matrix.to/#/#lrcmux:oss.zone"
      target="_blank"
      rel="noopener"
      class="hover:text-ink transition-colors">Matrix</a
    >
  </nav>
</div>
