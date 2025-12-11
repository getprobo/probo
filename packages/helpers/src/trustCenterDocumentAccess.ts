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

interface ITrustCenterDocumentAccessInfo {
  variant: string;
  type: string;
  persisted: boolean;
  name: string,
  typeLabel: string,
  category: string;
  id: string;
  status: TrustCenterDocumentAccessStatus;
}

export type TrustCenterDocumentAccessInfo = ITrustCenterDocumentAccessInfo & (
  {
    variant: "info",
    type: "document",
  } | {
    variant: "success",
    type: "report",
  } | {
    variant: "highlight",
    type: "file",
  }
)

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
      status: docAccess.status,
    };
  }

  throw new Error("Unknown trust center access document type");
}
