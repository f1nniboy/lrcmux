<script lang="ts">
  import { page } from "$app/state";

  interface Props {
    title?: string;
    description?: string;
    og?: { type?: "website" | "music.song"; image?: string };
  }
  let { title, description, og }: Props = $props();
</script>

<svelte:head>
  <title>{title ? `${title} - ${page.url.hostname}` : page.url.hostname}</title>
  <meta property="og:type" content={og?.type ?? "website"} />
  <meta property="og:url" content={page.url.href} />
  {#if title}
    <meta property="og:site_name" content={page.url.hostname} />
  {/if}
  <meta property="og:title" content={title || page.url.hostname} />
  {#if description}
    <meta name="description" content={description} />
    <meta property="og:description" content={description} />
  {/if}
  {#if og?.image}
    <meta property="og:image" content={og.image} />
  {/if}
</svelte:head>
