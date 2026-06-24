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

import type { ProviderInfo } from "../AddAccessReviewSourceDialog";

// DATADOG_SITES labels are technical identifiers (region code + hostname),
// intentionally not wrapped in __(). The dialog's prose strings are.
export const DATADOG_SITES: { value: string; label: string }[] = [
  { value: "US1", label: "US1 (app.datadoghq.com)" },
  { value: "US3", label: "US3 (us3.datadoghq.com)" },
  { value: "US5", label: "US5 (us5.datadoghq.com)" },
  { value: "EU1", label: "EU1 (app.datadoghq.eu)" },
  { value: "AP1", label: "AP1 (ap1.datadoghq.com)" },
  { value: "AP2", label: "AP2 (ap2.datadoghq.com)" },
  { value: "US1-FED", label: "US1-FED (app.ddog-gov.com)" },
];

export function mapAPIKeyExtraSettingToField(
  provider: string,
  settingKey: string,
): string | null {
  switch (provider) {
    case "TALLY":
      if (settingKey === "organizationId") return "tallyOrganizationId";
      break;
    case "SENTRY":
      if (settingKey === "organizationSlug") return "sentryOrganizationSlug";
      break;
    case "SUPABASE":
      if (settingKey === "organizationSlug") return "supabaseOrganizationSlug";
      break;
    case "GITHUB":
      if (settingKey === "organization") return "githubOrganization";
      break;
    case "GRAFANA":
      if (settingKey === "baseUrl") return "grafanaBaseUrl";
      break;
    case "SIGNOZ":
      if (settingKey === "baseUrl") return "signozBaseUrl";
      break;
    case "ONE_PASSWORD":
      if (settingKey === "scimBridgeUrl") return "onePasswordScimBridgeUrl";
      break;
    case "METABASE":
      if (settingKey === "instanceUrl") return "metabaseInstanceUrl";
      break;
    case "POSTHOG":
      if (settingKey === "region") return "posthogRegion";
      if (settingKey === "instanceUrl") return "posthogInstanceUrl";
      break;
    case "OKTA":
      if (settingKey === "domain") return "oktaDomain";
      break;
    case "BETTER_STACK":
      if (settingKey === "teamName") return "betterStackTeamName";
      break;
    case "QOVERY":
      if (settingKey === "organizationId") return "qoveryOrganizationId";
      break;
    case "RENDER":
      if (settingKey === "workspaceId") return "renderWorkspaceId";
      break;
    case "NEON":
      if (settingKey === "organizationId") return "neonOrganizationId";
      break;
    case "SCALEWAY":
      if (settingKey === "organizationId") return "scalewayOrganizationId";
      break;
    case "CRISP":
      if (settingKey === "websiteId") return "crispWebsiteId";
      break;
  }
  return null;
}

export function mapClientCredentialsExtraSettingToField(
  provider: string,
  settingKey: string,
): string | null {
  switch (provider) {
    case "ONE_PASSWORD":
      if (settingKey === "accountId") return "onePasswordAccountId";
      if (settingKey === "region") return "onePasswordRegion";
      break;
  }
  return null;
}

export function hasRequiredExtraSettings(
  settings: ReadonlyArray<{ readonly key: string; readonly required: boolean }>,
  values: Record<string, string>,
): boolean {
  return settings
    .filter(s => s.required)
    .every(s => values[s.key]?.trim());
}

// buildExtraFields flattens a provider's extra settings into the input-field map
// the create mutations expect: each non-empty, trimmed value keyed by its
// provider-specific input field name (via mapFn), skipping settings that map to
// nothing. Shared by the API-key and client-credentials dialogs.
export function buildExtraFields(
  provider: ProviderInfo,
  values: Record<string, string>,
  mapFn: (provider: string, settingKey: string) => string | null,
): Record<string, string> {
  const extraFields: Record<string, string> = {};
  for (const setting of provider.extraSettings) {
    const value = values[setting.key]?.trim();
    if (!value) {
      continue;
    }
    const fieldName = mapFn(provider.provider, setting.key);
    if (fieldName) {
      extraFields[fieldName] = value;
    }
  }
  return extraFields;
}

// Accepts either a bare subdomain ("acme") or a pasted host
// ("https://acme.zendesk.com/") and reduces it to the bare subdomain the
// backend expects as the `site` query param.
export function cleanZendeskSubdomain(raw: string): string {
  let value = raw.trim();
  value = value.replace(/^https?:\/\//i, "");
  // Drop any path, query, or fragment from a pasted URL/host so only the
  // host label survives (e.g. "acme.zendesk.com/agent?x=1" -> "acme").
  value = value.replace(/[/?#].*$/, "");
  value = value.replace(/\.zendesk\.com$/i, "");
  return value.trim();
}

// connectOAuthProvider builds the connector-initiate URL for an OAuth provider
// and navigates the browser to it, kicking off the authorization redirect.
export function connectOAuthProvider(
  organizationId: string,
  info: ProviderInfo,
  extras?: Record<string, string>,
) {
  const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
  const url = new URL("/api/console/v1/connectors/initiate", baseURL);
  url.searchParams.append("organization_id", organizationId);
  url.searchParams.append("provider", info.provider);
  for (const scope of info.oauth2Scopes) {
    url.searchParams.append("scope", scope);
  }
  if (extras) {
    for (const [k, v] of Object.entries(extras)) {
      url.searchParams.append(k, v);
    }
  }
  url.searchParams.append(
    "continue",
    `/organizations/${organizationId}/access-reviews/sources`,
  );
  window.location.assign(url.toString());
}
