export const trustCenterDocumentAccessStatus = {
  REQUESTED: "REQUESTED",
  GRANTED: "GRANTED",
  REJECTED: "REJECTED",
  REVOKED: "REVOKED",
} as const;

export type TrustCenterDocumentAccessStatus = (typeof trustCenterDocumentAccessStatus)[keyof typeof trustCenterDocumentAccessStatus];

export type TrustCenterDocumentAccess = {
  id: string;
  status: TrustCenterDocumentAccessStatus;
  document?: {
    id: string;
    title: string;
    documentType: string;
  } | null;
  report?: {
    id: string;
    name: string | null | undefined;
    file?: {
      fileName: string;
    } | null;
    frameworkType?: string | null;
    framework?: {
      name: string;
    } | null;
  } | null;
  trustCenterFile?: {
    id: string;
    name: string;
    category: string;
  } | null;
};
