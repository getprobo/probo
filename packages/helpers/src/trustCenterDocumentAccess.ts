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

export type TrustCenterDocumentAccessInfo = {
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
