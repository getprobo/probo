import { sprintf } from "@probo/helpers";
import { useList } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Button, Card, Checkbox, IconArrowDown, IconCheckmark1, IconCrossLargeX, IconSignature, IconTrashCan, Tbody, Th, Thead, Tr, useConfirm } from "@probo/ui";
import { type ComponentProps, use, useRef } from "react";
import { usePaginationFragment } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { DocumentListFragment$key } from "#/__generated__/core/DocumentListFragment.graphql";
import type { DocumentsListQuery } from "#/__generated__/core/DocumentsListQuery.graphql";
import { BulkExportDialog, type BulkExportDialogRef } from "#/components/documents/BulkExportDialog";
import { type Order, SortableTable, SortableTh } from "#/components/SortableTable";
import { useBulkDeleteDocumentsMutation, useBulkExportDocumentsMutation } from "#/hooks/graph/DocumentGraph";
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
  ) {
    documents(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "DocumentsListQuery_documents" filters: ["orderBy"]) {
      __id
      edges {
        node {
          id
          canUpdate: permission(action: "core:document:update")
          canDelete: permission(action: "core:document:delete")
          canRequestSignatures: permission(
            action: "core:document-version:request-signature"
          )
          ...DocumentListItemFragment
        }
      }
    }
  }
`;

export function DocumentList(props: {
  fKey: DocumentListFragment$key;
  onConnectionIdChange: (connectionId: string) => void;
}) {
  const { fKey, onConnectionIdChange } = props;

  const organizationId = useOrganizationId();
  const { email: defaultEmail } = use(CurrentUser);
  const bulkExportDialogRef = useRef<BulkExportDialogRef>(null);
  const { __ } = useTranslate();

  const pagination = usePaginationFragment<DocumentsListQuery, DocumentListFragment$key>(
    fragment,
    fKey,
  );

  const documents = pagination.data.documents.edges
    .map(({ node }) => node);
  const connectionId = pagination.data.documents.__id;

  const [bulkDeleteDocuments] = useBulkDeleteDocumentsMutation();
  const [bulkExportDocuments, isBulkExporting]
    = useBulkExportDocumentsMutation();
  const { list: selection, toggle, clear, reset } = useList<string>([]);
  const confirm = useConfirm();

  const canDeleteAny = documents.some(({ canDelete }) => canDelete);
  const canUpdateAny = documents.some(({ canUpdate }) => canUpdate);
  const canRequestAnySignatures = documents.some(
    ({ canRequestSignatures }) => canRequestSignatures,
  );
  const hasAnyAction = canDeleteAny || canUpdateAny;

  const handleBulkDelete = () => {
    const documentCount = selection.length;
    confirm(
      () =>
        bulkDeleteDocuments({
          variables: {
            input: { documentIds: selection },
          },
        }).then(() => {
          clear();
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete %s document%s. This action cannot be undone.",
          ),
          documentCount,
          documentCount > 1 ? "s" : "",
        ),
      },
    );
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
      ...(options.withWatermark
        && options.watermarkEmail && { watermarkEmail: options.watermarkEmail }),
    };

    await bulkExportDocuments({
      variables: { input },
    });
    clear();
  };

  const handleOrderChange = (order: Order) => {
    onConnectionIdChange(
      ConnectionHandler.getConnectionID(
        organizationId,
        "DocumentsListQuery_documents",
        { orderBy: order },
      ),
    );
  };

  return documents.length > 0
    ? (
        <SortableTable
          {...pagination}
          refetch={pagination.refetch as ComponentProps<typeof SortableTable>["refetch"]}
        >
          <Thead>
            {selection.length === 0
              ? (
                  <Tr>
                    <Th className="w-18">
                      <Checkbox
                        checked={
                          selection.length === documents.length
                          && documents.length > 0
                        }
                        onChange={() => reset(documents.map(d => d.id))}
                      />
                    </Th>
                    <SortableTh field="TITLE" className="min-w-0" onOrderChange={handleOrderChange}>
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
                    <Th className="w-20">{__("Signatures")}</Th>
                    {hasAnyAction && <Th className="w-18"></Th>}
                  </Tr>
                )
              : (
                  <Tr>
                    <Th colspan={10} compact>
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
                          {canUpdateAny && (
                            <PublishDocumentsDialog
                              documentIds={selection}
                              onSave={clear}
                            >
                              <Button
                                icon={IconCheckmark1}
                                className="py-0.5 px-2 text-xs h-6 min-h-6"
                              >
                                {__("Publish")}
                              </Button>
                            </PublishDocumentsDialog>
                          )}
                          {canRequestAnySignatures && (
                            <SignatureDocumentsDialog
                              documentIds={selection}
                              onSave={clear}
                            >
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
              {__("No documents yet")}
            </h3>
            <p className="text-txt-tertiary mb-4">
              {__("Create your first document to get started.")}
            </p>
          </div>
        </Card>
      );
}
