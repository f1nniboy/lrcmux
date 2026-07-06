import type { LyricsFormat, SyncLevel } from "./types";
import { API_URL } from "./env";

interface GetLyricsOptions {
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
    throw new LyricsError(res.status, await failureMessage(res));
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
  return `${API_URL}/get?${params.toString()}`;
}

async function failureMessage(res: Response): Promise<string> {
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

