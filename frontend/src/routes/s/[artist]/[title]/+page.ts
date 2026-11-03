import { fromSlug } from "$lib/slug";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ params, setHeaders }) => {
  setHeaders({ "cache-control": "public, max-age=86400" });
  return { artist: fromSlug(params.artist), title: fromSlug(params.title) };
};
