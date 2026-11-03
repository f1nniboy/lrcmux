<script lang="ts">
  import type { Snippet } from "svelte";
  import { Dialog } from "bits-ui";
  import { X } from "@lucide/svelte";

  interface Props {
    open?: boolean;
    // the button that opens it, rendered in place
    trigger?: Snippet;
    triggerClass?: string;
    icon?: string;
    title: Snippet;
    description?: Snippet;
    children: Snippet;
  }
  let {
    open = $bindable(false),
    trigger,
    triggerClass,
    icon,
    title,
    description,
    children,
  }: Props = $props();
</script>

<Dialog.Root bind:open>
  {#if trigger}
    <Dialog.Trigger class={triggerClass}>{@render trigger()}</Dialog.Trigger>
  {/if}
  <Dialog.Portal>
    <Dialog.Overlay class="dialog-overlay" />
    <Dialog.Content class="dialog">
      <div class="flex items-center gap-3 pr-8">
        {#if icon}
          <img
            class="size-10 shrink-0 rounded-lg object-cover"
            alt=""
            src={icon}
          />
        {/if}
        <div class="flex min-w-0 flex-col gap-1">
          <Dialog.Title class="text-lg/tight font-semibold"
            >{@render title()}</Dialog.Title
          >
          {#if description}
            <Dialog.Description class="text-sm text-muted-foreground"
              >{@render description()}</Dialog.Description
            >
          {/if}
        </div>
      </div>
      {@render children()}
      <Dialog.Close
        class="absolute top-4 right-4 icon-button"
        aria-label="Close"><X /></Dialog.Close
      >
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
