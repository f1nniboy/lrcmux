export type SyncLevel = "word" | "line" | "none";

export interface Word {
  start: number;
  end: number;
  text: string;
}

export interface Line {
  start?: number;
  end?: number;
  text: string;
  words?: Word[];
}

export interface LyricsSource {
  id: string;
  name: string;
}

export interface LyricsMeta {
  source?: LyricsSource;
  level: SyncLevel;
}

export interface LyricsResult {
  meta: LyricsMeta;
  lines: Line[];
  track: Track;
}

export interface TrackCover {
  small?: string;
  medium?: string;
  big?: string;
}

export interface Track {
  isrc: string;
  title: string;
  duration: number;
  artist: string;
  album: string;
  cover: TrackCover;
}

export type LyricsFormat = "lrc" | "txt" | "json" | "srt" | "vtt";

export interface FormatDef {
  id: LyricsFormat;
  rich?: boolean;
  minLevel?: SyncLevel;
}

export const FORMATS: FormatDef[] = [
  { id: "txt", rich: true },
  { id: "lrc" },
  { id: "srt", minLevel: "line" },
  { id: "vtt", minLevel: "line" },
  { id: "json" },
];

export const LEVELS: SyncLevel[] = ["word", "line", "none"];

export const LEVEL_RANK: Record<SyncLevel, number> = {
  word: 2,
  line: 1,
  none: 0,
};

export interface DeezerArtist {
  id: number;
  name: string;
  picture_small?: string;
  picture_medium?: string;
  picture_big?: string;
}

export interface DeezerAlbum {
  id: number;
  title: string;
  cover_small?: string;
  cover_medium?: string;
  cover_big?: string;
}

export interface DeezerTrack {
  id: number;
  title: string;
  title_short: string;
  isrc: string;
  duration: number;
  preview?: string;
  explicit_lyrics?: boolean;
  type?: string;
  readable?: boolean;
  artist: DeezerArtist;
  album: DeezerAlbum;
}

export interface DeezerSearchResponse {
  data: DeezerTrack[];
  total: number;
  next?: string;
}
