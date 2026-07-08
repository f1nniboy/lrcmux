import { error } from "@sveltejs/kit";
import { fromSlug } from "$lib/slug";
import {
  getArtistByName,
  getArtistTopTracks,
  deezerToAPITrack,
} from "$lib/deezer";
import type { PageServerLoad } from "./$types";
import type { Track } from "$lib/types";

export const load: PageServerLoad = async ({ params, fetch, setHeaders }) => {
  const name = fromSlug(params.artist);

  const artist = await getArtistByName({ fetch }, name);
  if (!artist) error(404);

  const deezerTracks = await getArtistTopTracks({ fetch }, artist.id);
  const tracks: Track[] = deezerTracks.map((dt) => ({
    ...deezerToAPITrack(dt),
    isrc: dt.isrc || String(dt.id),
  }));

  setHeaders({ "cache-control": "public, max-age=86400" });
  return { artist, tracks };
};
