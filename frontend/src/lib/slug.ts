export function toSlug(s: string): string {
  return [...s]
    .map((c) => {
      if (c === " ") return "-";
      if (c === "-") return "%2D";
      return encodeURIComponent(c);
    })
    .join("");
}

export function fromSlug(s: string): string {
  return decodeURIComponent(s.replaceAll("-", " "));
}
