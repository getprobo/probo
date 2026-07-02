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

import { getSelfResourceUrls } from "./self-origin";

const EXTENSION_URL_RE = /(?:chrome|moz|safari-web)-extension:\/\//;
const STACK_URL_RE =
  /(?:https?|(?:chrome|moz|safari-web)-extension):\/\/[^\s)'"`]+/g;
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

  const selfUrls = getSelfResourceUrls();

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
    if (selfUrls.has(result)) continue;

    url = result.length > MAX_INITIATOR_URL_LENGTH
      ? result.slice(0, MAX_INITIATOR_URL_LENGTH)
      : result;
  }

  return { url, fromExtension };
}
