export interface Guide {
  id: string;
  name: string;
  icon: string;
  url: string;
  // apps that use another lyrics API, pointed at lrcmux's endpoint for it
  compat?: string;
  steps?: string[];
}

export const guides: Guide[] = [
  {
    id: "beets",
    name: "Beets",
    icon: "https://beets.readthedocs.io/en/stable/_static/beets_logo_nobg.png",
    url: "https://beets.io",
    steps: [
      "Add <strong>lyrics</strong> to the <strong>plugins</strong> list in your configuration.",
      "Run <strong>beet lyrics</strong> to fetch synced lyrics for your library.",
    ],
  },
  {
    id: "navidrome",
    name: "Navidrome",
    icon: "https://raw.githubusercontent.com/navidrome/navidrome/master/resources/logo-192x192.png",
    url: "https://www.navidrome.org",
    steps: [
      'Install the <a href="https://github.com/J0R6IT0/navidrome-lyrics-plugin" target="_blank" rel="noopener" class="link">navidrome-lyrics-plugin</a> if you haven\'t already.',
      "Click your <strong>account icon</strong> in the top right and open <strong>Plugins</strong>.",
      "Click on the <strong>nd-lyrics</strong> plugin, then scroll down to <strong>Lyrics providers</strong>.",
      "Press <strong>+</strong>, click the new slot, and select <strong>lrcmux</strong> as the provider.",
      "Click <strong>Save</strong>.",
    ],
  },
  {
    id: "chorus",
    name: "Chorus",
    icon: "https://raw.githubusercontent.com/f1nniboy/chorus/refs/heads/main/data/icons/hicolor/scalable/apps/space.f1nn.chorus.svg",
    url: "https://github.com/f1nniboy/chorus",
  },
  {
    id: "youly",
    name: "YouLy+",
    icon: "https://raw.githubusercontent.com/ibratabian17/YouLyPlus/refs/heads/main/icons/icon128.png",
    url: "https://github.com/ibratabian17/YouLyPlus",
    compat: "/compat/kpoe",
    steps: [
      "Click the YouLy+ icon in your browser toolbar and open <strong>More Settings</strong>.",
      "Under <strong>Sources</strong>, remove all other providers and add <strong>Custom KPoe Server</strong>.",
      "Set <strong>Custom KPoe Server URL</strong> to the endpoint URL above.",
    ],
  },
  {
    id: "lrcget",
    name: "LRCGET",
    icon: "https://raw.githubusercontent.com/tranxuanthang/lrcget/refs/heads/main/app-icon.png",
    url: "https://github.com/tranxuanthang/lrcget",
    compat: "/compat/lrclib",
    steps: [
      "Open LRCGET and click the <strong>three dots</strong> menu, then open <strong>Settings</strong>.",
      "Under <strong>LRCLIB instance</strong>, replace the default URL with the endpoint URL above.",
      "Click <strong>Save</strong>.",
    ],
  },
];
