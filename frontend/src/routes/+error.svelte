<script lang="ts">
  import { page } from "$app/state";

  // horrible, but there's no built-in HTTP code -> message map
  const HTTP_MESSAGE: Record<number, string> = {
    404: "Not Found",
    429: "Too Many Requests",
    500: "Internal Error",
  };

  const message = $derived(
    page.error?.message && page.error.message !== HTTP_MESSAGE[page.status]
      ? page.error.message
      : page.status === 404
        ? "This page doesn't exist."
        : page.status === 429
          ? "Slow down a bit."
          : "Something went wrong.",
  );
</script>

<svelte:head>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div
  class="flex-1 flex flex-col items-center justify-center gap-6 px-5 min-w-0"
>
  <h1 class="text-9xl font-black text-cue tracking-tight leading-none">
    {page.status}
  </h1>
  <p class="text-muted text-center max-w-sm">
    {message}
  </p>
  {#if page.url.pathname !== "/"}
    <a
      href="/"
      class="inline-flex items-center gap-1.5 text-sm font-medium text-paper bg-ink px-4 py-2 rounded-md hover:opacity-80 transition-all no-underline"
    >
      Go home
    </a>
  {/if}
</div>
