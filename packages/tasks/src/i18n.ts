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

export const TASKS_NAMESPACE = "@probo/tasks";

const catalogs = {
  "en-US": () => import("./_locales/en-US.json"),
  "es-ES": () => import("./_locales/es-ES.json"),
  "fr-FR": () => import("./_locales/fr-FR.json"),
  "nl-NL": () => import("./_locales/nl-NL.json"),
} as const;

export function registerTasksCatalogs(
  lookup: Map<string, () => Promise<{ default: unknown }>>,
  catalogKey: (namespace: string, language: string) => string,
) {
  for (const [language, load] of Object.entries(catalogs)) {
    lookup.set(catalogKey(TASKS_NAMESPACE, language), load);
  }
}
