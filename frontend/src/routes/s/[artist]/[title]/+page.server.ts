import { env } from "$env/dynamic/private";
import type { LyricsResult } from "$lib/types";
import { fromSlug } from "$lib/slug";
import type { PageServerLoad } from "./$types";

const apiUrl = env.API_URL ?? "http://localhost:8080";

export type { PageData } from "./$types";

export const load: PageServerLoad = async ({ params, fetch, setHeaders }) => {
  const artist = fromSlug(params.artist);
  const title = fromSlug(params.title);

  const p = new URLSearchParams({
    artist,
    title,
    format: "json",
    fetch: "cache",
  });

  try {
    const res = await fetch(`${apiUrl}/get?${p}`);
    if (res.ok) {
      const result: LyricsResult = await res.json();
      setHeaders({ "cache-control": "public, max-age=86400" });
      return { artist, title, lyrics: result };
    }
  } catch {}

  return { artist, title };
};
