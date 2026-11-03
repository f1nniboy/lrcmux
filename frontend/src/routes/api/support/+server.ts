import { json } from "@sveltejs/kit";
import { platforms, type PlatformId, type Totals } from "$lib/support";
import type { RequestHandler } from "./$types";

type Fetch = typeof fetch;

function platformUrl(id: PlatformId) {
  const platform = platforms.find((p) => p.id === id);
  if (!platform) throw new Error(`no ${id} platform`);
  return platform.url;
}

// the public sponsors page needs no token, and its goal is the one the site shows
async function github(fetch: Fetch) {
  const res = await fetch(platformUrl("github"));
  if (!res.ok) throw new Error(`github: status ${res.status}`);
  const text = (await res.text()).replace(/<[^>]+>/g, " ").replace(/\s+/g, " ");
  const goal = text.match(/(\d+)% towards \$([\d,.]+) per month goal/);
  if (!goal) throw new Error("github: no monthly goal on the page");
  const target = Number(goal[2].replaceAll(",", ""));
  return { raised: (Number(goal[1]) / 100) * target, goal: target };
}

// liberapay counts weekly, in the account's own currency
async function liberapay(fetch: Fetch) {
  const res = await fetch(`${platformUrl("liberapay")}/public.json`);
  if (!res.ok) throw new Error(`liberapay: status ${res.status}`);
  const { receiving } = await res.json();
  const weekly =
    Number(receiving.amount) * (await usdRate(fetch, receiving.currency));
  return (weekly * 52) / 12;
}

async function usdRate(fetch: Fetch, currency: string) {
  if (currency === "USD") return 1;
  const res = await fetch(
    `https://api.frankfurter.dev/v1/latest?base=${currency}&symbols=USD`,
  );
  if (!res.ok) throw new Error(`frankfurter: status ${res.status}`);
  return (await res.json()).rates.USD as number;
}

// ko-fi is left out, its goal counts one-off tips too
export const GET: RequestHandler = async ({ fetch, url }) => {
  // liberapay refuses requests without one
  const get: Fetch = (input, init) =>
    fetch(input, {
      ...init,
      headers: { "user-agent": `lrcmux (${url.origin})`, ...init?.headers },
    });
  try {
    const [gh, lp] = await Promise.all([github(get), liberapay(get)]);
    const totals: Totals = {
      raised: Math.round((gh.raised + lp) * 100) / 100,
      goal: gh.goal,
    };
    return json(totals, {
      headers: { "cache-control": "public, max-age=86400" },
    });
  } catch (e) {
    console.warn("support: totals unavailable:", e);
    return json(null, { headers: { "cache-control": "public, max-age=3600" } });
  }
};
