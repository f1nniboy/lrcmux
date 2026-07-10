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
  {#if description}
    <meta name="description" content={description} />
  {/if}
  {#if og}
    <meta content={og.type ?? "website"} property="og:type" />
    <meta content={page.url.href} property="og:url" />
    <meta content={title || page.url.hostname} property="og:title" />
    {#if title}
      <meta content={page.url.hostname} property="og:site_name" />
    {/if}
    {#if description}
      <meta content={description} property="og:description" />
    {/if}
    {#if og.image}
      <meta
        content={new URL(og.image, page.url.origin).href}
        property="og:image"
      />
    {/if}
  {/if}
</svelte:head>
