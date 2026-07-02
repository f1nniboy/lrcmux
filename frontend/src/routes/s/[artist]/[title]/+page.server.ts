import { env } from "$env/dynamic/private";
import type { LyricsResult } from "$lib/types";
import { fromSlug } from "$lib/slug";

export interface PageData {
  artist: string;
  title: string;
  lyrics?: LyricsResult;
}

const apiUrl = env.API_URL ?? "http://localhost:8080";

export async function load({
  params,
  fetch,
}: {
  params: Record<string, string>;
  fetch: typeof globalThis.fetch;
}) {
  const artist = fromSlug(params.artist ?? "");
  const title = fromSlug(params.title ?? "");

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
      return { artist, title, lyrics: result };
    }
  } catch {}

  return { artist, title };
}
