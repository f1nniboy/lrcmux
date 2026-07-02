import type { RequestHandler } from "@sveltejs/kit";

const routes = import.meta.glob("/src/routes/**/+page.svelte");

const pages = Object.keys(routes)
  .map((p) => p.replace("/src/routes", "").replace("/+page.svelte", "") || "/")
  .filter((p) => !p.includes("["));

export const GET: RequestHandler = ({ url }) => {
  const body = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${pages.map((p) => `  <url><loc>${url.origin}${p}</loc></url>`).join("\n")}
</urlset>`;

  return new Response(body, {
    headers: {
      "Content-Type": "application/xml",
      "Cache-Control": "public, max-age=3600",
    },
  });
};
