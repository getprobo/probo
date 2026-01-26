import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { ActionDropdown, DropdownItem, IconArrowDown, IconPencil, IconTrashCan, useConfirm } from "@probo/ui";
import { use, useRef } from "react";
import { loadQuery, useFragment } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { DocumentActionsDropdownFragment$key } from "#/__generated__/core/DocumentActionsDropdownFragment.graphql";
import type { DocumentActionsDropdownn_exportVersionMutation } from "#/__generated__/core/DocumentActionsDropdownn_exportVersionMutation.graphql";
import type { DocumentLayoutQuery$data } from "#/__generated__/core/DocumentLayoutQuery.graphql";
import { PdfDownloadDialog, type PdfDownloadDialogRef } from "#/components/documents/PdfDownloadDialog";
import { coreEnvironment } from "#/environments";
import { useDeleteDocumentMutation, useDeleteDraftDocumentVersionMutation } from "#/hooks/graph/DocumentGraph";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";
import type { NodeOf } from "#/types";

import { documentLayoutQuery } from "../DocumentLayout";

import UpdateVersionDialog from "./UpdateVersionDialog";

const fragment = graphql`
  fragment DocumentActionsDropdownFragment on Document {
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
  currentVersion: NodeOf<Extract<DocumentLayoutQuery$data["document"], { __typename: "Document" }>["versions"]>;
  fKey: DocumentActionsDropdownFragment$key;
  isDraft: boolean;
}) {
  const { currentVersion, fKey, isDraft } = props;

  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { __ } = useTranslate();
  const { email: defaultEmail } = use(CurrentUser);
  const updateDialogRef = useRef<{ open: () => void }>(null);
  const pdfDownloadDialogRef = useRef<PdfDownloadDialogRef>(null);
  const confirm = useConfirm();

  const document = useFragment<DocumentActionsDropdownFragment$key>(fragment, fKey);
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
        new Promise<void>((resolve) => {
          void deleteDocument({
            variables: {
              input: { documentId: document.id },
            },
            onSuccess() {
              void navigate(`/organizations/${organizationId}/documents`);
              resolve();
            },
            onError: () => resolve(),
          });
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
        new Promise<void>((resolve) => {
          void deleteDraftDocumentVersion({
            variables: {
              input: { documentVersionId: currentVersion.id },
              connections: [document.versions.__id],
            },
            onSuccess() {
              loadQuery(
                coreEnvironment,
                documentLayoutQuery,
                { documentId: document.id },
                { fetchPolicy: "network-only" },
              );

              resolve();
            },
            onError: () => resolve(),
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the draft version %s of \"%s\". This action cannot be undone.",
          ),
          currentVersion.version,
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
      documentVersionId: currentVersion.id,
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
          link.download = `${document.title}-v${currentVersion.version}.pdf`;
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
          && currentVersion.canDeleteDraft && (
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
