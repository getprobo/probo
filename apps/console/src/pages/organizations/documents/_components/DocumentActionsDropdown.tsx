// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { formatError } from "@probo/helpers";
import { ActionDropdown, DropdownItem, IconArchive, IconArrowDown, IconTrashCan, useConfirm, useToast } from "@probo/ui";
import { use, useRef } from "react";
import { useTranslation } from "react-i18next";
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
import { DocumentsConnectionKey } from "#/hooks/graph/DocumentGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

import { DeleteDocumentDialog, type DeleteDocumentDialogRef } from "./DeleteDocumentDialog";

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
  const { t } = useTranslation();
  const { email: defaultEmail } = use(CurrentUser);
  const pdfDownloadDialogRef = useRef<PdfDownloadDialogRef>(null);
  const deleteDocumentDialogRef = useRef<DeleteDocumentDialogRef>(null);
  const confirm = useConfirm();
  const { toast } = useToast();

  const document = useFragment<DocumentActionsDropdown_documentFragment$key>(documentFragment, documentFragmentRef);
  const version = useFragment<DocumentActionsDropdown_versionFragment$key>(versionFragment, versionFragmentRef);

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
                toast({ title: t("documentActions.errors.title"), description: formatError(t("documentActions.errors.archive"), errors), variant: "error" });
              } else {
                toast({ title: t("documentActions.messages.successTitle"), description: t("documentActions.messages.archived"), variant: "success" });
              }
              resolve();
            },
            onError(error) {
              toast({ title: t("documentActions.errors.title"), description: error.message, variant: "error" });
              resolve();
            },
          });
        }),
      {
        message: t("documentActions.confirmations.archive", { title: version.title }),
        variant: "danger",
        label: t("documentActions.actions.archive"),
      },
    );
  };

  const handleUnarchive = () => {
    unarchiveDocument({
      variables: { input: { documentId: document.id } },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({ title: t("documentActions.errors.title"), description: formatError(t("documentActions.errors.unarchive"), errors), variant: "error" });
        } else {
          toast({ title: t("documentActions.messages.successTitle"), description: t("documentActions.messages.unarchived"), variant: "success" });
        }
      },
      onError(error) {
        toast({ title: t("documentActions.errors.title"), description: error.message, variant: "error" });
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
                toast({ title: t("documentActions.errors.title"), description: formatError(t("documentActions.errors.deleteDraft"), errors), variant: "error" });
              } else {
                toast({ title: t("documentActions.messages.successTitle"), description: t("documentActions.messages.draftDeleted"), variant: "success" });
                onVersionChanged();
                void navigate(`/organizations/${organizationId}/documents/${document.id}/description`);
              }
              resolve();
            },
            onError(error) {
              toast({ title: t("documentActions.errors.title"), description: error.message, variant: "error" });
              resolve();
            },
          });
        }),
      {
        message: t("documentActions.confirmations.deleteDraft"),
        variant: "danger",
        label: t("documentActions.actions.deleteDraft"),
      },
    );
  };

  const documentsConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    DocumentsConnectionKey,
    { orderBy: { direction: "ASC", field: "TITLE" } },
  );

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
            title: t("documentActions.errors.title"),
            description: errors[0]?.message || t("documentActions.errors.generatePdf"),
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
        toast({ title: t("documentActions.errors.title"), description: error.message, variant: "error" });
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
      <DeleteDocumentDialog
        ref={deleteDocumentDialogRef}
        documentId={document.id}
        documentTitle={version.title}
        connections={[documentsConnectionId]}
        onSuccess={() => void navigate(`/organizations/${organizationId}/documents`)}
      />
      <ActionDropdown variant="secondary">
        <DropdownItem
          onClick={() => pdfDownloadDialogRef.current?.open()}
          icon={IconArrowDown}
          disabled={isExporting}
        >
          {t("documentActions.actions.downloadPdf")}
        </DropdownItem>
        {document.canDeleteDraft && version.status === "DRAFT" && !(version.major === 0 && version.minor === 1) && (
          <DropdownItem
            icon={IconTrashCan}
            disabled={isDeletingDraft}
            onClick={handleDeleteDraft}
          >
            {t("documentActions.actions.deleteDraft")}
          </DropdownItem>
        )}
        {document.canArchive && document.status === "ACTIVE" && (
          <DropdownItem
            icon={IconArchive}
            disabled={isArchiving}
            onClick={handleArchive}
          >
            {t("documentActions.actions.archiveDocument")}
          </DropdownItem>
        )}
        {document.canUnarchive && document.status === "ARCHIVED" && (
          <DropdownItem
            icon={IconArchive}
            disabled={isUnarchiving}
            onClick={handleUnarchive}
          >
            {t("documentActions.actions.unarchiveDocument")}
          </DropdownItem>
        )}
        {document.canDelete && (
          <DropdownItem
            variant="danger"
            icon={IconTrashCan}
            onClick={() => deleteDocumentDialogRef.current?.open()}
          >
            {t("documentActions.actions.deleteDocument")}
          </DropdownItem>
        )}
      </ActionDropdown>
    </>
  );
}
