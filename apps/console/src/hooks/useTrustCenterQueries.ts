import { useQuery, useMutation } from "@tanstack/react-query";
import { GraphQLError } from "graphql";
import { buildEndpoint } from "/providers/RelayProviders";

export interface TrustCenterDocument {
  id: string;
  title: string;
  documentType: string;
}

export interface TrustCenterAudit {
  id: string;
  framework: {
    name: string;
  };
  report: {
    id: string;
    filename: string;
    downloadUrl: string | null;
  } | null;
}

export interface TrustCenterVendor {
  id: string;
  name: string;
  category: string;
  privacyPolicyUrl?: string | null;
  websiteUrl?: string | null;
}

interface TrustCenterQueryData {
  trustCenterBySlug: {
    id: string;
    active: boolean;
    slug: string;
    isUserAuthenticated: boolean;
    organization: {
      id: string;
      name: string;
      logoUrl: string | null;
    };
    documents: {
      edges: Array<{
        node: TrustCenterDocument;
      }>;
    };
    audits: {
      edges: Array<{
        node: TrustCenterAudit;
      }>;
    };
    vendors: {
      edges: Array<{
        node: TrustCenterVendor;
      }>;
    };
  } | null;
}

interface GraphQLResponse<T = unknown> {
  data?: T;
  errors?: GraphQLError[];
}

interface ExportDocumentPDFData {
  exportDocumentPDF: {
    data: string;
  };
}

interface CreateTrustCenterAccessData {
  createTrustCenterAccess: {
    trustCenterAccess: {
      id: string;
      email: string;
      name: string;
    };
  };
}

interface TrustCenterQueryVariables {
  slug: string;
}

interface ExportDocumentPDFVariables {
  input: {
    documentId: string;
  };
}

interface CreateTrustCenterAccessVariables {
  input: {
    trustCenterId: string;
    email: string;
    name: string;
  };
}

type GraphQLVariables = TrustCenterQueryVariables | ExportDocumentPDFVariables | CreateTrustCenterAccessVariables | Record<string, never>;

async function trustCenterGraphQLRequest<T = unknown>(
  operationName: string,
  query: string,
  variables: GraphQLVariables = {}
): Promise<GraphQLResponse<T>> {
  const response = await fetch(buildEndpoint("/api/trust/v1/graphql"), {
    method: "POST",
    credentials: "include",
    headers: {
      Accept: "application/graphql-response+json; charset=utf-8, application/json; charset=utf-8",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      operationName,
      query,
      variables,
    }),
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const result = await response.json();
  return result;
}

function isCriticalError(error: GraphQLError): boolean {
  const message = error.message?.toLowerCase() || '';
  const path = error.path || [];

  if (message.includes('access denied') || message.includes('authentication required')) {
    if (path.length > 2) {
      return false;
    }
  }

  return true;
}

const TRUST_CENTER_QUERY = `
  query PublicTrustCenterPageQuery($slug: String!) {
    trustCenterBySlug(slug: $slug) {
      id
      active
      slug
      isUserAuthenticated
      organization {
        id
        name
        logoUrl
      }
      documents(first: 100) {
        edges {
          node {
            id
            title
            documentType
          }
        }
      }
      audits(first: 100) {
        edges {
          node {
            id
            framework {
              name
            }
            report {
              id
              filename
              downloadUrl
            }
          }
        }
      }
      vendors(first: 100) {
        edges {
          node {
            id
            name
            category
            websiteUrl
            privacyPolicyUrl
          }
        }
      }
    }
  }
`;

const EXPORT_DOCUMENT_PDF_MUTATION = `
  mutation PublicTrustCenterDocumentsExportPDFMutation(
    $input: ExportDocumentPDFInput!
  ) {
    exportDocumentPDF(input: $input) {
      data
    }
  }
`;

const CREATE_TRUST_CENTER_ACCESS_MUTATION = `
  mutation PublicTrustCenterAccessRequestDialogMutation(
    $input: CreateTrustCenterAccessInput!
  ) {
    createTrustCenterAccess(input: $input) {
      trustCenterAccess {
        id
        email
        name
      }
    }
  }
`;

export function useTrustCenterQuery(slug: string) {
  return useQuery<TrustCenterQueryData>({
    queryKey: ["trust-center", slug],
    queryFn: async () => {
      const result = await trustCenterGraphQLRequest<TrustCenterQueryData>(
        "PublicTrustCenterPageQuery",
        TRUST_CENTER_QUERY,
        { slug }
      );

      if (result.errors && result.errors.length > 0) {
        const criticalErrors = result.errors.filter(isCriticalError);

        if (criticalErrors.length > 0) {
          throw new Error(
            `GraphQL error: ${criticalErrors.map((e) => e.message).join(", ")}`
          );
        }
      }

      if (!result.data) {
        throw new Error("No data returned from GraphQL query");
      }

      return result.data;
    },
    enabled: !!slug,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (failureCount, error) => {
      if (error.message.includes("UNAUTHENTICATED") || error.message.includes("401")) {
        return false;
      }
      return failureCount < 3;
    },
  });
}

export function useExportDocumentPDF() {
  return useMutation<ExportDocumentPDFData, Error, string>({
    mutationFn: async (documentId: string) => {
      const result = await trustCenterGraphQLRequest<ExportDocumentPDFData>(
        "PublicTrustCenterDocumentsExportPDFMutation",
        EXPORT_DOCUMENT_PDF_MUTATION,
        { input: { documentId } }
      );

      if (result.errors && result.errors.length > 0) {
        throw new Error(
          `GraphQL error: ${result.errors.map((e) => e.message).join(", ")}`
        );
      }

      if (!result.data) {
        throw new Error("No data returned from mutation");
      }

      return result.data;
    },
  });
}

export function useCreateTrustCenterAccess() {
  return useMutation<CreateTrustCenterAccessData, Error, { trustCenterId: string; email: string; name: string }>({
    mutationFn: async (input: { trustCenterId: string; email: string; name: string }) => {
      const result = await trustCenterGraphQLRequest<CreateTrustCenterAccessData>(
        "PublicTrustCenterAccessRequestDialogMutation",
        CREATE_TRUST_CENTER_ACCESS_MUTATION,
        { input }
      );

      if (result.errors && result.errors.length > 0) {
        throw new Error(
          `GraphQL error: ${result.errors.map((e) => e.message).join(", ")}`
        );
      }

      if (!result.data) {
        throw new Error("No data returned from mutation");
      }

      return result.data;
    },
  });
}
