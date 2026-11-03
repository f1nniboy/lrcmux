import { API_URL } from "$lib/env";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = () => {
  const body = `# lrcmux

> A lyrics API that aggregates every public provider.

The name is written as "lrcmux" (all lowercase, no space).

The full reference is the OpenAPI spec at ${API_URL}/openapi.json. It covers every endpoint, parameter, response format and schema.

No API key is needed. You must send a \`User-Agent\` that names your app or a way to contact you (e-mail, etc.).

- [OpenAPI spec](${API_URL}/openapi.json)
- [Sources](${API_URL}/sources): the providers lrcmux searches
- [Source code](https://github.com/f1nniboy/lrcmux)
`;

  return new Response(body, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Cache-Control": "public, max-age=86400",
    },
  });
};
