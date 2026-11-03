<script lang="ts">
  import { resolve } from "$app/paths";
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

<div class="page flex-1 items-center justify-center gap-6 py-24 text-center">
  <div class="flex flex-col gap-1.5">
    <p class="font-mono text-sm text-cue-text">{page.status}</p>
    <h1 class="section-title">{message}</h1>
  </div>
  {#if page.route.id !== "/"}
    <a class="button" href={resolve("/")}>Go home</a>
  {/if}
</div>
