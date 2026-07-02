import type { RequestHandler } from "@sveltejs/kit";

export const POST: RequestHandler = async ({ request }) => {
  const res = await fetch("https://plausible.io/api/event", {
    method: "POST",
    headers: request.headers,
    body: request.body,
    duplex: "half",
  } as RequestInit);
  return new Response(res.body, { status: res.status });
};
