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

import { loadQuery, type PreloadedQuery } from "react-relay";
import type { OperationType } from "relay-runtime";

import { stripLocaleFromPathname } from "#/lib/i18n/locale";
import { environment } from "#/lib/relay/environment";
import HomePageQueryNode from "#/pages/__generated__/HomePageQuery.graphql";
import MainLayoutQueryNode from "#/pages/__generated__/MainLayoutQuery.graphql";

// Long enough for the route loaders to mount and retain the same operations
// through useQueryLoader; after that the early references are redundant.
const PRELOAD_RETAIN_MS = 30_000;

// Routes under /:lang that render outside MainLayout.
const STANDALONE_PATHS = new Set(["/full-name", "/nda"]);

// Without this, a first visit loads the i18n catalog, the MainLayout chunk,
// MainLayoutQuery, the HomePage chunk and HomePageQuery one after another,
// because each step only starts once the previous one rendered. Starting them
// all at boot makes them download in parallel. The lazy() routes reuse the
// in-flight module imports, and Relay dedupes the in-flight requests issued
// later by the *Loader components, so routes and loaders stay unchanged.
export function preloadInitialRoute(pathname: string): void {
  const appPath = stripLocaleFromPathname(pathname).replace(/\/+$/, "") || "/";

  if (STANDALONE_PATHS.has(appPath)) {
    return;
  }

  const preloaded: PreloadedQuery<OperationType>[] = [];

  // Failures are ignored here: lazy() retries the same import on render.
  import("#/pages/MainLayoutLoader").catch(() => {});
  preloaded.push(loadQuery(environment, MainLayoutQueryNode, {}));

  if (appPath === "/") {
    import("#/pages/HomePageLoader").catch(() => {});
    preloaded.push(loadQuery(environment, HomePageQueryNode, {}));
  }

  setTimeout(() => {
    for (const queryRef of preloaded) {
      queryRef.dispose();
    }
  }, PRELOAD_RETAIN_MS);
}
