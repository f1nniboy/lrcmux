import { env } from "$env/dynamic/private";
import type { RequestHandler } from "@sveltejs/kit";

export const GET: RequestHandler = async () => {
  const key = env.PLAUSIBLE_KEY;
  if (!key) return new Response(null, { status: 404 });

  const res = await fetch(`https://plausible.io/js/${key}.js`);
  const body = await res.arrayBuffer();
  return new Response(body, {
    headers: {
      "Content-Type": "application/javascript",
      "Cache-Control": "public, max-age=86400",
    },
  });
};
