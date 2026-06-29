import type { DeezerSearchResponse, DeezerTrack } from "./types";

const ENDPOINT = "https://api.deezer.com/search";

export interface SearchOptions {
  limit?: number;
  signal?: AbortSignal;
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

export async function searchDeezer(
  query: string,
  opts: SearchOptions = {},
): Promise<DeezerTrack[]> {
  const q = query.trim();
  if (!q) return [];

  const params = new URLSearchParams({
    q,
    limit: String(opts.limit ?? 8),
  });

  const data = await jsonp<
    DeezerSearchResponse | { error?: { message?: string } }
  >(`${ENDPOINT}?${params.toString()}`, opts.signal);

  if ("error" in data && data.error?.message) {
    throw new Error(`Deezer: ${data.error.message}`);
  }
  return ((data as DeezerSearchResponse).data ?? []).filter(
    (t) =>
      t.title &&
      t.artist?.name &&
      (t as DeezerTrack & { type?: string }).type !== "podcast" &&
      (t as DeezerTrack & { type?: string; readable?: boolean }).readable !==
        false &&
      // exclude DJ mixes, audiobook chapters, and other non-song "tracks"
      t.duration > 0 &&
      t.duration <= 900,
  );
}

export async function getTrending(
  limit = 12,
  signal?: AbortSignal,
): Promise<DeezerTrack[]> {
  const url = `https://api.deezer.com/chart/0/tracks?limit=${limit}`;
  const data = await jsonp<
    DeezerSearchResponse | { error?: { message?: string } }
  >(url, signal);
  if ("error" in data && data.error?.message) {
    throw new Error(`Deezer: ${data.error.message}`);
  }
  return ((data as DeezerSearchResponse).data ?? []).filter(
    (t) => t.title && t.artist?.name,
  );
}

export function makeDebouncedSearch(delayMs = 300) {
  let timeout: ReturnType<typeof setTimeout> | null = null;
  let controller: AbortController | null = null;

  function abort() {
    if (timeout) {
      clearTimeout(timeout);
      timeout = null;
    }
    if (controller) {
      controller.abort();
      controller = null;
    }
  }

  function run(
    query: string,
    opts: Omit<SearchOptions, "signal"> = {},
  ): Promise<DeezerTrack[]> {
    abort();
    return new Promise((resolve, reject) => {
      timeout = setTimeout(async () => {
        controller = new AbortController();
        try {
          const tracks = await searchDeezer(query, {
            ...opts,
            signal: controller.signal,
          });
          resolve(tracks);
        } catch (err) {
          if ((err as Error).name === "AbortError") {
            resolve([]);
            return;
          }
          reject(err);
        } finally {
          controller = null;
          timeout = null;
        }
      }, delayMs);
    });
  }

  return { run, abort };
}
