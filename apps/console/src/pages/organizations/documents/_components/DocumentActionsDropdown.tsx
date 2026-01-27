import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { ActionDropdown, DropdownItem, IconArrowDown, IconPencil, IconTrashCan, useConfirm } from "@probo/ui";
import { use, useRef } from "react";
import { useFragment } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { DocumentActionsDropdown_documentFragment$key } from "#/__generated__/core/DocumentActionsDropdown_documentFragment.graphql";
import type { DocumentActionsDropdown_versionFragment$key } from "#/__generated__/core/DocumentActionsDropdown_versionFragment.graphql";
import type { DocumentActionsDropdownn_exportVersionMutation } from "#/__generated__/core/DocumentActionsDropdownn_exportVersionMutation.graphql";
import { PdfDownloadDialog, type PdfDownloadDialogRef } from "#/components/documents/PdfDownloadDialog";
import { useDeleteDocumentMutation, useDeleteDraftDocumentVersionMutation } from "#/hooks/graph/DocumentGraph";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

import UpdateVersionDialog from "./UpdateVersionDialog";

const documentFragment = graphql`
  fragment DocumentActionsDropdown_documentFragment on Document {
    id
    title
    canUpdate: permission(action: "core:document:update")
    canDelete: permission(action: "core:document:delete")
    versions(first: 20) {
      __id
      totalCount
    }
    ...UpdateVersionDialogFragment
  }
`;

const versionFragment = graphql`
  fragment DocumentActionsDropdown_versionFragment on DocumentVersion {
    id
    version
    status
    canDeleteDraft: permission(action: "core:document-version:delete-draft")
  }
`;

const exportDocumentVersionMutation = graphql`
  mutation DocumentActionsDropdownn_exportVersionMutation(
    $input: ExportDocumentVersionPDFInput!
  ) {
    exportDocumentVersionPDF(input: $input) {
      data
    }
  }
`;

export function DocumentActionsDropdownn(props: {
  documentFragmentRef: DocumentActionsDropdown_documentFragment$key;
  versionFragmentRef: DocumentActionsDropdown_versionFragment$key;
}) {
  const { documentFragmentRef, versionFragmentRef } = props;

  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { __ } = useTranslate();
  const { email: defaultEmail } = use(CurrentUser);
  const updateDialogRef = useRef<{ open: () => void }>(null);
  const pdfDownloadDialogRef = useRef<PdfDownloadDialogRef>(null);
  const confirm = useConfirm();

  const document = useFragment<DocumentActionsDropdown_documentFragment$key>(documentFragment, documentFragmentRef);
  const version = useFragment<DocumentActionsDropdown_versionFragment$key>(versionFragment, versionFragmentRef);

  const isDraft = version.status === "DRAFT";

  const [deleteDocument, isDeleting] = useDeleteDocumentMutation();
  const [deleteDraftDocumentVersion, isDeletingDraft]
    = useDeleteDraftDocumentVersionMutation();
  const [exportDocumentVersion, isExporting]
    = useMutationWithToasts<DocumentActionsDropdownn_exportVersionMutation>(
      exportDocumentVersionMutation,
      {
        successMessage: __("PDF download started."),
        errorMessage: __("Failed to generate PDF"),
      },
    );

  const handleDelete = () => {
    confirm(
      () =>
        deleteDocument({
          variables: {
            input: { documentId: document.id },
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
    confirm(
      () =>
        deleteDraftDocumentVersion({
          variables: {
            input: { documentVersionId: version.id },
            connections: [document.versions.__id],
          },
          onSuccess() {
            window.location.href = `/organizations/${organizationId}/documents/${document.id}`;
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the draft version %s of \"%s\". This action cannot be undone.",
          ),
          version.version,
          document.title,
        ),
      },
    );
  };

  const handleExportDocumentVersion = async (options: {
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

    await exportDocumentVersion({
      variables: { input },
      onCompleted: (data, errors) => {
        if (errors?.length) {
          return;
        }

        if (data.exportDocumentVersionPDF) {
          const link = window.document.createElement("a");
          link.href = data.exportDocumentVersionPDF.data;
          link.download = `${document.title}-v${version.version}.pdf`;
          window.document.body.appendChild(link);
          link.click();
          window.document.body.removeChild(link);
        }
      },
    });
  };

  return (
    <>
      <UpdateVersionDialog
        ref={updateDialogRef}
        fKey={document}
      />
      <PdfDownloadDialog
        ref={pdfDownloadDialogRef}
        onDownload={options => void handleExportDocumentVersion(options)}
        isLoading={isExporting}
        defaultEmail={defaultEmail}
      />
      <ActionDropdown variant="secondary">
        {document.canUpdate && (
          <DropdownItem
            onClick={() => updateDialogRef.current?.open()}
            icon={IconPencil}
          >
            {isDraft ? __("Edit draft document") : __("Create new draft")}
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
