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
