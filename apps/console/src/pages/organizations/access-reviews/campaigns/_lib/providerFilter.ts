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

import { connectorProviderAccessibleName } from "#/pages/organizations/settings/integrations/_lib/connectorProviderLabel";

export type ProviderFilterOption = {
  value: string;
  label: string;
};

// One option per provider. A source with no connector (a CSV import, or a
// source deleted after the campaign snapshot) has no provider.
export function providerFilterOptions(
  sources: ReadonlyArray<{ provider: string | null; displayName: string | null }>,
): ProviderFilterOption[] {
  const labels = new Map<string, string>();
  for (const source of sources) {
    if (
      source.provider == null
      || source.displayName == null
      || labels.has(source.provider)
    ) {
      continue;
    }
    labels.set(
      source.provider,
      connectorProviderAccessibleName(source.provider, source.displayName),
    );
  }

  return [...labels.entries()]
    .map(([value, label]) => ({ value, label }))
    .sort((a, b) => a.label.localeCompare(b.label));
}

// An empty selection means the filter is off. A source with no provider is
// hidden while any provider is selected.
export function isSourceExcluded(
  filters: { providers: ReadonlyArray<string> },
  source: { provider: string | null },
): boolean {
  return filters.providers.length > 0
    && (source.provider == null || !filters.providers.includes(source.provider));
}
