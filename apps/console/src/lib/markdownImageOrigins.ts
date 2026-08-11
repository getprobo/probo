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

function trimViteEnv(key: string): string {
  // import.meta.env is loosely typed (and vite `define` injects these keys as
  // any); re-read through Record<string, unknown> before narrowing.
  const env: Record<string, unknown> = import.meta.env;
  const value = env[key];
  return typeof value === "string" ? value.trim() : "";
}

/**
 * Origins permitted for Markdown <img> src, matching console CSP img-src peers
 * beyond 'self' and data: (see apps/console/content-security-policy.txt.tmpl).
 */
export function consoleMarkdownImageOrigins(): string[] {
  const origins = new Set<string>();

  const appOrigin = trimViteEnv("VITE_APP_ORIGIN");
  if (appOrigin !== "") {
    origins.add(appOrigin.replace(/\/$/, ""));
  } else {
    const apiURL = trimViteEnv("VITE_API_URL");
    if (apiURL !== "") {
      try {
        const formatted
          = apiURL.startsWith("http://") || apiURL.startsWith("https://")
            ? apiURL
            : `https://${apiURL}`;
        origins.add(new URL(formatted).origin);
      } catch {
        // ignore malformed VITE_API_URL
      }
    }
  }

  const storageOrigin = trimViteEnv("VITE_FILE_STORAGE_ORIGIN");
  if (storageOrigin !== "") {
    origins.add(storageOrigin.replace(/\/$/, ""));
  }

  origins.add("https://www.google.com");

  return [...origins];
}
