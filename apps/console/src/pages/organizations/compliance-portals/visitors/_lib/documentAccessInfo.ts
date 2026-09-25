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

import type { CompliancePortalDocumentAccessStatus } from "@probo/coredata";
import type { CompliancePortalDocumentAccessInfo } from "@probo/helpers";

const resourceKindToType = {
  DOCUMENT: "document",
  REPORT: "report",
  FILE: "file",
} as const;

export function documentAccessInfoFromResource(resource: {
  kind: keyof typeof resourceKindToType;
  resourceId: string;
  name: string;
  category: string;
  status: CompliancePortalDocumentAccessStatus | null | undefined;
}): CompliancePortalDocumentAccessInfo {
  return {
    type: resourceKindToType[resource.kind],
    id: resource.resourceId,
    name: resource.name,
    category: resource.category,
    status: resource.status ?? null,
  };
}

export function documentAccessKey(item: CompliancePortalDocumentAccessInfo): string {
  return `${item.type}:${item.id}`;
}

export function documentAccessStatusColor(
  status: CompliancePortalDocumentAccessStatus,
): "amber" | "green" | "red" {
  switch (status) {
    case "REQUESTED":
      return "amber";
    case "GRANTED":
      return "green";
    case "REJECTED":
    case "REVOKED":
      return "red";
  }
}

export function rejectOrRevokeStatus(
  status: CompliancePortalDocumentAccessStatus | null,
): "REJECTED" | "REVOKED" {
  return status === "GRANTED" ? "REVOKED" : "REJECTED";
}

export function updateAccessInput(
  accessId: string,
  updates: CompliancePortalDocumentAccessInfo[],
) {
  const documents: { id: string; status: CompliancePortalDocumentAccessStatus }[] = [];
  const reports: { id: string; status: CompliancePortalDocumentAccessStatus }[] = [];
  const compliancePortalFiles: { id: string; status: CompliancePortalDocumentAccessStatus }[] = [];

  for (const update of updates) {
    if (update.status == null) {
      continue;
    }

    const item = { id: update.id, status: update.status };
    switch (update.type) {
      case "document":
        documents.push(item);
        break;
      case "report":
        reports.push(item);
        break;
      case "file":
        compliancePortalFiles.push(item);
        break;
    }
  }

  return {
    id: accessId,
    documents: documents.length > 0 ? documents : undefined,
    reports: reports.length > 0 ? reports : undefined,
    compliancePortalFiles: compliancePortalFiles.length > 0 ? compliancePortalFiles : undefined,
  };
}
