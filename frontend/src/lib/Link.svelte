<script>
  let { href, children } = $props();
</script>

<a
  {href}
  class="text-ink hover:text-cue underline underline-offset-4"
  target="_blank"
  rel="noopener"
>
  {@render children()}
</a>
