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

import { documentTypes, getDocumentTypeLabel, sprintf } from "@probo/helpers";
import { useList } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Card, Checkbox, IconArchive, IconArrowDown, IconCrossLargeX, IconSignature, IconTrashCan, IconUpload, Option, Select, Tbody, Th, Thead, Tr, useConfirm } from "@probo/ui";
import { type ComponentProps, use, useEffect, useRef, useState, useTransition } from "react";
import { usePaginationFragment } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { DocumentListBulkArchiveMutation } from "#/__generated__/core/DocumentListBulkArchiveMutation.graphql";
import type { DocumentListBulkUnarchiveMutation } from "#/__generated__/core/DocumentListBulkUnarchiveMutation.graphql";
import type { DocumentListFragment$key } from "#/__generated__/core/DocumentListFragment.graphql";
import type { DocumentOrderField, DocumentsListQuery, DocumentType } from "#/__generated__/core/DocumentsListQuery.graphql";
import { BulkExportDialog, type BulkExportDialogRef } from "#/components/documents/BulkExportDialog";
import { type Order, SortableTable, SortableTh } from "#/components/SortableTable";
import { useBulkDeleteDocumentsMutation, useBulkExportDocumentsMutation } from "#/hooks/graph/DocumentGraph";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

import { DocumentListItem } from "./DocumentListItem";
import { PublishDocumentsDialog } from "./PublishDocumentsDialog";
import { SignatureDocumentsDialog } from "./SignatureDocumentsDialog";

const fragment = graphql`
  fragment DocumentListFragment on Organization
  @refetchable(queryName: "DocumentsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "DocumentOrder"
      defaultValue: { field: TITLE, direction: ASC }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    status: { type: "[DocumentStatus!]", defaultValue: [ACTIVE] }
    documentTypes: { type: "[DocumentType!]", defaultValue: null }
  ) {
    documents(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: { status: $status documentTypes: $documentTypes }
    ) @connection(key: "DocumentsListQuery_documents" filters: ["orderBy", "filter"]) {
      __id
      edges {
        node {
          id
          canUpdate: permission(action: "core:document:update")
          canDelete: permission(action: "core:document:delete")
          canRequestSignatures: permission(
            action: "core:document-version:request-signature"
          )
          canArchive: permission(action: "core:document:archive")
          canUnarchive: permission(action: "core:document:unarchive")
          canSendSigningNotifications: permission(
            action: "core:document:send-signing-notifications"
          )
          ...DocumentListItemFragment
        }
      }
    }
  }
`;

const bulkArchiveMutation = graphql`
  mutation DocumentListBulkArchiveMutation($input: BulkArchiveDocumentsInput!) {
    bulkArchiveDocuments(input: $input) {
      documents {
        id
        status
        archivedAt
        canUpdate: permission(action: "core:document:update")
        canArchive: permission(action: "core:document:archive")
        canUnarchive: permission(action: "core:document:unarchive")
      }
    }
  }
`;

const bulkUnarchiveMutation = graphql`
  mutation DocumentListBulkUnarchiveMutation($input: BulkUnarchiveDocumentsInput!) {
    bulkUnarchiveDocuments(input: $input) {
      documents {
        id
        status
        archivedAt
        canUpdate: permission(action: "core:document:update")
        canArchive: permission(action: "core:document:archive")
        canUnarchive: permission(action: "core:document:unarchive")
      }
    }
  }
`;

