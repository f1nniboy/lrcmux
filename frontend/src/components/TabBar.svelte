<script lang="ts">
  interface Tab {
    id: string;
    label: string;
    disabled?: boolean;
  }

  interface Props {
    tabs: Tab[];
    active: string;
    onchange: (id: string) => void;
  }

  let { tabs, active, onchange }: Props = $props();
</script>

<div
  class="flex items-center gap-1 border-b border-rule px-2 py-1.5 overflow-x-auto"
  role="tablist"
>
  {#each tabs as tab (tab.id)}
    <button
      class="px-3 py-1.5 text-sm font-medium rounded-md transition-colors {tab.disabled
        ? 'text-muted opacity-35 cursor-not-allowed'
        : active === tab.id
          ? 'bg-ink text-paper cursor-pointer'
          : 'text-muted hover:text-ink hover:bg-paper cursor-pointer'}"
      aria-selected={active === tab.id}
      disabled={tab.disabled}
      onclick={() => tab.id !== active && onchange(tab.id)}
      role="tab"
      type="button"
    >
      {tab.label}
    </button>
  {/each}
</div>
