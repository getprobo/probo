// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

export const EXTENSION_URL_RE = /(?:chrome|moz|safari-web)-extension:\/\//;

const EXTENSION_PROTOCOLS = new Set([
  "chrome-extension:",
  "moz-extension:",
  "safari-web-extension:",
]);

// isExtensionCaller inspects the synchronous JS call stack for a frame
// whose URL is a browser extension origin. It is a best-effort check:
// some hooks (PerformanceObserver, browser internals firing events from
// their own threads) will not carry the original injector frame, so a
// `false` result does not guarantee the caller is first-party.
export function isExtensionCaller(): boolean {
  const stack = new Error().stack ?? "";
  return EXTENSION_URL_RE.test(stack);
}

// extensionContext is evaluated once at module load. We capture
// `document.currentScript?.src` here because that reference only
// resolves while the loading script is still executing -- by the time
// a detector's `start()` is called, currentScript will typically be
// null. Computing this lazily would defeat the check.
const extensionContext: boolean = (() => {
  if (typeof location !== "undefined" && EXTENSION_PROTOCOLS.has(location.protocol)) {
    return true;
  }
  if (typeof document !== "undefined") {
    const src = document.currentScript instanceof HTMLScriptElement
      ? document.currentScript.src
      : null;
    if (src && EXTENSION_URL_RE.test(src)) return true;
  }
  return false;
})();

// isExtensionContext reports whether the SDK itself is being executed
// from inside a browser extension (either an extension page, or a
// script loaded via an extension URL). Detectors use this to skip
// pre-existing-state scans that have no caller stack to inspect and
// would otherwise attribute the extension's own data to the host page.
export function isExtensionContext(): boolean {
  return extensionContext;
}
