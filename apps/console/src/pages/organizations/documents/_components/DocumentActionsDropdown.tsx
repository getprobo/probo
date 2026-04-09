// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { ActionDropdown, DropdownItem, IconArchive, IconArrowDown, IconPencil, IconTrashCan, useConfirm, useToast } from "@probo/ui";
import { use, useRef } from "react";
import { useFragment, useMutation } from "react-relay";
import { useNavigate, useParams } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { DocumentActionsDropdown_archiveMutation } from "#/__generated__/core/DocumentActionsDropdown_archiveMutation.graphql";
import type { DocumentActionsDropdown_createDraftMutation } from "#/__generated__/core/DocumentActionsDropdown_createDraftMutation.graphql";
import type { DocumentActionsDropdown_documentFragment$key } from "#/__generated__/core/DocumentActionsDropdown_documentFragment.graphql";
import type { DocumentActionsDropdown_exportVersionMutation } from "#/__generated__/core/DocumentActionsDropdown_exportVersionMutation.graphql";
import type { DocumentActionsDropdown_unarchiveMutation } from "#/__generated__/core/DocumentActionsDropdown_unarchiveMutation.graphql";
import type { DocumentActionsDropdown_versionFragment$key } from "#/__generated__/core/DocumentActionsDropdown_versionFragment.graphql";
import { PdfDownloadDialog, type PdfDownloadDialogRef } from "#/components/documents/PdfDownloadDialog";
import { DocumentsConnectionKey, useDeleteDocumentMutation, useDeleteDraftDocumentVersionMutation } from "#/hooks/graph/DocumentGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

const documentFragment = graphql`
  fragment DocumentActionsDropdown_documentFragment on Document {
    id
    title
    status
    canUpdate: permission(action: "core:document:update")
    canArchive: permission(action: "core:document:archive")
    canUnarchive: permission(action: "core:document:unarchive")
    canDelete: permission(action: "core:document:delete")
    versions(first: 1 orderBy: { field: CREATED_AT, direction: DESC }) {
      totalCount
      edges {
        node {
          id
          status
        }
      }
    }
  }
`;

const createDraftDocumentVersionMutation = graphql`
  mutation DocumentActionsDropdown_createDraftMutation(
    $input: CreateDraftDocumentVersionInput!
    $connections: [ID!]!
  ) {
    createDraftDocumentVersion(input: $input) {
      documentVersionEdge @prependEdge(connections: $connections) {
        node {
          id
          content
          status
          publishedAt
          major
          minor
          updatedAt
          signatures(first: 100) {
            edges {
              node {
                id
                state
              }
            }
          }
        }
      }
    }
  }
`;

const archiveDocumentMutation = graphql`
  mutation DocumentActionsDropdown_archiveMutation(
    $input: ArchiveDocumentInput!
  ) {
    archiveDocument(input: $input) {
      document {
        id
        status
        archivedAt
        canUpdate: permission(action: "core:document:update")
        canArchive: permission(action: "core:document:archive")
        canUnarchive: permission(action: "core:document:unarchive")
        canDelete: permission(action: "core:document:delete")
      }
    }
  }
`;

const unarchiveDocumentMutation = graphql`
  mutation DocumentActionsDropdown_unarchiveMutation(
    $input: UnarchiveDocumentInput!
  ) {
    unarchiveDocument(input: $input) {
      document {
        id
        status
        archivedAt
        canUpdate: permission(action: "core:document:update")
        canArchive: permission(action: "core:document:archive")
        canUnarchive: permission(action: "core:document:unarchive")
        canDelete: permission(action: "core:document:delete")
      }
    }
  }
`;

const versionFragment = graphql`
  fragment DocumentActionsDropdown_versionFragment on DocumentVersion {
    id
    major
    minor
    status
    canDeleteDraft: permission(action: "core:document-version:delete-draft")
  }
`;

const exportDocumentVersionMutation = graphql`
  mutation DocumentActionsDropdown_exportVersionMutation(
    $input: ExportDocumentVersionPDFInput!
  ) {
    exportDocumentVersionPDF(input: $input) {
      data
    }
  }
`;

