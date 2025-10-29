import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useTranslate } from "@probo/i18n";
import type { TrustCenterDocumentGraphUpdateMutation } from "./__generated__/TrustCenterDocumentGraphUpdateMutation.graphql";

export const trustCenterDocumentsQuery = graphql`
  query TrustCenterDocumentGraphQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        documents(first: 100) {
          edges {
            node {
              id
              ...TrustCenterDocumentsCardFragment
            }
          }
        }
      }
    }
  }
`;

export const updateDocumentVisibilityMutation = graphql`
  mutation TrustCenterDocumentGraphUpdateMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      document {
        id
        trustCenterVisibility
        ...TrustCenterDocumentsCardFragment
      }
    }
  }
`;

export function useUpdateDocumentVisibilityMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<TrustCenterDocumentGraphUpdateMutation>(
    updateDocumentVisibilityMutation,
    {
      successMessage: __("Document visibility updated successfully."),
      errorMessage: __("Failed to update document visibility"),
    }
  );
}
