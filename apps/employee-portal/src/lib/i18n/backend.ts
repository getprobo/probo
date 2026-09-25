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

import type { BackendModule, ReadCallback, ResourceKey } from "i18next";

// The default namespace for app-wide chrome strings, sourced from src/_locales/.
export const DEFAULT_NAMESPACE = "app";

type CatalogModule = { default: ResourceKey };

// Vite turns each translation JSON into its own lazily-imported chunk. The keys
// are project-root-absolute paths, e.g.
// "/src/pages/organizations/measures/_locales/en-US.json".
const catalogs = import.meta.glob<CatalogModule>("/src/**/_locales/*.json");

const CATALOG_PATH = /^\/src\/(.*)_locales\/([^/]+)\.json$/;

// Map "<namespace>\u0000<language>" -> lazy importer for that catalog chunk.
function buildLookup(): Map<string, () => Promise<CatalogModule>> {
  const lookup = new Map<string, () => Promise<CatalogModule>>();

  for (const [path, importer] of Object.entries(catalogs)) {
    const match = CATALOG_PATH.exec(path);
    if (!match) {
      continue;
    }

    const [, prefix, language] = match;
    // prefix is the path between "src/" and "_locales/", e.g. "pages/foo/".
    // Drop the leading "pages/" and trailing slash; an empty prefix (src/_locales)
    // is the app-wide default namespace.
    const namespace
      = prefix.replace(/^pages\//, "").replace(/\/$/, "") || DEFAULT_NAMESPACE;

    lookup.set(catalogKey(namespace, language), importer);
  }

  return lookup;
}

function catalogKey(namespace: string, language: string): string {
  return `${namespace}\u0000${language}`;
}

const lookup = buildLookup();

// Custom i18next backend that resolves a (language, namespace) pair to its lazy
// JSON chunk. A missing catalog resolves to an empty resource so i18next falls
// through to fallbackLng rather than treating it as a hard load error.
export const globBackend: BackendModule = {
  type: "backend",
  init() {},
  read(language: string, namespace: string, callback: ReadCallback) {
    const importer = lookup.get(catalogKey(namespace, language));

    if (!importer) {
      callback(null, {});
      return;
    }

    importer().then(
      module => callback(null, module.default),
      (error: unknown) =>
        callback(error instanceof Error ? error : new Error(String(error)), null),
    );
  },
};
