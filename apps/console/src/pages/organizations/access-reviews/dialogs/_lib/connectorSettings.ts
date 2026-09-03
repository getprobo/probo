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

import type { ConnectorProtocol } from "#/__generated__/core/connectorProviderInfoFields_installableProtocols.graphql";

// DATADOG_SITES labels are technical identifiers (region code + hostname),
// intentionally not translated. The dialog's prose strings are.
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
    case "LANGFUSE":
      if (settingKey === "baseUrl") return "langfuseBaseUrl";
      break;
    case "AUTHENTIK":
      if (settingKey === "baseUrl") return "authentikBaseUrl";
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
    case "SEGMENT":
      if (settingKey === "region") return "segmentRegion";
      break;
    case "CRISP":
      if (settingKey === "websiteId") return "crispWebsiteId";
      break;
  }
  return null;
}

// Same grammar as pkg/awsx/arn.RoleARNPattern, with the three supported
// partitions inlined so the field rejects other partitions immediately.
export const AWS_IAM_ROLE_ARN_PATTERN
  = "arn:(aws-us-gov|aws-cn|aws):iam::([0-9]{12}):role(?:/[\\w+=,.@\\-]+)*/[\\w+=,.@\\-]{1,64}";

const awsIAMRoleARN = new RegExp(`^${AWS_IAM_ROLE_ARN_PATTERN}$`);

export function isAWSRoleARN(value: string): boolean {
  return awsIAMRoleARN.test(value.trim());
}

export function awsAccountIDFromRoleARN(value: string): string | null {
  const match = value.trim().match(awsIAMRoleARN);
  if (!match) {
    return null;
  }

  return match[2] ?? null;
}

// Immediate name while the worker assumes the role and replaces the
// account ID with the official account name (or the sign-in alias).
export function awsAccessReviewSourceName(
  displayName: string,
  roleArn: string,
): string {
  const accountID = awsAccountIDFromRoleARN(roleArn);
  if (!accountID) {
    return displayName;
  }

  return `${displayName} / ${accountID}`;
}

const GCP_PROVIDER_RESOURCE_PATTERN
  = /^(?:https:\/\/iam\.googleapis\.com\/|\/\/iam\.googleapis\.com\/)?projects\/([1-9][0-9]*)\/locations\/global\/workloadIdentityPools\/([a-z][a-z0-9-]{3,31})\/providers\/([a-z][a-z0-9-]{3,31})\/?$/;

const GCP_SERVICE_ACCOUNT_EMAIL_PATTERN
  = /^[a-z][a-z0-9-]{4,28}[a-z0-9]@[a-z][a-z0-9-]{4,28}[a-z0-9]\.iam\.gserviceaccount\.com$/;

export function isGCPWorkloadIdentityProvider(value: string): boolean {
  return GCP_PROVIDER_RESOURCE_PATTERN.test(value.trim());
}

export function isGCPServiceAccountEmail(value: string): boolean {
  return GCP_SERVICE_ACCOUNT_EMAIL_PATTERN.test(value.trim());
}

export function gcpProjectNumberFromProvider(value: string): string | null {
  const match = value.trim().match(GCP_PROVIDER_RESOURCE_PATTERN);
  if (!match) {
    return null;
  }

  return match[1] ?? null;
}

export function gcpAccessReviewSourceName(
  displayName: string,
  providerResource: string,
): string {
  const projectNumber = gcpProjectNumberFromProvider(providerResource);
  if (!projectNumber) {
    return displayName;
  }

  return `${displayName} / ${projectNumber}`;
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

// buildExtraFields flattens one connect path's extra settings into the
// input-field map that path's create mutation expects: each non-empty, trimmed
// value keyed by its provider-specific input field name (via mapFn), skipping
// settings that map to nothing. Each dialog passes the settings list for its own
// path together with the matching mapFn — a provider offering both paths
// (1Password) declares different settings on each.
export function buildExtraFields(
  provider: string,
  settings: ReadonlyArray<{ readonly key: string }>,
  values: Record<string, string>,
  mapFn: (provider: string, settingKey: string) => string | null,
): Record<string, string> {
  const extraFields: Record<string, string> = {};
  for (const setting of settings) {
    const value = values[setting.key]?.trim();
    if (!value) {
      continue;
    }
    const fieldName = mapFn(provider, setting.key);
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
  provider: string,
  oauth2Scopes: ReadonlyArray<string>,
  extras?: Record<string, string>,
) {
  connectProviderProtocol(organizationId, provider, "OAUTH2", {
    oauth2Scopes,
    extras,
  });
}

// buildConnectorInitiateURL builds the start-connect URL for a protocol.
// OAuth2 uses /connectors/initiate; GitHub App has its own endpoint.
export function buildConnectorInitiateURL(
  organizationId: string,
  provider: string,
  protocol: ConnectorProtocol,
  options?: {
    oauth2Scopes?: ReadonlyArray<string>;
    connectorId?: string;
    extras?: Record<string, string>;
  },
): string {
  const baseURL = import.meta.env.VITE_API_URL || window.location.origin;
  const path
    = protocol === "GITHUB_APP"
      ? "/api/console/v1/connectors/github-app/initiate"
      : "/api/console/v1/connectors/initiate";
  const url = new URL(path, baseURL);
  url.searchParams.append("organization_id", organizationId);
  if (protocol !== "GITHUB_APP") {
    url.searchParams.append("provider", provider);
  }
  if (options?.connectorId) {
    url.searchParams.append("connector_id", options.connectorId);
  }
  if (protocol !== "GITHUB_APP") {
    for (const scope of options?.oauth2Scopes ?? []) {
      url.searchParams.append("scope", scope);
    }
    if (options?.extras) {
      for (const [k, v] of Object.entries(options.extras)) {
        url.searchParams.append(k, v);
      }
    }
  }
  url.searchParams.append(
    "continue",
    `/organizations/${organizationId}/access-reviews/connections`,
  );
  return url.toString();
}

// connectProviderProtocol builds the connector-initiate URL for any configured
// protocol and navigates the browser to it.
export function connectProviderProtocol(
  organizationId: string,
  provider: string,
  protocol: ConnectorProtocol,
  options?: {
    oauth2Scopes?: ReadonlyArray<string>;
    connectorId?: string;
    extras?: Record<string, string>;
  },
) {
  window.location.assign(
    buildConnectorInitiateURL(organizationId, provider, protocol, options),
  );
}
