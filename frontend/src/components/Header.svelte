<script lang="ts">
  import ExternalLink from "./ExternalLink.svelte";
  import { BookOpen } from "@lucide/svelte";
  import { resolve } from "$app/paths";
  import { page } from "$app/state";
  import Wordmark from "./Wordmark.svelte";
  import SearchBar from "./SearchBar.svelte";
  import { siGithub, siMatrix } from "simple-icons";
  import BrandIcon from "./BrandIcon.svelte";
  import SupportDialog from "./SupportDialog.svelte";

  // home has its own, larger search
  const search = $derived(page.route.id !== "/");
</script>

<header class="sticky top-0 z-30 border-b border-border bg-card">
  <div
    class="page grid grid-cols-[1fr_auto] grid-rows-[3.5rem_auto] items-center gap-x-4 sm:grid-cols-[auto_minmax(0,1fr)_auto] sm:grid-rows-[3.5rem] sm:gap-x-6"
  >
    <a
      class="col-start-1 row-start-1 -ml-2.5 justify-self-start rounded-lg px-2.5 py-1 text-xl transition-colors hover:bg-hover"
      href={resolve("/")}><Wordmark /></a
    >
    {#if search}
      <div
        class="col-span-full row-start-2 pb-3 sm:col-span-1 sm:col-start-2 sm:row-start-1 sm:pb-0"
      >
        <SearchBar />
      </div>
    {/if}
    <nav
      class="col-start-2 row-start-1 -mr-2.5 flex items-center gap-1 justify-self-end sm:col-start-3"
    >
      <a class="nav-link" href={resolve("/docs")}
        ><BookOpen /><span class="max-sm:sr-only">Docs</span></a
      >
      <ExternalLink class="nav-link" href="https://github.com/f1nniboy/lrcmux"
        ><BrandIcon icon={siGithub} /><span class="max-sm:sr-only">Source</span
        ></ExternalLink
      >
      <ExternalLink class="nav-link" href="https://matrix.to/#/#lrcmux:oss.zone"
        ><BrandIcon icon={siMatrix} /><span class="max-sm:sr-only">Chat</span
        ></ExternalLink
      >
      <SupportDialog />
    </nav>
  </div>
</header>
