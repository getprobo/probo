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

import type { ConnectorConnectionStatus } from "#/__generated__/core/ConnectorListItemStatus_connector.graphql";
import type { ProviderOrganizationsStatus } from "#/__generated__/core/ConnectorListItem_connector.graphql";

export type { ConnectorConnectionStatus };

export type ConnectionIssueKey = "reconnect" | "notAuthorized" | "credentials";

export type ConnectionTone = "green" | "amber" | "red";

export interface ConnectionSignal {
  status: ConnectorConnectionStatus;
  organizationsStatus: ProviderOrganizationsStatus;
  canReconnect: boolean;
}

export interface ConnectionPresentation {
  status: ConnectorConnectionStatus;
  issue: ConnectionIssueKey | null;
}

// A linked record can arrive before the probe. Absent is not a failure.
export function connectionSignalFrom(connector: {
  connectionStatus?: ConnectorConnectionStatus | null;
  canReconnect?: boolean | null;
  providerOrganizations?: { status?: ProviderOrganizationsStatus | null } | null;
}): ConnectionSignal | null {
  if (connector.connectionStatus == null) {
    return null;
  }

  return {
    status: connector.connectionStatus,
    organizationsStatus: connector.providerOrganizations?.status ?? "NOT_APPLICABLE",
    canReconnect: connector.canReconnect ?? false,
  };
}

export function presentConnection(signal: ConnectionSignal): ConnectionPresentation {
  if (signal.status === "NOT_AUTHORIZED") {
    return { status: signal.status, issue: "notAuthorized" };
  }

  if (
    signal.status === "CONNECTED"
    && signal.organizationsStatus !== "UNAVAILABLE"
  ) {
    return { status: "CONNECTED", issue: null };
  }

  // An unreadable organization list fails the connector, even when the probe says connected.
  if (signal.status === "CONNECTED") {
    return signal.canReconnect
      ? { status: "RECONNECT_REQUIRED", issue: "reconnect" }
      : { status: "DISCONNECTED", issue: "credentials" };
  }

  return {
    status: signal.status,
    issue: signal.status === "RECONNECT_REQUIRED" || signal.canReconnect
      ? "reconnect"
      : "credentials",
  };
}

export function connectionIssueKeys(
  presented: readonly ConnectionPresentation[],
): ConnectionIssueKey[] {
  const present = new Set<ConnectionIssueKey>();
  for (const item of presented) {
    if (item.issue != null) {
      present.add(item.issue);
    }
  }

  return (["reconnect", "notAuthorized", "credentials"] as const).filter(key => present.has(key));
}

// A reconnect beside a disconnect stays amber. Only a uniform status uses that status's tone.
export function aggregateConnectionTone(
  presented: readonly ConnectionPresentation[],
): ConnectionTone {
  const [first] = presented;
  if (first != null && presented.every(item => item.status === first.status)) {
    return connectionTone(first.status);
  }

  return "amber";
}

export function connectionTone(status: ConnectorConnectionStatus): ConnectionTone {
  if (status === "CONNECTED") {
    return "green";
  }
  if (status === "RECONNECT_REQUIRED" || status === "NOT_AUTHORIZED") {
    return "amber";
  }
  return "red";
}

export function groupByProvider<T extends { provider: string }>(
  connectors: readonly T[],
): T[][] {
  const groups: T[][] = [];
  const indexByProvider = new Map<string, number>();

  for (const connector of connectors) {
    const index = indexByProvider.get(connector.provider);
    if (index == null) {
      indexByProvider.set(connector.provider, groups.length);
      groups.push([connector]);
      continue;
    }

    groups[index].push(connector);
  }

  return groups;
}
