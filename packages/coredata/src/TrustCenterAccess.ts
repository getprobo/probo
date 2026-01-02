import type { TrustCenterDocumentAccess } from "./TrustCenterDocumentAccess";

export interface TrustCenterAccess {
  id: string;
  email: string;
  name: string;
  active: boolean;
  hasAcceptedNonDisclosureAgreement: boolean;
  createdAt: string;
  lastTokenExpiresAt?: string | null;
  pendingRequestCount: number;
  activeCount: number;
  documentAccesses?: TrustCenterDocumentAccess[];
}
