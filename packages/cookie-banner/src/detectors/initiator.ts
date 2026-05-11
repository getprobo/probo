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

// getInitiatorURL walks the current call stack and returns the first
// third-party script URL (as origin+pathname). It deliberately skips:
//   - browser extension URLs (chrome/moz/safari-web-extension://)
//   - the SDK's API origin (instrumentation frames)
//   - the page's own origin (we want the third-party loader, not first-party code)
//
// Returns null when no third-party frame is found (anonymous/eval/inline
// scripts, or writes originating from the page itself).
export function getInitiatorURL(apiOrigin: string): string | null {
  const stack = new Error().stack;
  if (!stack) return null;

  STACK_URL_RE.lastIndex = 0;
  let m: RegExpExecArray | null;
  while ((m = STACK_URL_RE.exec(stack)) !== null) {
    const raw = m[0];
    if (EXTENSION_URL_RE.test(raw)) continue;

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
    if (result.length > MAX_INITIATOR_URL_LENGTH) {
      return result.slice(0, MAX_INITIATOR_URL_LENGTH);
    }
    return result;
  }
  return null;
}
