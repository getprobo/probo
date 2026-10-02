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
    case "NEW_RELIC":
      if (settingKey === "region") return "newRelicRegion";
      break;
    case "RETOOL":
      if (settingKey === "baseUrl") return "retoolBaseUrl";
      break;
    case "TWINGATE":
      if (settingKey === "network") return "twingateNetwork";
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

// Each connect path passes its own settings list and map. A provider that
// offers both paths (1Password) declares different settings on each.
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
