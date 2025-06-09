import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "../useMutationWithToasts";
import type { DocumentGraphDeleteMutation } from "./__generated__/DocumentGraphDeleteMutation.graphql";

export const documentsQuery = graphql`
  query DocumentGraphListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...DocumentsPageListFragment
    }
  }
`;

export const DocumentsConnectionKey = "DocumentsPageFragment_documents";

const deleteDocumentMutation = graphql`
  mutation DocumentGraphDeleteMutation(
    $input: DeleteDocumentInput!
    $connections: [ID!]!
  ) {
    deleteDocument(input: $input) {
      deletedDocumentId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteDocumentMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<DocumentGraphDeleteMutation>(
    deleteDocumentMutation,
    {
      successMessage: __("Document deleted successfully."),
      errorMessage: __("Failed to delete document. Please try again."),
    }
  );
}

export const documentNodeQuery = graphql`
  query DocumentGraphNodeQuery($documentId: ID!) {
    node(id: $documentId) {
      ... on Document {
        ...DocumentDetailPageDocumentFragment
      }
    }
  }
`;
