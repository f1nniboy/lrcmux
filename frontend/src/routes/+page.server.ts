import { API_URL } from "$lib/server/env";
import { deezerToAPITrack, getChart } from "$lib/deezer";
import type { PageServerLoad } from "./$types";
import type { Source, Track } from "$lib/types";

// the page works without either list, it just leaves that part out
export const load: PageServerLoad = async ({ fetch, setHeaders }) => {
  const [sources, chart]: [Source[], Track[]] = await Promise.all([
    fetch(`${API_URL}/sources`)
      .then((res) => (res.ok ? res.json() : []))
      .catch(() => []),
    getChart({ fetch }, 4)
      .then((tracks) => tracks.map(deezerToAPITrack))
      .catch(() => []),
  ]);
  // a page missing a part shouldn't stay cached for an hour
  if (sources.length && chart.length)
    setHeaders({ "cache-control": "public, max-age=3600" });
  return { sources, chart };
};
