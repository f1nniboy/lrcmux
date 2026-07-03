<script lang="ts">
  import Wordmark from "$components/Wordmark.svelte";
</script>

<footer class="border-t border-rule bg-paper mt-auto">
  <div
    class="mx-auto w-full max-w-3xl flex items-center justify-between gap-4 px-5 sm:px-8 h-14"
  >
    <span class="text-sm text-muted"
      >&copy; {new Date().getFullYear()} <Wordmark /></span
    >

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
