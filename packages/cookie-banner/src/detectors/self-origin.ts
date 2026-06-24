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

const HTTP_URL_RE = /https?:\/\/[^\s)'"`]+/;
const LINE_COL_SUFFIX_RE = /:\d+(?::\d+)?$/;

function normalize(raw: string): string | null {
  const cleaned = raw.replace(LINE_COL_SUFFIX_RE, "");
  try {
    const parsed = new URL(cleaned);
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") return null;
    return parsed.origin + parsed.pathname;
  } catch {
    return null;
  }
}

// selfResourceUrls is computed once at module load. It holds the
// normalized `origin + pathname` URL(s) of the SDK bundle itself so the
// detectors never attribute a tracker -- or report the bundle as a
// resource -- to our own served script.
//
// We capture `document.currentScript?.src` here because that reference
// only resolves while the loading script is still executing; by the time
// a detector runs it is typically null. We also derive the executing URL
// from the load-time stack as a fallback for `.mjs`/bundler builds where
// `currentScript` is null. Exclusion is URL-level (origin + pathname),
// not origin-level, because the bundle is commonly served from a shared
// CDN (e.g. cdn.jsdelivr.net) that also hosts unrelated trackers.
const selfResourceUrls: ReadonlySet<string> = (() => {
  const urls = new Set<string>();

  if (
    typeof document !== "undefined"
    && document.currentScript instanceof HTMLScriptElement
    && document.currentScript.src
  ) {
    const fromSrc = normalize(document.currentScript.src);
    if (fromSrc) urls.add(fromSrc);
  }

  const stack = new Error().stack;
  if (stack) {
    const match = stack.match(HTTP_URL_RE);
    if (match) {
      const fromStack = normalize(match[0]);
      if (fromStack) urls.add(fromStack);
    }
  }

  return urls;
})();

// getSelfResourceUrls returns the normalized `origin + pathname` URL(s)
// of the SDK bundle itself. The set is empty when nothing is resolvable,
// in which case callers behave exactly as before.
export function getSelfResourceUrls(): ReadonlySet<string> {
  return selfResourceUrls;
}
