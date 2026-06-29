<script lang="ts">
  interface Props {
    size?: number;
  }
  let { size = 16 }: Props = $props();

  let stroke = $derived(Math.max(2, Math.round(size / 15)));
</script>

<span
  class="inline-block rounded-full animate-spin align-middle opacity-70 box-border border-current border-solid"
  style:width="{size}px"
  style:height="{size}px"
  style:border-width="{stroke}px"
  style:border-top-color="transparent"
  role="status"
  aria-label="Loading"
></span>
