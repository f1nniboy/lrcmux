import { getTrending, deezerToAPITrack } from "$lib/deezer";
import type { PageServerLoad } from "./$types";
import type { Track } from "$lib/types";

export const load: PageServerLoad = async ({ fetch, setHeaders }) => {
  let trending: Track[] = [];
  try {
    const tracks = await getTrending({ fetch }, 12);
    trending = tracks.map((dt) => ({
      ...deezerToAPITrack(dt),
      isrc: String(dt.id),
    }));
  } catch {
    // ignore
  }
  if (trending.length > 0)
    setHeaders({ "cache-control": "public, max-age=3600" });
  return { trending };
};
