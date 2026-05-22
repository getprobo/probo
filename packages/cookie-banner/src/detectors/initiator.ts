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

const EXTENSION_URL_RE = /(?:chrome|moz|safari-web)-extension:\/\//;
const STACK_URL_RE = /https?:\/\/[^\s)'"`]+/g;
const LINE_COL_SUFFIX_RE = /:\d+(?::\d+)?$/;
const MAX_INITIATOR_URL_LENGTH = 1024;

export interface InitiatorContext {
  url: string | null;
  fromExtension: boolean;
}

// getInitiatorURL walks the current call stack once and returns:
//   - `url`: the first third-party HTTP(S) script URL it finds
//     (as origin+pathname), skipping the SDK's API origin and the
//     page's own origin. Returns null when no such frame exists.
//   - `fromExtension`: true if any frame on the stack was a
//     browser-extension URL (chrome/moz/safari-web-extension://).
//     Page-world extensions (MV3 main world, userscripts with
//     @grant none) reliably leave such a frame; isolated-world
//     content scripts use a different realm and never reach this
//     function in the first place.
//
// Both signals come from the same single stack walk, so callers
// that need either or both pay only one `new Error().stack` cost.
export function getInitiatorURL(apiOrigin: string): InitiatorContext {
  const stack = new Error().stack;
  if (!stack) return { url: null, fromExtension: false };

  let fromExtension = false;
  let url: string | null = null;

  STACK_URL_RE.lastIndex = 0;
  let m: RegExpExecArray | null;
  while ((m = STACK_URL_RE.exec(stack)) !== null) {
    const raw = m[0];
    if (EXTENSION_URL_RE.test(raw)) {
      fromExtension = true;
      continue;
    }

    if (url !== null) continue;

    const cleaned = raw.replace(LINE_COL_SUFFIX_RE, "");

    let parsed: URL;
    try {
      parsed = new URL(cleaned);
    } catch {
      continue;
    }

    if (parsed.origin === apiOrigin) continue;
    if (parsed.origin === location.origin) continue;

    const result = parsed.origin + parsed.pathname;
    url = result.length > MAX_INITIATOR_URL_LENGTH
      ? result.slice(0, MAX_INITIATOR_URL_LENGTH)
      : result;
  }

  return { url, fromExtension };
}
