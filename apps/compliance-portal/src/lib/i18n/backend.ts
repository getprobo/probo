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
