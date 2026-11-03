import { error } from "@sveltejs/kit";
import { fromSlug } from "$lib/slug";
import {
  getArtistByName,
  getArtistTopTracks,
  deezerToAPITrack,
} from "$lib/deezer";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ params, fetch, setHeaders }) => {
  const name = fromSlug(params.artist);

  const artist = await getArtistByName({ fetch }, name);
  if (!artist) error(404, "We couldn't find this artist.");

  const tracks = (await getArtistTopTracks({ fetch }, artist.id)).map(
    deezerToAPITrack,
  );

  setHeaders({ "cache-control": "public, max-age=86400" });
  return { artist, tracks };
};
