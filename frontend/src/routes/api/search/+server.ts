import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";
import { searchDeezer, deezerToAPITrack } from "$lib/deezer";

export const GET: RequestHandler = async ({ url, fetch }) => {
  const q = url.searchParams.get("q") ?? "";
  const limit = Math.min(Number(url.searchParams.get("limit")) || 8, 20);

  if (q.trim().length < 2) return json([]);

  try {
    const results = await searchDeezer({ fetch, limit }, q);
    return json(results.map(deezerToAPITrack), {
      headers: { "cache-control": "public, max-age=300" },
    });
  } catch {
    return json([], { status: 502 });
  }
};
