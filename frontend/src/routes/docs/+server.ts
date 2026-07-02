import { ScalarApiReference } from "@scalar/sveltekit";
import { API_URL } from "$lib/env";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = () =>
  ScalarApiReference({
    url: `${API_URL}/openapi.json`,
    servers: [{ url: API_URL }],
    hiddenClients: true,
    defaultOpenAllTags: true,
    agent: { disabled: true },
    layout: "classic",
  })();
