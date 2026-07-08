import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";
import { getTrending, deezerToAPITrack } from "$lib/deezer";

export const GET: RequestHandler = async ({ fetch, url }) => {
  const limit = Math.min(
    parseInt(url.searchParams.get("limit") ?? "12", 10),
    50,
  );
  try {
    const tracks = await getTrending({ fetch }, limit);
    return json(
      tracks.map((dt) => ({ ...deezerToAPITrack(dt), isrc: String(dt.id) })),
      { headers: { "cache-control": "public, max-age=3600" } },
    );
  } catch {
    return json([]);
  }
};
