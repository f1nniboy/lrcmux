<script lang="ts">
  import type { ClassValue } from "svelte/elements";
  import type { SimpleIcon } from "simple-icons";

  interface Props {
    icon: SimpleIcon;
    size?: number;
    class?: ClassValue;
  }
  let { icon, size = 16, class: cls }: Props = $props();
</script>

<svg
  class={["shrink-0", cls]}
  aria-hidden="true"
  fill="currentColor"
  height={size}
  viewBox="0 0 24 24"
  width={size}
>
  <path d={icon.path} />
</svg>
