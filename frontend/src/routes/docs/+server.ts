import { ScalarApiReference } from "@scalar/sveltekit";
import { API_URL } from "$lib/env";
import type { RequestHandler } from "./$types";

// the site's colours and fonts, since scalar renders its own page. the doubled class
// has a higher specificity than scalar's theme, which loads after this
const customCss = `
@import url("https://fonts.googleapis.com/css2?family=Geist:wght@400;500;600&family=Geist+Mono:wght@400;500&display=swap");
.light-mode.light-mode, .dark-mode.dark-mode {
  --scalar-font: "Geist", ui-sans-serif, system-ui, sans-serif;
  --scalar-font-code: "Geist Mono", ui-monospace, monospace;
  --scalar-radius: 8px;
  --scalar-radius-lg: 12px;
}
.light-mode.light-mode {
  --scalar-background-1: oklch(0.972 0.008 85);
  --scalar-background-2: oklch(0.94 0.01 85);
  --scalar-background-3: oklch(0.88 0.008 70);
  --scalar-color-1: oklch(0.16 0.012 60);
  --scalar-color-2: oklch(0.46 0.008 60);
  --scalar-color-3: oklch(0.5 0.008 60);
  --scalar-color-orange: oklch(0.52 0.16 45);
  --scalar-color-accent: oklch(0.5 0.12 65);
  --scalar-background-accent: #ffb3001f;
  --scalar-border-color: oklch(0.88 0.008 70);
}
.dark-mode.dark-mode {
  --scalar-background-1: oklch(0.13 0.006 60);
  --scalar-background-2: oklch(0.165 0.006 60);
  --scalar-background-3: oklch(0.2 0.006 60);
  --scalar-color-1: oklch(0.96 0.005 60);
  --scalar-color-2: oklch(0.68 0.01 60);
  --scalar-color-3: oklch(0.6 0.01 60);
  --scalar-color-accent: #ffb300;
  --scalar-background-accent: #ffb3001f;
  --scalar-border-color: oklch(0.26 0.006 60);
}
`;

export const GET: RequestHandler = ({ url }) =>
  ScalarApiReference({
    url: `${API_URL}/openapi.json`,
    servers: [{ url: API_URL }],
    pageTitle: `Docs - ${url.hostname}`,
    hiddenClients: true,
    defaultOpenAllTags: true,
    agent: { disabled: true },
    layout: "classic",
    theme: "none",
    withDefaultFonts: false,
    customCss,
  })();
