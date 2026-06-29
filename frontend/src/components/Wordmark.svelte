<script lang="ts">
  interface Props {
    class?: string;
  }
  let { class: cls = "" }: Props = $props();
</script>

<span class="font-bold {cls}"
  ><span class="text-ink">lrc</span><span class="text-cue">mux</span></span
>
