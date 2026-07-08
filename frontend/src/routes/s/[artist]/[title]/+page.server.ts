import { error } from "@sveltejs/kit";
import { env } from "$env/dynamic/private";
import type { LyricsResult } from "$lib/types";
import { fromSlug } from "$lib/slug";
import type { PageServerLoad } from "./$types";

const apiUrl = env.API_URL ?? "http://localhost:8080";

export type { PageData } from "./$types";

export const load: PageServerLoad = async ({
  params,
  fetch,
  setHeaders,
  request,
}) => {
  const artist = fromSlug(params.artist);
  const title = fromSlug(params.title);

  const origUA = request.headers.get("user-agent");
  const ua = origUA ? `lrcmux ${origUA}` : "lrcmux";

  const res = await fetch(
    `${apiUrl}/get?${new URLSearchParams({ artist, title })}`,
    { headers: { "User-Agent": ua } },
  ).catch(() => null);

  if (!res) error(500);
  if (res.status === 404)
    error(404, "We couldn't find the lyrics for this track.");
  if (!res.ok) error(res.status);

  const lyrics: LyricsResult = await res.json();
  setHeaders({ "cache-control": "public, max-age=86400" });
  return { artist, title, lyrics };
};
