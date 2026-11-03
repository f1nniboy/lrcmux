<script lang="ts">
  import type { HTMLAnchorAttributes } from "svelte/elements";

  let { children, ...rest }: HTMLAnchorAttributes = $props();
</script>

<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
<a {...rest} rel="noopener noreferrer" target="_blank">{@render children?.()}</a
>
