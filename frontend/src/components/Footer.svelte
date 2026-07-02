<script lang="ts">
  import Wordmark from "./Wordmark.svelte";
</script>

<footer class="border-t border-rule mt-auto">
  <div
    class="mx-auto w-full max-w-3xl flex items-center justify-between gap-4 px-5 sm:px-8 h-14"
  >
    <Wordmark class="text-sm tracking-tight" />

    <nav class="flex items-center gap-6 text-sm shrink-0">
      <a
        href="https://matrix.to/#/#lrcmux:oss.zone"
        target="_blank"
        rel="noopener"
        class="text-muted hover:text-ink transition-colors">Matrix room</a
      >
    </nav>
  </div>
</footer>
