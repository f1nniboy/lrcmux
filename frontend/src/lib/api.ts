import type { LyricsFormat, LyricsResult, SyncLevel } from "./types";

export interface GetLyricsOptions {
  level?: SyncLevel;
  signal?: AbortSignal;
}

export class LyricsError extends Error {
  constructor(
    public status: number,
    message: string,
    public retryAfter = 0,
  ) {
    super(message);
    this.name = "LyricsError";
  }
}

export async function getLyricsJSON(
  artist: string,
  title: string,
  opts: Omit<GetLyricsOptions, "format"> = {},
): Promise<LyricsResult> {
  const res = await fetch(downloadURL(artist, title, "json", opts), {
    signal: opts.signal,
  });
  if (!res.ok) {
    throw new LyricsError(
      res.status,
      await failureMessage(res),
      retryAfterSeconds(res),
    );
  }
  return (await res.json()) as LyricsResult;
}

export async function getLyricsText(
  artist: string,
  title: string,
  format: LyricsFormat,
  opts: Omit<GetLyricsOptions, "format"> = {},
): Promise<string> {
  const res = await fetch(downloadURL(artist, title, format, opts), {
    signal: opts.signal,
  });
  if (!res.ok) {
    throw new LyricsError(
      res.status,
      await failureMessage(res),
      retryAfterSeconds(res),
    );
  }
  return res.text();
}

export function downloadURL(
  artist: string,
  title: string,
  format: LyricsFormat,
  opts: GetLyricsOptions = {},
): string {
  const params = new URLSearchParams();
  params.set("artist", artist);
  params.set("title", title);
  params.set("format", format);
  if (opts.level) params.set("level", opts.level);
  return `/api/get?${params.toString()}`;
}

async function failureMessage(res: Response): Promise<string> {
  if (res.status === 404)
    return "No provider has this track, or it doesn't exist.";
  try {
    const text = await res.text();
    if (text) {
      try {
        const body = JSON.parse(text);
        if (body.detail) return String(body.detail).slice(0, 500);
      } catch {
        /* not JSON */
      }
      return text.slice(0, 500);
    }
  } catch {
    // ignore
  }
  return `Request failed (${res.status})`;
}

function retryAfterSeconds(res: Response): number {
  return parseInt(res.headers.get("Retry-After") ?? "0", 10) || 0;
}
