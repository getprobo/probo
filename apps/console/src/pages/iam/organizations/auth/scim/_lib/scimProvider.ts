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

import type { TonedCardTone } from "#/components/TonedCard/variants";

export type ScimConnectorProvider = "GOOGLE_WORKSPACE" | "MICROSOFT_365";

export type SCIMProviderCopy
  = "googleWorkspaceConnector"
    | "microsoft365Connector";

export type BridgeStatus = "connected" | "syncing" | "error" | "disabled";

export const bridgeStatusTone: Record<BridgeStatus, TonedCardTone> = {
  connected: "green",
  syncing: "amber",
  error: "red",
  disabled: "sand",
};

export const bridgeStatusBadgeColor: Record<
  BridgeStatus,
  "green" | "amber" | "red" | "neutral"
> = {
  connected: "green",
  syncing: "amber",
  error: "red",
  disabled: "neutral",
};

export function scimProviderCopy(type: string): SCIMProviderCopy {
  return type === "MICROSOFT_365"
    ? "microsoft365Connector"
    : "googleWorkspaceConnector";
}

export function bridgeStatus(state: string): BridgeStatus {
  if (state === "DISABLED") {
    return "disabled";
  }
  if (state === "FAILED") {
    return "error";
  }
  if (state === "PENDING" || state === "SYNCING") {
    return "syncing";
  }
  return "connected";
}

export function isExcludedUserEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}

export function initiateScimConnectorUrl({
  organizationId,
  provider,
  scopes,
}: {
  organizationId: string;
  provider: ScimConnectorProvider;
  scopes: readonly string[];
}) {
  const url = new URL(
    "/api/console/v1/connectors/initiate",
    import.meta.env.VITE_API_URL || window.location.origin,
  );
  url.searchParams.set("organization_id", organizationId);
  url.searchParams.set("provider", provider);
  for (const scope of scopes) {
    url.searchParams.append("scope", scope);
  }

  const continueUrl = new URL(window.location.origin);
  continueUrl.pathname = `/organizations/${encodeURIComponent(organizationId)}/settings/auth/scim`;
  url.searchParams.set("continue", continueUrl.pathname);

  return url.toString();
}
