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
import { ActionDropdown, DropdownItem, IconArchive, IconArrowDown, IconTrashCan, useConfirm, useToast } from "@probo/ui";
import { use, useRef } from "react";
import { useFragment, useMutation } from "react-relay";
import { useNavigate } from "react-router";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { DocumentActionsDropdown_archiveMutation } from "#/__generated__/core/DocumentActionsDropdown_archiveMutation.graphql";
import type { DocumentActionsDropdown_deleteDocumentDraftMutation } from "#/__generated__/core/DocumentActionsDropdown_deleteDocumentDraftMutation.graphql";
import type { DocumentActionsDropdown_documentFragment$key } from "#/__generated__/core/DocumentActionsDropdown_documentFragment.graphql";
import type { DocumentActionsDropdown_exportVersionMutation } from "#/__generated__/core/DocumentActionsDropdown_exportVersionMutation.graphql";
import type { DocumentActionsDropdown_unarchiveMutation } from "#/__generated__/core/DocumentActionsDropdown_unarchiveMutation.graphql";
import type { DocumentActionsDropdown_versionFragment$key } from "#/__generated__/core/DocumentActionsDropdown_versionFragment.graphql";
import { PdfDownloadDialog, type PdfDownloadDialogRef } from "#/components/documents/PdfDownloadDialog";
import { DocumentsConnectionKey, useDeleteDocumentMutation } from "#/hooks/graph/DocumentGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

const documentFragment = graphql`
  fragment DocumentActionsDropdown_documentFragment on Document {
    id
    status
    canArchive: permission(action: "core:document:archive")
    canUnarchive: permission(action: "core:document:unarchive")
    canDelete: permission(action: "core:document:delete")
    canDeleteDraft: permission(action: "core:document:delete-draft")
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
        canArchive: permission(action: "core:document:archive")
        canUnarchive: permission(action: "core:document:unarchive")
        canDelete: permission(action: "core:document:delete")
      }
    }
  }
`;

const deleteDocumentDraftMutation = graphql`
  mutation DocumentActionsDropdown_deleteDocumentDraftMutation(
    $input: DeleteDocumentDraftInput!
  ) {
    deleteDocumentDraft(input: $input) {
      document {
        id
        status
      }
    }
  }
`;

const versionFragment = graphql`
  fragment DocumentActionsDropdown_versionFragment on DocumentVersion {
    id
    title
    major
    minor
    status
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
  onVersionChanged: () => void;
}) {
  const { documentFragmentRef, versionFragmentRef, onVersionChanged } = props;

  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { __ } = useTranslate();
  const { email: defaultEmail } = use(CurrentUser);
  const pdfDownloadDialogRef = useRef<PdfDownloadDialogRef>(null);
  const confirm = useConfirm();
  const { toast } = useToast();

  const document = useFragment<DocumentActionsDropdown_documentFragment$key>(documentFragment, documentFragmentRef);
  const version = useFragment<DocumentActionsDropdown_versionFragment$key>(versionFragment, versionFragmentRef);

  const [deleteDocument, isDeleting] = useDeleteDocumentMutation();
  const [archiveDocument, isArchiving]
    = useMutation<DocumentActionsDropdown_archiveMutation>(archiveDocumentMutation);
  const [unarchiveDocument, isUnarchiving]
    = useMutation<DocumentActionsDropdown_unarchiveMutation>(unarchiveDocumentMutation);
  const [deleteDocumentDraft, isDeletingDraft]
    = useMutation<DocumentActionsDropdown_deleteDocumentDraftMutation>(deleteDocumentDraftMutation);
  const [exportDocumentVersion, isExporting]
    = useMutation<DocumentActionsDropdown_exportVersionMutation>(exportDocumentVersionMutation);

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
          version.title,
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

  const handleDeleteDraft = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteDocumentDraft({
            variables: { input: { documentId: document.id } },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({ title: __("Error"), description: formatError(__("Failed to delete draft"), errors), variant: "error" });
              } else {
                toast({ title: __("Success"), description: __("Draft deleted successfully."), variant: "success" });
                onVersionChanged();
                void navigate(`/organizations/${organizationId}/documents/${document.id}/description`);
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
        message: __("This will delete the current draft and revert to the last published version."),
        variant: "danger",
        label: __("Delete draft"),
      },
    );
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
          version.title,
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
          link.download = `${version.title}-v${version.major}.${version.minor}.pdf`;
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
        <DropdownItem
          onClick={() => pdfDownloadDialogRef.current?.open()}
          icon={IconArrowDown}
          disabled={isExporting}
        >
          {__("Download PDF")}
        </DropdownItem>
        {document.canDeleteDraft && version.status === "DRAFT" && !(version.major === 0 && version.minor === 1) && (
          <DropdownItem
            icon={IconTrashCan}
            disabled={isDeletingDraft}
            onClick={handleDeleteDraft}
          >
            {__("Delete draft")}
          </DropdownItem>
        )}
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
