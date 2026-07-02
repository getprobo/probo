// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

const EXTENSION_URL_RE = /(?:chrome|moz|safari-web)-extension:\/\//;

const EXTENSION_PROTOCOLS = new Set([
  "chrome-extension:",
  "moz-extension:",
  "safari-web-extension:",
]);

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
