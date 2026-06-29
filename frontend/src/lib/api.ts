import type {
  DeezerTrack,
  LyricsFormat,
  LyricsResult,
  SyncLevel,
} from "./types";

export interface GetLyricsOptions {
  level?: SyncLevel;
  signal?: AbortSignal;
}

export class LyricsError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "LyricsError";
  }
}

export async function getLyricsJSON(
  track: DeezerTrack,
  opts: Omit<GetLyricsOptions, "format"> = {},
): Promise<LyricsResult> {
  const res = await fetch(downloadURL(track, "json", opts), {
    signal: opts.signal,
  });
  if (!res.ok) {
    throw new LyricsError(res.status, await failureMessage(res));
  }
  return (await res.json()) as LyricsResult;
}

export async function getLyricsText(
  track: DeezerTrack,
  format: LyricsFormat,
  opts: Omit<GetLyricsOptions, "format"> = {},
): Promise<string> {
  const res = await fetch(downloadURL(track, format, opts), {
    signal: opts.signal,
  });
  if (!res.ok) {
    throw new LyricsError(res.status, await failureMessage(res));
  }
  return res.text();
}

export function downloadURL(
  track: DeezerTrack,
  format: LyricsFormat,
  opts: GetLyricsOptions = {},
): string {
  const params = new URLSearchParams();
  if (track.isrc) {
    params.set("isrc", track.isrc);
  } else {
    params.set("artist", track.artist.name);
    params.set("title", track.title);
  }
  params.set("format", format);
  if (opts.level) params.set("level", opts.level);
  return `/api/get?${params.toString()}`;
}

async function failureMessage(res: Response): Promise<string> {
  if (res.status === 404)
    return "No provider has this track, or it doesn't exist.";
  if (res.status === 429) return "Rate limited, try again in a bit.";
  try {
    const text = await res.text();
    if (text) {
      try {
        const body = JSON.parse(text);
        if (body.detail) return String(body.detail).slice(0, 500);
      } catch {
        /* not json */
      }
      return text.slice(0, 500);
    }
  } catch {
    // ignore
  }
  return `Request failed (${res.status})`;
}