export function DocumentActionsDropdown(props: {
  documentFragmentRef: DocumentActionsDropdown_documentFragment$key;
  versionFragmentRef: DocumentActionsDropdown_versionFragment$key;
  onRefetch: () => void;
}) {
  const { documentFragmentRef, versionFragmentRef, onRefetch } = props;

  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { versionId } = useParams();
  const { __ } = useTranslate();
  const { email: defaultEmail } = use(CurrentUser);
  const pdfDownloadDialogRef = useRef<PdfDownloadDialogRef>(null);
  const confirm = useConfirm();
  const { toast } = useToast();

  const document = useFragment<DocumentActionsDropdown_documentFragment$key>(documentFragment, documentFragmentRef);
  const version = useFragment<DocumentActionsDropdown_versionFragment$key>(versionFragment, versionFragmentRef);

  const lastVersion = document.versions.edges[0].node;
  const isLastVersionPublished = lastVersion.status === "PUBLISHED";
  const isDraft = version.status === "DRAFT";

  const [createDraftDocumentVersion, isCreatingDraft]
    = useMutation<DocumentActionsDropdown_createDraftMutation>(createDraftDocumentVersionMutation);
  const [deleteDocument, isDeleting] = useDeleteDocumentMutation();
  const [archiveDocument, isArchiving]
    = useMutation<DocumentActionsDropdown_archiveMutation>(archiveDocumentMutation);
  const [unarchiveDocument, isUnarchiving]
    = useMutation<DocumentActionsDropdown_unarchiveMutation>(unarchiveDocumentMutation);
  const [deleteDraftDocumentVersion, isDeletingDraft]
    = useDeleteDraftDocumentVersionMutation();
  const [exportDocumentVersion, isExporting]
    = useMutation<DocumentActionsDropdown_exportVersionMutation>(exportDocumentVersionMutation);

  const handleCreateDraft = () => {
    const connectionId = ConnectionHandler.getConnectionID(document.id, "DocumentversionsDropdownMenu_versions");
    createDraftDocumentVersion({
      variables: {
        input: {
          documentID: document.id,
        },
        connections: [connectionId],
      },
      onCompleted: (response, errors) => {
        if (errors) {
          toast({
            variant: "error",
            title: __("Error creating draft"),
            description:
                  errors[0]?.message || __("An unknown error occurred"),
          });
          return;
        }

        const newVersionId
          = response.createDraftDocumentVersion.documentVersionEdge.node.id;

        void navigate(`/organizations/${organizationId}/documents/${document.id}/versions/${newVersionId}`);
      },
      onError(error) {
        toast({ title: __("Error"), description: error.message, variant: "error" });
      },
    });
  };

  const handleArchive = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          archiveDocument({
            variables: { input: { documentId: document.id } },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({ title: __("Error"), description: formatError(__("Failed to archive document"), errors), variant: "error" });
              } else {
                toast({ title: __("Success"), description: __("Document archived successfully."), variant: "success" });
              }
              resolve();
            },
            onError(error) {
              toast({ title: __("Error"), description: error.message, variant: "error" });
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __("This will archive the document \"%s\". It will no longer be editable."),
          document.title,
        ),
        variant: "danger",
        label: __("Archive"),
      },
    );
  };

  const handleUnarchive = () => {
    unarchiveDocument({
      variables: { input: { documentId: document.id } },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: __("Error"), description: formatError(__("Failed to unarchive document"), errors), variant: "error" });
        } else {
          toast({ title: __("Success"), description: __("Document unarchived successfully."), variant: "success" });
        }
      },
      onError(error) {
        toast({ title: __("Error"), description: error.message, variant: "error" });
      },
    });
  };

  const handleDelete = () => {
    const connectionId = ConnectionHandler.getConnectionID(
      organizationId,
      DocumentsConnectionKey,
      { orderBy: { direction: "ASC", field: "TITLE" } },
    );
    confirm(
      () =>
        deleteDocument({
          variables: {
            input: { documentId: document.id },
            connections: [connectionId],
          },
          onSuccess() {
            void navigate(`/organizations/${organizationId}/documents`);
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the document \"%s\". This action cannot be undone.",
          ),
          document.title,
        ),
      },
    );
  };

  const handleDeleteDraft = () => {
    const versionsConnectionId = ConnectionHandler.getConnectionID(document.id, "DocumentversionsDropdownMenu_versions");
    const lastVersionConnectionId = ConnectionHandler.getConnectionID(
      document.id,
      "DocumentversionsDropdownMenu_lastVersion",
      { orderBy: { field: "CREATED_AT", direction: "DESC" } },
    );
    confirm(
      () =>
        deleteDraftDocumentVersion({
          variables: {
            input: { documentVersionId: version.id },
            connections: [versionsConnectionId, lastVersionConnectionId],
          },
          onSuccess() {
            if (versionId) {
              void navigate(`/organizations/${organizationId}/documents/${document.id}`);
            } else {
              onRefetch();
            }
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the draft version %s of \"%s\". This action cannot be undone.",
          ),
          `${version.major}.${version.minor}`,
          document.title,
        ),
      },
    );
  };

  const handleExportDocumentVersion = (options: {
    withWatermark: boolean;
    withSignatures: boolean;
    watermarkEmail?: string;
  }) => {
    const input = {
      documentVersionId: version.id,
      withWatermark: options.withWatermark,
      withSignatures: options.withSignatures,
      ...(options.withWatermark
        && options.watermarkEmail && { watermarkEmail: options.watermarkEmail }),
    };

    exportDocumentVersion({
      variables: { input },
      onCompleted: (data, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: errors[0]?.message || __("Failed to generate PDF"),
            variant: "error",
          });
          return;
        }

        if (data.exportDocumentVersionPDF) {
          const link = window.document.createElement("a");
          link.href = data.exportDocumentVersionPDF.data;
          link.download = `${document.title}-v${version.major}.${version.minor}.pdf`;
          window.document.body.appendChild(link);
          link.click();
          window.document.body.removeChild(link);
        }
      },
      onError(error) {
        toast({ title: __("Error"), description: error.message, variant: "error" });
      },
    });
  };

  return (
    <>
      <PdfDownloadDialog
        ref={pdfDownloadDialogRef}
        onDownload={handleExportDocumentVersion}
        isLoading={isExporting}
        defaultEmail={defaultEmail}
      />
      <ActionDropdown variant="secondary">
        {document.canUpdate && isLastVersionPublished && (
          <DropdownItem
            onClick={handleCreateDraft}
            icon={IconPencil}
            disabled={isCreatingDraft}
          >
            {__("Create new draft")}
          </DropdownItem>
        )}
        {isDraft
          && document.versions.totalCount > 1
          && version.canDeleteDraft && (
          <DropdownItem
            onClick={handleDeleteDraft}
            icon={IconTrashCan}
            disabled={isDeletingDraft}
          >
            {__("Delete draft document")}
          </DropdownItem>
        )}
        <DropdownItem
          onClick={() => pdfDownloadDialogRef.current?.open()}
          icon={IconArrowDown}
          disabled={isExporting}
        >
          {__("Download PDF")}
        </DropdownItem>
        {document.canArchive && document.status === "ACTIVE" && (
          <DropdownItem
            icon={IconArchive}
            disabled={isArchiving}
            onClick={handleArchive}
          >
            {__("Archive document")}
          </DropdownItem>
        )}
        {document.canUnarchive && document.status === "ARCHIVED" && (
          <DropdownItem
            icon={IconArchive}
            disabled={isUnarchiving}
            onClick={handleUnarchive}
          >
            {__("Unarchive document")}
          </DropdownItem>
        )}
        {document.canDelete && (
          <DropdownItem
            variant="danger"
            icon={IconTrashCan}
            disabled={isDeleting}
            onClick={handleDelete}
          >
            {__("Delete document")}
          </DropdownItem>
        )}
      </ActionDropdown>
    </>
  );
}
