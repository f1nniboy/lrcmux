import type {
  DeezerArtist,
  DeezerSearchResponse,
  DeezerTrack,
  Track,
} from "./types";

const BASE_URL = "https://api.deezer.com";

export interface DeezerRequestOptions {
  type?: "jsonp" | "json";
  fetch?: typeof globalThis.fetch;
  signal?: AbortSignal;
}

export interface SearchOptions extends DeezerRequestOptions {
  limit?: number;
}

let jsonpSeq = 0;

function jsonp<T>(url: string, signal?: AbortSignal): Promise<T> {
  return new Promise((resolve, reject) => {
    const cb = `lrcmuxJsonp_${++jsonpSeq}`;
    const script = document.createElement("script");
    let settled = false;

    const cleanup = () => {
      if (settled) return;
      settled = true;
      try {
        delete (window as unknown as Record<string, unknown>)[cb];
      } catch {
        (window as unknown as Record<string, unknown>)[cb] = undefined;
      }
      script.remove();
    };

    (window as unknown as Record<string, (d: T) => void>)[cb] = (data: T) => {
      cleanup();
      resolve(data);
    };

    script.onerror = () => {
      cleanup();
      reject(new Error("Deezer request failed"));
    };

    if (signal) {
      if (signal.aborted) {
        cleanup();
        reject(new DOMException("Aborted", "AbortError"));
        return;
      }
      signal.addEventListener(
        "abort",
        () => {
          cleanup();
          reject(new DOMException("Aborted", "AbortError"));
        },
        { once: true },
      );
    }

    const sep = url.includes("?") ? "&" : "?";
    script.src = `${url}${sep}output=jsonp&callback=${cb}`;
    document.head.appendChild(script);
  });
}

async function deezerRequest<T>(
  opts: DeezerRequestOptions,
  path: string,
  params: Record<string, string> = {},
): Promise<T> {
  const url = new URL(`${BASE_URL}${path}`);
  for (const [k, v] of Object.entries(params)) url.searchParams.set(k, v);

  let data: unknown;
  if (opts.type === "json") {
    const fetchFn = opts.fetch ?? globalThis.fetch;
    const res = await fetchFn(url.toString(), { signal: opts.signal });
    if (!res.ok) throw new Error(`Deezer request failed (${res.status})`);
    data = await res.json();
  } else {
    data = await jsonp<unknown>(url.toString(), opts.signal);
  }

  const err = (data as { error?: { message?: string } })?.error;
  if (err?.message) throw new Error(`Deezer: ${err.message}`);

  return data as T;
}

export function deezerToAPITrack(dt: DeezerTrack): Track {
  return {
    isrc: dt.isrc,
    title: dt.title,
    duration: dt.duration,
    artist: dt.artist.name,
    album: dt.album.title,
    cover: {
      small: dt.album.cover_small,
      medium: dt.album.cover_medium,
      big: dt.album.cover_big,
    },
  };
}

export async function searchDeezer(
  opts: SearchOptions,
  query: string,
): Promise<DeezerTrack[]> {
  const q = query.trim();
  if (!q) return [];

  const data = await deezerRequest<DeezerSearchResponse>(opts, "/search", {
    q,
    limit: String(opts.limit ?? 8),
  });

  return (data.data ?? []).filter(
    (t) =>
      t.type !== "podcast" &&
      t.readable !== false &&
      // exclude DJ mixes, audiobook chapters, and other non-song "tracks"
      t.duration > 0 &&
      t.duration <= 900,
  );
}

export async function getTrending(
  opts: DeezerRequestOptions,
  limit = 12,
): Promise<DeezerTrack[]> {
  const data = await deezerRequest<DeezerSearchResponse>(
    opts,
    "/chart/0/tracks",
    { limit: String(limit) },
  );
  return (data.data ?? []).filter((t) => t.title && t.artist?.name);
}

export async function getArtistByName(
  opts: DeezerRequestOptions,
  name: string,
): Promise<DeezerArtist | null> {
  const data = await deezerRequest<{ data: DeezerArtist[] }>(
    opts,
    "/search/artist",
    { q: name, limit: "1" },
  );
  return data.data?.[0] ?? null;
}

export async function getArtistTopTracks(
  opts: DeezerRequestOptions,
  artistId: number,
  limit = 50,
): Promise<DeezerTrack[]> {
  const data = await deezerRequest<DeezerSearchResponse>(
    opts,
    `/artist/${artistId}/top`,
    { limit: String(limit) },
  );
  return data.data ?? [];
}
