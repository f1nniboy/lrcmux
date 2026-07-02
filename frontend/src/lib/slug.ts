export function toSlug(s: string): string {
  return encodeURIComponent(s.replaceAll(" ", "-"));
}

export function fromSlug(s: string): string {
  return decodeURIComponent(s).replaceAll("-", " ");
}
