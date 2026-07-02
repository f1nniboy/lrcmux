import { error } from "@sveltejs/kit";
import { fromSlug } from "$lib/slug";
import {
  getArtistByName,
  getArtistTopTracks,
  deezerToAPITrack,
  type DeezerRequestOptions,
} from "$lib/deezer";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ params, fetch, setHeaders }) => {
  const name = fromSlug(params.artist);
  const opts: DeezerRequestOptions = { type: "json", fetch };

  const artist = await getArtistByName(opts, name);
  if (!artist) error(404, "The specified artist couldn't be found.");

  const deezerTracks = await getArtistTopTracks(opts, artist.id);
  const tracks = deezerTracks.map((dt) => ({
    ...deezerToAPITrack(dt),
    isrc: dt.isrc || String(dt.id),
  }));

  setHeaders({ "cache-control": "public, max-age=3600" });
  return { artist, tracks };
};
