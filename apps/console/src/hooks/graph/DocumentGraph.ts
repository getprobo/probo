import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "../useMutationWithToasts";
import type { DocumentGraphDeleteMutation } from "./__generated__/DocumentGraphDeleteMutation.graphql";
import type { DocumentGraphSendSigningNotificationsMutation } from "./__generated__/DocumentGraphSendSigningNotificationsMutation.graphql";
import type { DocumentGraphDeleteDraftMutation } from "./__generated__/DocumentGraphDeleteDraftMutation.graphql";

export const documentsQuery = graphql`
  query DocumentGraphListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...DocumentsPageListFragment
    }
  }
`;

export const DocumentsConnectionKey = "DocumentsListQuery_documents";

const deleteDocumentMutation = graphql`
  mutation DocumentGraphDeleteMutation(
    $input: DeleteDocumentInput!
  ) {
    deleteDocument(input: $input) {
      deletedDocumentId @deleteRecord
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

const deleteDraftDocumentVersionMutation = graphql`
  mutation DocumentGraphDeleteDraftMutation(
    $input: DeleteDraftDocumentVersionInput!
    $connections: [ID!]!
  ) {
    deleteDraftDocumentVersion(input: $input) {
      deletedDocumentVersionId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteDraftDocumentVersionMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<DocumentGraphDeleteDraftMutation>(
    deleteDraftDocumentVersionMutation,
    {
      successMessage: __("Draft deleted successfully."),
      errorMessage: __("Failed to delete draft. Please try again."),
    }
  );
}

const sendSigningNotificationsMutation = graphql`
  mutation DocumentGraphSendSigningNotificationsMutation(
    $input: SendSigningNotificationsInput!
  ) {
    sendSigningNotifications(input: $input) {
      success
    }
  }
`;

export function useSendSigningNotificationsMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<DocumentGraphSendSigningNotificationsMutation>(
    sendSigningNotificationsMutation,
    {
      successMessage: __("Signing notifications sent successfully."),
      errorMessage: __("Failed to send signing notifications. Please try again."),
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
