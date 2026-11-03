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

import { matchRoutes, type RouteObject } from "react-router";

// A URL that yields no route segments is reported unchanged.
export function routePattern(
  routes: RouteObject[],
  pathname: string,
  basename?: string,
): string {
  const matches = matchRoutes(routes, pathname, basename);
  if (matches === null) {
    return pathname;
  }

  let pattern = "";
  for (const match of matches) {
    if (match.route.index) {
      continue;
    }
    const path = match.route.path;
    if (path === undefined || path === "") {
      continue;
    }
    pattern = path.startsWith("/") ? path : `${pattern}/${path}`;
  }

  if (pattern === "") {
    return pathname;
  }

  const normalized = pattern.replaceAll(/\/{2,}/g, "/");
  if (normalized.length > 1 && normalized.endsWith("/")) {
    return normalized.slice(0, -1);
  }
  return normalized;
}