export function DocumentList(props: {
  fKey: DocumentListFragment$key;
  onConnectionIdChange: (connectionId: string) => void;
  onCanSendNotificationsChange?: (can: boolean) => void;
  tab: "ACTIVE" | "ARCHIVED";
}) {
  const { fKey, onConnectionIdChange, onCanSendNotificationsChange, tab } = props;

  const organizationId = useOrganizationId();
  const { email: defaultEmail } = use(CurrentUser);
  const bulkExportDialogRef = useRef<BulkExportDialogRef>(null);
  const { __ } = useTranslate();

  const pagination = usePaginationFragment<DocumentsListQuery, DocumentListFragment$key>(
    fragment,
    fKey,
  );

  const [documentTypeFilter, setDocumentTypeFilter] = useState<DocumentType | null>(null);
  const [isPending, startTransition] = useTransition();

  const refetch = pagination.refetch;
  useEffect(() => {
    startTransition(() => {
      refetch(
        { status: [tab], documentTypes: documentTypeFilter ? [documentTypeFilter] : null },
        { fetchPolicy: "store-and-network" },
      );
    });
  }, [tab, refetch, documentTypeFilter]);

  const documents = pagination.data.documents.edges.map(({ node }) => node);
  const connectionId = pagination.data.documents.__id;

  const [bulkDeleteDocuments] = useBulkDeleteDocumentsMutation();
  const [bulkExportDocuments, isBulkExporting] = useBulkExportDocumentsMutation();
  const [bulkArchiveDocuments, isBulkArchiving] = useMutationWithToasts<DocumentListBulkArchiveMutation>(
    bulkArchiveMutation,
    { successMessage: __("Documents archived successfully."), errorMessage: __("Failed to archive documents") },
  );
  const [bulkUnarchiveDocuments, isBulkUnarchiving] = useMutationWithToasts<DocumentListBulkUnarchiveMutation>(
    bulkUnarchiveMutation,
    { successMessage: __("Documents unarchived successfully."), errorMessage: __("Failed to unarchive documents") },
  );
  const { list: selection, toggle, clear, reset } = useList<string>([]);
  const confirm = useConfirm();

  const canDeleteAny = documents.some(({ canDelete }) => canDelete);
  const canUpdateAny = documents.some(({ canUpdate }) => canUpdate);
  const canRequestAnySignatures = documents.some(({ canRequestSignatures }) => canRequestSignatures);
  const canArchiveAny = documents.some(({ canArchive }) => canArchive);
  const canUnarchiveAny = documents.some(({ canUnarchive }) => canUnarchive);
  const canSendAnySignatureNotifications = documents.some(
    ({ canSendSigningNotifications }) => canSendSigningNotifications,
  );
  const hasAnyAction = tab === "ARCHIVED" ? canUnarchiveAny || canDeleteAny : canDeleteAny || canUpdateAny;

  useEffect(() => {
    onConnectionIdChange(connectionId);
  }, [connectionId, onConnectionIdChange]);

  useEffect(() => {
    onCanSendNotificationsChange?.(canSendAnySignatureNotifications);
  }, [canSendAnySignatureNotifications, onCanSendNotificationsChange]);

  const handleDocumentTypeFilterChange = (value: string) => {
    const newType = value === "ALL" ? null : (value as DocumentType);
    clear();
    setDocumentTypeFilter(newType);
    onConnectionIdChange(
      ConnectionHandler.getConnectionID(
        organizationId,
        "DocumentsListQuery_documents",
        {
          orderBy: { direction: "ASC", field: "TITLE" },
          filter: { status: [tab], documentTypes: newType ? [newType] : null },
        },
      ),
    );
  };

  const handleBulkDelete = () => {
    const documentCount = selection.length;
    confirm(
      () =>
        bulkDeleteDocuments({
          variables: { input: { documentIds: selection } },
          updater: (store) => {
            const conn = store.get(connectionId);
            if (conn) {
              selection.forEach(id => ConnectionHandler.deleteNode(conn, id));
            }
          },
        }).then(() => {
          clear();
        }),
      {
        message: sprintf(
          __("This will permanently delete %s document%s. This action cannot be undone."),
          documentCount,
          documentCount > 1 ? "s" : "",
        ),
      },
    );
  };

  const handleBulkArchive = () => {
    void bulkArchiveDocuments({
      variables: { input: { documentIds: selection } },
      updater: (store) => {
        const conn = store.get(connectionId);
        if (conn) {
          selection.forEach(id => ConnectionHandler.deleteNode(conn, id));
        }
      },
      onSuccess: clear,
    });
  };

  const handleBulkUnarchive = () => {
    void bulkUnarchiveDocuments({
      variables: { input: { documentIds: selection } },
      updater: (store) => {
        const conn = store.get(connectionId);
        if (conn) {
          selection.forEach(id => ConnectionHandler.deleteNode(conn, id));
        }
      },
      onSuccess: clear,
    });
  };

  const handleBulkExport = async (options: {
    withWatermark: boolean;
    withSignatures: boolean;
    watermarkEmail?: string;
  }) => {
    const input = {
      documentIds: selection,
      withWatermark: options.withWatermark,
      withSignatures: options.withSignatures,
      ...(options.withWatermark && options.watermarkEmail && { watermarkEmail: options.watermarkEmail }),
    };
    await bulkExportDocuments({ variables: { input } });
    clear();
  };

  const handleOrderChange = (order: Order) => {
    onConnectionIdChange(
      ConnectionHandler.getConnectionID(
        organizationId,
        "DocumentsListQuery_documents",
        {
          orderBy: order,
          filter: { status: [tab], documentTypes: documentTypeFilter ? [documentTypeFilter] : null },
        },
      ),
    );
  };

  const refetchWithFilters: ComponentProps<typeof SortableTable>["refetch"] = ({ order }) => {
    pagination.refetch({
      order: { direction: order.direction, field: order.field as DocumentOrderField },
      status: [tab],
      documentTypes: documentTypeFilter ? [documentTypeFilter] : null,
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Select
          value={documentTypeFilter ?? "ALL"}
          onValueChange={handleDocumentTypeFilterChange}
        >
          <Option value="ALL">{__("All types")}</Option>
          {documentTypes.map(type => (
            <Option key={type} value={type}>
              {getDocumentTypeLabel(__, type) ?? type}
            </Option>
          ))}
        </Select>
      </div>
      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        {documents.length > 0
          ? (
              <SortableTable
                {...pagination}
                refetch={refetchWithFilters}
              >
                <Thead>
                  {selection.length === 0
                    ? (
                        <Tr>
                          <Th className="w-18">
                            <Checkbox
                              checked={selection.length === documents.length && documents.length > 0}
                              onChange={() => reset(documents.map(d => d.id))}
                            />
                          </Th>
                          <SortableTh field="TITLE" className="min-w-0 pr-12" onOrderChange={handleOrderChange}>
                            {__("Name")}
                          </SortableTh>
                          <Th className="w-24">{__("Status")}</Th>
                          <Th className="w-20">{__("Version")}</Th>
                          <SortableTh field="DOCUMENT_TYPE" className="w-28" onOrderChange={handleOrderChange}>
                            {__("Type")}
                          </SortableTh>
                          <Th className="w-32">{__("Classification")}</Th>
                          <Th className="w-60">{__("Approvers")}</Th>
                          <Th className="w-60">{__("Last update")}</Th>
                          <Th className="w-20">{__("Approvals")}</Th>
                          <Th className="w-20">{__("Signatures")}</Th>
                          {hasAnyAction && <Th className="w-18"></Th>}
                        </Tr>
                      )
                    : (
                        <Tr>
                          <Th colspan={hasAnyAction ? 11 : 10} compact>
                            <div className="flex justify-between items-center h-8">
                              <div className="flex gap-2 items-center">
                                {sprintf(__("%s documents selected"), selection.length)}
                                {" "}
                                -
                                <button
                                  onClick={clear}
                                  className="flex gap-1 items-center hover:text-txt-primary"
                                >
                                  <IconCrossLargeX size={12} />
                                  {__("Clear selection")}
                                </button>
                              </div>
                              <div className="flex gap-2 items-center">
                                {tab === "ARCHIVED"
                                  ? (
                                      <>
                                        {canUnarchiveAny && (
                                          <Button
                                            variant="secondary"
                                            icon={IconArchive}
                                            onClick={handleBulkUnarchive}
                                            disabled={isBulkUnarchiving}
                                            className="py-0.5 px-2 text-xs h-6 min-h-6"
                                          >
                                            {__("Unarchive")}
                                          </Button>
                                        )}
                                        {canDeleteAny && (
                                          <Button
                                            variant="danger"
                                            icon={IconTrashCan}
                                            onClick={handleBulkDelete}
                                            className="py-0.5 px-2 text-xs h-6 min-h-6"
                                          >
                                            {__("Delete")}
                                          </Button>
                                        )}
                                      </>
                                    )
                                  : (
                                      <>
                                        {canUpdateAny && (
                                          <PublishDocumentsDialog documentIds={selection} onSave={clear}>
                                            <Button
                                              icon={IconUpload}
                                              className="py-0.5 px-2 text-xs h-6 min-h-6"
                                            >
                                              {__("Publish")}
                                            </Button>
                                          </PublishDocumentsDialog>
                                        )}
                                        {canRequestAnySignatures && (
                                          <SignatureDocumentsDialog documentIds={selection} onSave={clear}>
                                            <Button
                                              variant="secondary"
                                              icon={IconSignature}
                                              className="py-0.5 px-2 text-xs h-6 min-h-6"
                                            >
                                              {__("Request signature")}
                                            </Button>
                                          </SignatureDocumentsDialog>
                                        )}
                                        <BulkExportDialog
                                          ref={bulkExportDialogRef}
                                          onExport={handleBulkExport}
                                          isLoading={isBulkExporting}
                                          defaultEmail={defaultEmail}
                                          selectedCount={selection.length}
                                        >
                                          <Button
                                            variant="secondary"
                                            icon={IconArrowDown}
                                            className="py-0.5 px-2 text-xs h-6 min-h-6"
                                          >
                                            {__("Export")}
                                          </Button>
                                        </BulkExportDialog>
                                        {canArchiveAny && (
                                          <Button
                                            variant="secondary"
                                            icon={IconArchive}
                                            onClick={handleBulkArchive}
                                            disabled={isBulkArchiving}
                                            className="py-0.5 px-2 text-xs h-6 min-h-6"
                                          >
                                            {__("Archive")}
                                          </Button>
                                        )}
                                        {canDeleteAny && (
                                          <Button
                                            variant="danger"
                                            icon={IconTrashCan}
                                            onClick={handleBulkDelete}
                                            className="py-0.5 px-2 text-xs h-6 min-h-6"
                                          >
                                            {__("Delete")}
                                          </Button>
                                        )}
                                      </>
                                    )}
                              </div>
                            </div>
                          </Th>
                        </Tr>
                      )}
                </Thead>
                <Tbody>
                  {documents.map(document => (
                    <DocumentListItem
                      checked={selection.includes(document.id)}
                      onCheck={() => toggle(document.id)}
                      key={document.id}
                      fragmentRef={document}
                      connectionId={connectionId}
                      hasAnyAction={hasAnyAction}
                    />
                  ))}
                </Tbody>
              </SortableTable>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {tab === "ARCHIVED" ? __("No archived documents") : __("No documents yet")}
                  </h3>
                  {tab !== "ARCHIVED" && (
                    <p className="text-txt-tertiary mb-4">
                      {__("Create your first document to get started.")}
                    </p>
                  )}
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}
