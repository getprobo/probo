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

import {
  isGCPServiceAccountEmail,
  isGCPWorkloadIdentityProvider,
} from "../../dialogs/_lib/connectorSettings";

export type GcpTerraformConnector = {
  projectId: string;
  workloadIdentityProvider: string;
  serviceAccountEmail: string;
};

export type ParseGcpTerraformConnectorsResult
  = | { ok: true; connectors: GcpTerraformConnector[] }
    | { ok: false; error: "empty" | "invalidJson" | "invalidShape" }
    | { ok: false; error: "invalidEntry"; projectId: string };

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isConnectorEntry(value: unknown): value is {
  workload_identity_provider: string;
  service_account_email: string;
} {
  return (
    isRecord(value)
    && typeof value.workload_identity_provider === "string"
    && typeof value.service_account_email === "string"
  );
}

function isConnectorsMap(
  value: unknown,
): value is Record<string, unknown> {
  if (!isRecord(value)) {
    return false;
  }

  const entries = Object.values(value);
  return entries.length > 0 && entries.every(isConnectorEntry);
}

function unwrapConnectorsMap(raw: unknown): unknown {
  if (!isRecord(raw)) {
    return raw;
  }

  if (isRecord(raw.connectors) && isConnectorsMap(raw.connectors.value)) {
    return raw.connectors.value;
  }

  if (isConnectorsMap(raw.value)) {
    return raw.value;
  }

  return raw;
}

export function parseGcpTerraformConnectors(
  raw: string,
): ParseGcpTerraformConnectorsResult {
  const trimmed = raw.trim();
  if (trimmed === "") {
    return { ok: false, error: "empty" };
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(trimmed);
  } catch {
    return { ok: false, error: "invalidJson" };
  }

  const map = unwrapConnectorsMap(parsed);
  if (!isConnectorsMap(map)) {
    return { ok: false, error: "invalidShape" };
  }

  const projectIds = Object.keys(map).sort();
  const connectors: GcpTerraformConnector[] = [];

  for (const projectId of projectIds) {
    const entry = map[projectId];
    if (!isConnectorEntry(entry)) {
      return { ok: false, error: "invalidEntry", projectId };
    }

    const workloadIdentityProvider = entry.workload_identity_provider.trim();
    const serviceAccountEmail = entry.service_account_email.trim();

    if (
      !isGCPWorkloadIdentityProvider(workloadIdentityProvider)
      || !isGCPServiceAccountEmail(serviceAccountEmail)
    ) {
      return { ok: false, error: "invalidEntry", projectId };
    }

    connectors.push({
      projectId,
      workloadIdentityProvider,
      serviceAccountEmail,
    });
  }

  return { ok: true, connectors };
}
