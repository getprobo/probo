import type { TrustCenterDocumentAccess, TrustCenterDocumentAccessStatus } from "@probo/coredata";

export function getTrustCenterDocumentAccessStatusBadgeVariant(status: TrustCenterDocumentAccessStatus) {
  switch (status) {
    case "REQUESTED":
      return "warning" as const;
    case "GRANTED":
      return "success" as const;
    case "REJECTED":
    case "REVOKED":
      return "danger" as const;
  }
}

export function getTrustCenterDocumentAccessStatusLabel(status: TrustCenterDocumentAccessStatus, __: (key: string) => string) {
  switch (status) {
    case "REQUESTED":
      return __("requested");
    case "GRANTED":
      return __("granted");
    case "REJECTED":
      return __("rejected");
    case "REVOKED":
      return __("revoked");
  }
}

export type TrustCenterDocumentAccessInfo = {
  persisted: boolean;
  variant: "info",
  name: string,
  type: "document",
  typeLabel: string,
  category: string;
  id: string;
  requested: boolean;
  active: boolean;
  status: TrustCenterDocumentAccessStatus;
} | {
  persisted: boolean;
  variant: "success",
  name: string,
  type: "report",
  typeLabel: string,
  category: string;
  id: string;
  requested: boolean;
  active: boolean;
  status: TrustCenterDocumentAccessStatus;
} | {
  persisted: boolean;
  variant: "highlight",
  name: string,
  type: "file",
  typeLabel: string,
  category: string;
  id: string;
  requested: boolean;
  active: boolean;
  status: TrustCenterDocumentAccessStatus;
}

export function getTrustCenterDocumentAccessInfo(
  docAccess: TrustCenterDocumentAccess,
  __: (key: string) => string
): TrustCenterDocumentAccessInfo {
  if (docAccess.document) {
    return {
      persisted: docAccess.id !== docAccess.document.id,
      variant: "info" as const,
      name: docAccess.document.title,
      type: "document",
      typeLabel: __("Document"),
      category: docAccess.document.documentType,
      id: docAccess.document.id,
      requested: docAccess.requested,
      active: docAccess.active,
      status: docAccess.status,
    };
  }
  if (docAccess.report) {
    return {
      persisted: docAccess.id !== docAccess.report.id,
      variant: "success" as const,
      name: docAccess.report.filename,
      type: "report",
      typeLabel: __("Report"),
      category: docAccess.report.audit?.framework.name ?? "",
      id: docAccess.report.id,
      requested: docAccess.requested,
      active: docAccess.active,
      status: docAccess.status,
    };
  }
  if (docAccess.trustCenterFile) {
    return {
      persisted: docAccess.id !== docAccess.trustCenterFile.id,
      variant: "highlight" as const,
      name: docAccess.trustCenterFile.name,
      type: "file",
      typeLabel: __("File"),
      category: docAccess.trustCenterFile.category,
      id: docAccess.trustCenterFile.id,
      requested: docAccess.requested,
      active: docAccess.active,
      status: docAccess.status,
    };
  }

  throw new Error("Unknown trust center access document type");
}
