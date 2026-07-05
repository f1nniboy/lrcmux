import { error } from "@sveltejs/kit";
import { fromSlug } from "$lib/slug";
import { getArtistByName, type DeezerRequestOptions } from "$lib/deezer";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ params, fetch, setHeaders }) => {
  const name = fromSlug(params.artist);
  const opts: DeezerRequestOptions = { type: "json", fetch };

  const artist = await getArtistByName(opts, name);
  if (!artist) error(404);

  setHeaders({ "cache-control": "public, max-age=86400" });
  return { artist };
};
