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

import type {
  documentAccessInfo_documentAccess$data,
  documentAccessInfo_documentAccess$key,
} from "#/__generated__/core/documentAccessInfo_documentAccess.graphql";

export const documentAccessInfoFragment = graphql`
  fragment documentAccessInfo_documentAccess on CompliancePortalDocumentAccess @inline {
    id
    status
    document {
      id
      versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
        edges {
          node {
            title
            documentType
          }
        }
      }
    }
    reportFile {
      id
      fileName
    }
    audit {
      framework {
        name
      }
    }
    compliancePortalFile {
      id
      name
      category
    }
  }
`;

export function documentAccessInfoFrom(
  fragmentRef: documentAccessInfo_documentAccess$key,
  t: (key: string) => string,
): CompliancePortalDocumentAccessInfo {
  const node = readInlineData(documentAccessInfoFragment, fragmentRef);
  return toDocumentAccessInfo(node, t);
}

function toDocumentAccessInfo(
  node: documentAccessInfo_documentAccess$data,
  t: (key: string) => string,
): CompliancePortalDocumentAccessInfo {
  if (node.document) {
    return {
      persisted: node.id !== node.document.id,
      variant: "info",
      name: node.document.versions?.edges[0]?.node.title ?? "",
      type: "document",
      typeLabel: t("documentAccessList.types.document"),
      category: node.document.versions?.edges[0]?.node.documentType ?? "",
      id: node.document.id,
      status: node.status,
    };
  }
  if (node.reportFile) {
    return {
      persisted: node.id !== node.reportFile.id,
      variant: "success",
      name: node.reportFile.fileName,
      type: "report",
      typeLabel: t("documentAccessList.types.report"),
      category: node.audit?.framework?.name ?? "",
      id: node.reportFile.id,
      status: node.status,
    };
  }
  if (node.compliancePortalFile) {
    return {
      persisted: node.id !== node.compliancePortalFile.id,
      variant: "highlight",
      name: node.compliancePortalFile.name,
      type: "file",
      typeLabel: t("documentAccessList.types.file"),
      category: node.compliancePortalFile.category,
      id: node.compliancePortalFile.id,
      status: node.status,
    };
  }
  throw new Error("Unknown compliance page access document type");
}

export function documentAccessTypeColor(
  variant: CompliancePortalDocumentAccessInfo["variant"],
): "sky" | "green" | "gold" {
  switch (variant) {
    case "info":
      return "sky";
    case "success":
      return "green";
    case "highlight":
      return "gold";
  }
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
  status: CompliancePortalDocumentAccessStatus,
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
