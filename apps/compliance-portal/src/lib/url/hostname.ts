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

// Prepend https:// when a URL carries no http(s) scheme, so a protocol-less
// value (e.g. "blaxel.ai") parses as absolute instead of being treated as a
// relative path.
function withHttpScheme(url: string): string {
  return /^https?:\/\//i.test(url) ? url : `https://${url}`;
}

// Show only the hostname for a URL (e.g. "https://blaxel.ai/x" -> "blaxel.ai"),
// falling back to the raw value when it cannot be parsed.
export function hostnameOf(url: string): string {
  try {
    return new URL(withHttpScheme(url)).hostname;
  } catch {
    return url;
  }
}

// Build a safe absolute href for an external link, normalizing the scheme so a
// protocol-less value does not resolve as a relative link.
export function externalHref(url: string): string {
  return withHttpScheme(url);
}
