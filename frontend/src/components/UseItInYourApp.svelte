<script lang="ts">
  import { resolve } from "$app/paths";
</script>

<section
  class="flex flex-col sm:flex-row sm:items-end sm:justify-between gap-5"
>
  <div>
    <h2 class="text-2xl sm:text-3xl font-semibold text-ink mb-2">
      Ready to integrate?
    </h2>
    <p class="text-muted">
      Easy-to-use API, with sane rate limits and no auth required.
    </p>
  </div>
  <a
    class="self-start sm:self-end inline-flex items-center gap-2 bg-ink text-paper px-5 py-3 rounded-md font-medium hover:bg-cue transition-colors"
    href={resolve("/docs")}
  >
    Read the docs
    <svg
      aria-hidden="true"
      fill="none"
      height="16"
      stroke="currentColor"
      stroke-linecap="round"
      stroke-linejoin="round"
      stroke-width="2"
      viewBox="0 0 24 24"
      width="16"
    >
      <path d="M5 12h14" />
      <path d="m12 5 7 7-7 7" />
    </svg>
  </a>
</section>
