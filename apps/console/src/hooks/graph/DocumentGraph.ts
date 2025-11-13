import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "../useMutationWithToasts";
import type { DocumentGraphDeleteMutation } from "./__generated__/DocumentGraphDeleteMutation.graphql";
import type { DocumentGraphSendSigningNotificationsMutation } from "./__generated__/DocumentGraphSendSigningNotificationsMutation.graphql";
import type { DocumentGraphDeleteDraftMutation } from "./__generated__/DocumentGraphDeleteDraftMutation.graphql";
import type { DocumentGraphBulkExportDocumentsMutation } from "./__generated__/DocumentGraphBulkExportDocumentsMutation.graphql";

export const documentsQuery = graphql`
  query DocumentGraphListQuery($organizationId: ID!, $includeSignatures: Boolean = false, $useRequestedDocuments: Boolean = false) {
    organization: node(id: $organizationId) {
      id
      ...DocumentsPageListFragment @skip(if: $useRequestedDocuments) @arguments(includeSignatures: $includeSignatures)
      ...DocumentsPageRequestedListFragment @include(if: $useRequestedDocuments) @arguments(includeSignatures: $includeSignatures)
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
      errorMessage: __("Failed to delete document"),
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
      errorMessage: __("Failed to delete draft"),
    }
  );
}

const bulkDeleteDocumentsMutation = graphql`
  mutation DocumentGraphBulkDeleteDocumentsMutation(
    $input: BulkDeleteDocumentsInput!
  ) {
    bulkDeleteDocuments(input: $input) {
      deletedDocumentIds @deleteRecord
    }
  }
`;

export function useBulkDeleteDocumentsMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts(
    bulkDeleteDocumentsMutation,
    {
      successMessage: __("Documents deleted successfully."),
      errorMessage: __("Failed to delete documents"),
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
      errorMessage: __("Failed to send signing notifications"),
    }
  );
}

const bulkExportDocumentsMutation = graphql`
  mutation DocumentGraphBulkExportDocumentsMutation(
    $input: BulkExportDocumentsInput!
  ) {
    bulkExportDocuments(input: $input) {
      exportJobId
    }
  }
`;

export function useBulkExportDocumentsMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<DocumentGraphBulkExportDocumentsMutation>(
    bulkExportDocumentsMutation,
    {
      successMessage: __("Document export started successfully. You will receive an email when the export is ready."),
      errorMessage: __("Failed to start document export"),
    }
  );
}

export const documentNodeQuery = graphql`
  query DocumentGraphNodeQuery($documentId: ID!, $includeControls: Boolean!, $includeSignatures: Boolean!, $useRequestedVersions: Boolean = false) {
    node(id: $documentId) {
      ... on Document {
        ...DocumentDetailPageDocumentFragment @arguments(includeControls: $includeControls, includeSignatures: $includeSignatures, useRequestedVersions: $useRequestedVersions)
      }
    }
  }
`;
