<script lang="ts">
  import ExternalLink from "./ExternalLink.svelte";
  import { ArrowUpRight, Heart } from "@lucide/svelte";
  import { platforms, type Totals } from "$lib/support";
  import BrandIcon from "./BrandIcon.svelte";
  import Modal from "./Modal.svelte";
  import Wordmark from "./Wordmark.svelte";
  import { prefersReducedMotion, Tween } from "svelte/motion";
  import { cubicOut } from "svelte/easing";

  const usd = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    trailingZeroDisplay: "stripIfInteger",
  });

  // undefined while loading, null when they couldn't be fetched
  let totals = $state<Totals | null>();
  const shown = new Tween(0, { easing: cubicOut });
  const share = $derived(totals ? shown.current / totals.goal : 0);

  // the dialog's content only mounts on open, so this runs then
  function load() {
    if (totals !== undefined) return;
    fetch("/api/support")
      .then((res) => res.json() as Promise<Partial<Totals> | null>)
      .then((t) => {
        // a response cached before its shape changed counts as unavailable
        totals =
          typeof t?.raised === "number" && typeof t.goal === "number"
            ? { raised: t.raised, goal: t.goal }
            : null;
        if (totals) {
          void shown.set(totals.raised, {
            duration: prefersReducedMotion.current ? 0 : 700,
          });
        }
      })
      .catch(() => (totals = null));
  }
</script>

<Modal triggerClass="nav-cue">
  {#snippet trigger()}<Heart /><span class="max-sm:sr-only">Support</span
    >{/snippet}
  {#snippet title()}Support <Wordmark />{/snippet}
  {#snippet description()}While it will always be free, donations help cover
    what it costs to run.{/snippet}

  {#if totals !== null}
    <div class="flex flex-col gap-2.5 rounded-xl bg-cue/10 p-4" {@attach load}>
      <div class="flex items-baseline justify-between gap-3">
        <p class="flex items-baseline gap-1.5">
          {#if totals}
            <span class="text-lg/tight font-semibold tabular-nums"
              >{usd.format(shown.current)}</span
            >
            <span class="text-sm text-muted-foreground"
              >of {usd.format(totals.goal)} / month</span
            >
          {:else}
            <span
              class="rounded-md bg-cue/20 text-lg/tight font-semibold text-transparent motion-safe:animate-pulse"
              aria-hidden="true">$0.00</span
            >
            <span
              class="rounded-md bg-cue/20 text-sm text-transparent motion-safe:animate-pulse"
              aria-hidden="true">of $00 / month</span
            >
          {/if}
        </p>
        {#if totals}
          <span class="text-sm font-medium text-cue-text tabular-nums"
            >{Math.round(share * 100)}%</span
          >
        {/if}
      </div>
      <div
        class={[
          "h-2 overflow-hidden rounded-full bg-cue/20",
          !totals && "motion-safe:animate-pulse",
        ]}
        aria-label="Covered by monthly donations"
        aria-valuemax={totals?.goal}
        aria-valuemin={0}
        aria-valuenow={totals?.raised}
        role="progressbar"
      >
        <div
          style:width="{Math.min(share, 1) * 100}%"
          class="h-full rounded-full bg-cue"
        ></div>
      </div>
    </div>
  {/if}

  <ul class="flex flex-col gap-2">
    {#each platforms as p (p.name)}
      <li>
        <ExternalLink
          class="group flex items-center gap-3 card-button px-3.5 py-2.5"
          href={p.url}
        >
          <BrandIcon
            class="text-muted-foreground transition-colors group-hover:text-foreground"
            icon={p.icon}
          /><span class="min-w-0 flex-1 truncate font-medium">{p.name}</span>
          <ArrowUpRight class="shrink-0 text-muted-foreground" />
        </ExternalLink>
      </li>
    {/each}
  </ul>
</Modal>
