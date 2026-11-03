<script lang="ts">
  interface Props {
    size?: number;
  }
  let { size = 16 }: Props = $props();

  let stroke = $derived(Math.max(2, Math.round(size / 15)));
</script>

<span
  style:width="{size}px"
  style:height="{size}px"
  style:border-width="{stroke}px"
  style:border-top-color="transparent"
  class="box-border inline-block animate-spin rounded-full border-solid border-current align-middle opacity-70"
  aria-label="Loading"
  role="status"
></span>
