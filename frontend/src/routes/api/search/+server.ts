import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";
import { searchDeezer, deezerToAPITrack } from "$lib/deezer";

export const GET: RequestHandler = async ({ url, fetch }) => {
  const q = url.searchParams.get("q") ?? "";
  const limit = Math.min(
    parseInt(url.searchParams.get("limit") ?? "8", 10),
    20,
  );

  if (q.trim().length < 2) return json([]);

  try {
    const results = await searchDeezer({ fetch, limit }, q);
    return json(results.map(deezerToAPITrack));
  } catch {
    return json([]);
  }
};
