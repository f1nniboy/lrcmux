import type { SimpleIcon } from "simple-icons";
import { siGithubsponsors, siKofi, siLiberapay } from "simple-icons";

export type PlatformId = "github" | "liberapay" | "kofi";

export interface Platform {
  id: PlatformId;
  name: string;
  url: string;
  icon: SimpleIcon;
}

// monthly donations against the github sponsors goal, in us dollars
export interface Totals {
  raised: number;
  goal: number;
}

export const platforms: Platform[] = [
  {
    id: "github",
    name: "GitHub Sponsors",
    url: "https://github.com/sponsors/f1nniboy",
    icon: siGithubsponsors,
  },
  {
    id: "liberapay",
    name: "Liberapay",
    url: "https://liberapay.com/f1nniboy",
    icon: siLiberapay,
  },
  {
    id: "kofi",
    name: "Ko-fi",
    url: "https://ko-fi.com/f1nniboy",
    icon: siKofi,
  },
];
