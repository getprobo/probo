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
import { graphql, readInlineData } from "relay-runtime";

import type { documentAccessInfo_audit$key } from "#/__generated__/core/documentAccessInfo_audit.graphql";
import type { documentAccessInfo_document$key } from "#/__generated__/core/documentAccessInfo_document.graphql";
import type { documentAccessInfo_file$key } from "#/__generated__/core/documentAccessInfo_file.graphql";

export const documentAccessInfoDocumentFragment = graphql`
  fragment documentAccessInfo_document on Document
  @inline
  @argumentDefinitions(
    compliancePortalId: { type: "ID!" }
    accessId: { type: "ID!" }
  ) {
    id
    versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
      edges {
        node {
          title
          documentType
        }
      }
    }
    compliancePortalDocument(compliancePortalId: $compliancePortalId) {
      visibility
    }
    compliancePortalDocumentAccess(compliancePortalAccessId: $accessId) {
      id
      status
    }
  }
`;

export const documentAccessInfoAuditFragment = graphql`
  fragment documentAccessInfo_audit on Audit
  @inline
  @argumentDefinitions(
    compliancePortalId: { type: "ID!" }
    accessId: { type: "ID!" }
  ) {
    reportFile {
      id
      fileName
    }
    framework {
      name
    }
    compliancePortalAudit(compliancePortalId: $compliancePortalId) {
      visibility
    }
    compliancePortalDocumentAccess(compliancePortalAccessId: $accessId) {
      id
      status
    }
  }
`;

export const documentAccessInfoFileFragment = graphql`
  fragment documentAccessInfo_file on CompliancePortalFile
  @inline
  @argumentDefinitions(accessId: { type: "ID!" }) {
    id
    name
    category
    compliancePortalVisibility
    compliancePortalDocumentAccess(compliancePortalAccessId: $accessId) {
      id
      status
    }
  }
`;

export function documentAccessInfoFromDocument(
  fragmentRef: documentAccessInfo_document$key,
): CompliancePortalDocumentAccessInfo | null {
  const node = readInlineData(documentAccessInfoDocumentFragment, fragmentRef);
  if (node.compliancePortalDocument?.visibility !== "RESTRICTED") {
    return null;
  }

  return {
    type: "document",
    name: node.versions?.edges[0]?.node.title ?? "",
    category: node.versions?.edges[0]?.node.documentType ?? "",
    id: node.id,
    status: node.compliancePortalDocumentAccess?.status ?? null,
  };
}

export function documentAccessInfoFromAudit(
  fragmentRef: documentAccessInfo_audit$key,
): CompliancePortalDocumentAccessInfo | null {
  const node = readInlineData(documentAccessInfoAuditFragment, fragmentRef);
  if (node.compliancePortalAudit?.visibility !== "RESTRICTED" || node.reportFile == null) {
    return null;
  }

  return {
    type: "report",
    name: node.reportFile.fileName,
    category: node.framework?.name ?? "",
    id: node.reportFile.id,
    status: node.compliancePortalDocumentAccess?.status ?? null,
  };
}

export function documentAccessInfoFromFile(
  fragmentRef: documentAccessInfo_file$key,
): CompliancePortalDocumentAccessInfo | null {
  const node = readInlineData(documentAccessInfoFileFragment, fragmentRef);
  if (
    node.compliancePortalVisibility !== "RESTRICTED"
    && node.compliancePortalVisibility !== "NONE"
  ) {
    return null;
  }

  return {
    type: "file",
    name: node.name,
    category: node.category,
    id: node.id,
    status: node.compliancePortalDocumentAccess?.status ?? null,
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
