export function debounce<Args extends unknown[], T>(
  fn: (signal: AbortSignal, ...args: Args) => Promise<T>,
  delayMs: number,
): { run: (...args: Args) => Promise<T>; abort: () => void } {
  let timeout: ReturnType<typeof setTimeout> | null = null;
  let controller: AbortController | null = null;
  let pendingReject: ((reason: unknown) => void) | null = null;

  function abort() {
    if (timeout) {
      clearTimeout(timeout);
      timeout = null;
    }
    controller?.abort();
    controller = null;
    if (pendingReject) {
      pendingReject(new DOMException("Aborted", "AbortError"));
      pendingReject = null;
    }
  }

  function run(...args: Args): Promise<T> {
    abort();
    return new Promise((resolve, reject) => {
      pendingReject = reject;
      timeout = setTimeout(async () => {
        controller = new AbortController();
        try {
          resolve(await fn(controller.signal, ...args));
        } catch (err) {
          reject(err);
        } finally {
          controller = null;
          timeout = null;
          pendingReject = null;
        }
      }, delayMs);
    });
  }

  return { run, abort };
}
