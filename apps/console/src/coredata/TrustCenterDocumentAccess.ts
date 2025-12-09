export type TrustCenterDocumentAccess = {
  active: boolean;
  status: string;
  requested: boolean;
  document?: {
    id: string;
    title: string;
    documentType: string;
  } | null;
  report?: {
    id: string;
    filename: string;
    audit?: {
      id: string;
      framework: {
        name: string;
      };
    } | null;
  } | null;
  trustCenterFile?: {
    id: string;
    name: string;
    category: string;
  } | null;
};
