<script lang="ts">
  import { page } from "$app/state";
</script>

<div
  class="flex-1 flex flex-col items-center justify-center gap-6 px-5 min-w-0"
>
  <h1 class="text-9xl font-black text-cue tracking-tight leading-none">
    {page.status}
  </h1>
  <p class="text-muted text-center max-w-sm">
    {page.status === 404 ? "This page doesn't exist." : "Something went wrong."}
  </p>
  <a
    href="/"
    class="inline-flex items-center gap-1.5 text-sm font-medium text-paper bg-ink px-4 py-2 rounded-md hover:opacity-80 transition-all no-underline"
  >
    go home
  </a>
</div>
