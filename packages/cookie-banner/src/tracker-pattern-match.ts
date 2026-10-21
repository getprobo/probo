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

import type { Category } from "./types";

// globMatch follows the server's single-star tracker patterns: each * is one
// wildcard, and a name without * is exact.
function globMatch(pattern: string, name: string): boolean {
  const parts = pattern.split("*");
  if (parts.length === 1) {
    return pattern === name;
  }

  const [prefix, ...rest] = parts;
  if (prefix != null && !name.startsWith(prefix)) {
    return false;
  }

  let restName = name.slice(prefix?.length ?? 0);
  const suffix = rest[rest.length - 1] ?? "";
  if (!restName.endsWith(suffix)) {
    return false;
  }

  restName = restName.slice(0, restName.length - suffix.length);

  for (const part of rest.slice(0, -1)) {
    const idx = restName.indexOf(part);
    if (idx === -1) {
      return false;
    }

    restName = restName.slice(idx + part.length);
  }

  return true;
}

function matchesKind(trackerType: string, kind: "cookie" | "storage"): boolean {
  if (kind === "cookie") {
    return trackerType === "COOKIE";
  }

  return trackerType !== "COOKIE";
}

function fixedLength(pattern: string): number {
  return pattern.split("*").join("").length;
}

// selectTrackerCategory picks the published pattern for a detected name.
// An exact name wins. Otherwise the matching * glob with the longest fixed
// text wins, matching FindMatchingPattern.
export function selectTrackerCategory(
  categories: readonly Category[],
  kind: "cookie" | "storage",
  name: string,
): Category | undefined {
  let best: { category: Category; exact: boolean; fixed: number } | undefined;

  for (const category of categories) {
    for (const cookie of category.cookies) {
      if (!matchesKind(cookie.tracker_type, kind)) {
        continue;
      }

      const exact = cookie.name === name;
      if (!exact && !globMatch(cookie.name, name)) {
        continue;
      }

      const candidate = { category, exact, fixed: fixedLength(cookie.name) };
      if (
        best == null
        || (candidate.exact && !best.exact)
        || (candidate.exact === best.exact && candidate.fixed > best.fixed)
      ) {
        best = candidate;
      }
    }
  }

  return best?.category;
}
