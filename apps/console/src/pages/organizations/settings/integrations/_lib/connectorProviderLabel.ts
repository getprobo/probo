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

// Spoken and searchable region for an OVHcloud connector. The console shows
// this as a flag beside the name, so the words stay out of the label.
const ovhcloudRegionLabels: Record<string, string> = {
  OVHCLOUD: "EU",
  OVHCLOUD_US: "US",
};

export function ovhcloudRegionLabel(provider: string): string | null {
  return ovhcloudRegionLabels[provider] ?? null;
}

// The card shows a flag for the region, so the visible name stays "OVHcloud"
// even when the provider label still carries EU or US.
export function connectorProviderVisibleName(provider: string, name: string): string {
  if (ovhcloudRegionLabel(provider) == null) {
    return name;
  }

  return name.replace(/ (EU|US)$/, "");
}

export function connectorProviderAccessibleName(provider: string, name: string): string {
  const region = ovhcloudRegionLabel(provider);
  if (region == null) {
    return name;
  }

  return `${name}, ${region}`;
}

export function connectorProviderMatchesSearch(
  provider: string,
  displayName: string,
  normalizedSearch: string,
): boolean {
  if (normalizedSearch === "") {
    return true;
  }

  const haystack = [
    displayName,
    provider.replaceAll("_", " "),
    ovhcloudRegionLabel(provider),
  ].filter(part => part != null && part !== "").join(" ").toLowerCase();

  return haystack.includes(normalizedSearch);
}
