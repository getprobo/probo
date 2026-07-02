// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
