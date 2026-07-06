import type { HandleFetch } from "@sveltejs/kit";

export const handleFetch: HandleFetch = ({ event, request, fetch }) => {
  const userIP = event.request.headers.get("CF-Connecting-IP");
  if (userIP) request.headers.set("X-Real-IP", userIP);

  return fetch(request);
};
