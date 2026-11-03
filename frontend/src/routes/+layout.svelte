<script lang="ts">
  import "../app.css";
  import { setLucideProps } from "@lucide/svelte";
  import { onNavigate } from "$app/navigation";
  import Header from "$components/Header.svelte";

  let { children } = $props();

  setLucideProps({ size: 16, strokeWidth: 1.75 });

  onNavigate((navigation) => {
    if (!document.startViewTransition) return;
    return new Promise((resolve) => {
      document.startViewTransition(async () => {
        resolve();
        await navigation.complete;
      });
    });
  });
</script>

<div class="flex min-h-screen flex-col">
  <Header />
  <main class="flex flex-1 flex-col">
    {@render children()}
  </main>
</div>
