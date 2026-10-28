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

// Names that do not normalize to the connector display name.
const connectorProviderAliases: Record<string, readonly string[]> = {
  AWS: ["Amazon Web Services", "AWS"],
  GCP: ["Google Cloud", "Google Cloud Platform", "GCP"],
  AZURE: ["Microsoft Azure", "Azure"],
  ONE_PASSWORD: ["1Password", "OnePassword"],
  MONDAY: ["Monday.com", "Monday"],
  APOLLO: ["Apollo.io", "Apollo"],
  CLICKHOUSE: ["ClickHouse Cloud", "ClickHouse"],
  INCIDENT_IO: ["incident.io"],
  CAL_COM: ["Cal.com", "Calcom"],
  BETTER_STACK: ["Better Stack", "BetterStack"],
  MONGODB_ATLAS: ["MongoDB Atlas", "MongoDB"],
  ELASTIC_CLOUD: ["Elastic Cloud", "Elastic", "Elasticsearch"],
  OVHCLOUD: ["OVHcloud", "OVH"],
  MICROSOFT_365: ["Microsoft 365", "Office 365"],
  GOOGLE_WORKSPACE: ["Google Workspace"],
};

export interface ConnectorProviderName {
  provider: string;
  displayName: string;
}

export function normalizeThirdPartyName(value: string): string {
  return value
    .toLowerCase()
    .replaceAll("&", " and ")
    .replaceAll(/[^a-z0-9]+/g, " ")
    .trim()
    .replaceAll(/ +/g, " ");
}

export function matchingConnectorProviders(
  providers: readonly ConnectorProviderName[],
  thirdPartyNames: readonly string[],
): Set<string> {
  const names = new Set(thirdPartyNames.map(normalizeThirdPartyName));
  const matched = new Set<string>();

  for (const provider of providers) {
    const candidates = [
      provider.displayName,
      ...(connectorProviderAliases[provider.provider] ?? []),
    ];
    const matches = candidates.some(candidate => names.has(normalizeThirdPartyName(candidate)));

    if (matches) {
      matched.add(provider.provider);
    }
  }

  return matched;
}
