import { useTranslate } from "@probo/i18n";
import {
  PageHeader,
  Tbody,
  Thead,
  Tr,
  Th,
  Td,
  Avatar,
  Badge,
  IconTrashCan,
  Button,
  IconPlusLarge,
  useConfirm,
  ActionDropdown,
  DropdownItem,
  IconBell2,
  Checkbox,
  IconCrossLargeX,
  IconSignature,
  IconCheckmark1,
  IconArrowDown,
  Card,
  Spinner,
} from "@probo/ui";
import {
  useFragment,
  usePaginationFragment,
  useLazyLoadQuery,
} from "react-relay";
import { useRef, useEffect, useState } from "react";
import { graphql } from "relay-runtime";
import {
  documentsQuery,
  useDeleteDocumentMutation,
  useSendSigningNotificationsMutation,
  useBulkDeleteDocumentsMutation,
  useBulkExportDocumentsMutation,
} from "/hooks/graph/DocumentGraph";
import type { DocumentsPageUserEmailQuery } from "./__generated__/DocumentsPageUserEmailQuery.graphql";
import type { DocumentGraphListQuery } from "/hooks/graph/__generated__/DocumentGraphListQuery.graphql";
import type { DocumentsPageListFragment$key } from "./__generated__/DocumentsPageListFragment.graphql";
import type { DocumentsPageRequestedListFragment$key } from "./__generated__/DocumentsPageRequestedListFragment.graphql";
import { useList, usePageTitle } from "@probo/hooks";
import {
  sprintf,
  getDocumentTypeLabel,
  getDocumentClassificationLabel,
  formatDate,
} from "@probo/helpers";
import { CreateDocumentDialog } from "./dialogs/CreateDocumentDialog";
import type { DocumentsPageRowFragment$key } from "./__generated__/DocumentsPageRowFragment.graphql";
import { SortableTable, SortableTh } from "/components/SortableTable";
import { PublishDocumentsDialog } from "./dialogs/PublishDocumentsDialog.tsx";
import { SignatureDocumentsDialog } from "./dialogs/SignatureDocumentsDialog.tsx";
import {
  BulkExportDialog,
  type BulkExportDialogRef,
} from "/components/documents/BulkExportDialog";
import { Authorized, isAuthorized, fetchPermissions } from "/permissions";
import { useOrganizationId } from "/hooks/useOrganizationId";

const documentsFragment = graphql`
  fragment DocumentsPageListFragment on Organization
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
    includeSignatures: { type: "Boolean", defaultValue: false }
  ) {
    documents(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "DocumentsListQuery_documents") {
      __id
      edges {
        node {
          id
          ...DocumentsPageRowFragment @arguments(includeSignatures: $includeSignatures, useRequestedVersions: false)
        }
      }
    }
  }
`;

const requestedDocumentsFragment = graphql`
  fragment DocumentsPageRequestedListFragment on Organization
  @refetchable(queryName: "DocumentsRequestedListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "DocumentOrder"
      defaultValue: { field: TITLE, direction: ASC }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    includeSignatures: { type: "Boolean", defaultValue: false }
  ) {
    requestedDocuments(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "DocumentsRequestedListQuery_requestedDocuments") {
      __id
      edges {
        node {
          id
          ...DocumentsPageRowFragment @arguments(includeSignatures: $includeSignatures, useRequestedVersions: true)
        }
      }
    }
  }
`;

const UserEmailQuery = graphql`
  query DocumentsPageUserEmailQuery {
    viewer {
      user {
        email
      }
    }
  }
`;

function DocumentsPageContent() {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const canListDocuments = isAuthorized(organizationId, "Organization", "listDocuments");

  const queryData = useLazyLoadQuery<DocumentGraphListQuery>(
    documentsQuery,
    {
      organizationId,
      includeSignatures: false,
      useRequestedDocuments: !canListDocuments
    },
    { fetchPolicy: 'store-or-network' }
  );

  const organization = queryData.organization;

  const canViewSignatures = isAuthorized(organization.id, "DocumentVersion", "signatures");

  const regularPagination = usePaginationFragment<any, DocumentsPageListFragment$key>(
    documentsFragment,
    canListDocuments ? (organization as DocumentsPageListFragment$key) : null
  );

  const requestedPagination = usePaginationFragment<any, DocumentsPageRequestedListFragment$key>(
    requestedDocumentsFragment,
    !canListDocuments ? (organization as DocumentsPageRequestedListFragment$key) : null
  );

  const pagination = canListDocuments ? regularPagination : requestedPagination;

  const userEmailData = useLazyLoadQuery<DocumentsPageUserEmailQuery>(
    UserEmailQuery,
    {}
  );
  const defaultEmail = userEmailData.viewer.user.email;

  useEffect(() => {
    pagination.refetch({
      includeSignatures: canViewSignatures
    }, { fetchPolicy: 'network-only' });
  }, []);

  const documentsConnection = canListDocuments
    ? regularPagination.data?.documents
    : requestedPagination.data?.requestedDocuments;
  const documents = documentsConnection?.edges
    .map((edge: any) => edge.node)
    .filter(Boolean) || [];
  const connectionId = documentsConnection?.__id || "";
  const [sendSigningNotifications] = useSendSigningNotificationsMutation();
  const [bulkDeleteDocuments] = useBulkDeleteDocumentsMutation();
  const [bulkExportDocuments, isBulkExporting] =
    useBulkExportDocumentsMutation();
  const { list: selection, toggle, clear, reset } = useList<string>([]);
  const confirm = useConfirm();
  const bulkExportDialogRef = useRef<BulkExportDialogRef>(null);

  usePageTitle(__("Documents"));

  const hasAnyAction = isAuthorized(organization.id, "Document", "updateDocument") ||
    isAuthorized(organization.id, "Document", "deleteDocument");

  const handleSendSigningNotifications = () => {
    sendSigningNotifications({
      variables: {
        input: { organizationId: organization.id },
      },
    });
  };

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
            "This will permanently delete %s document%s. This action cannot be undone."
          ),
          documentCount,
          documentCount > 1 ? "s" : ""
        ),
      }
    );
  };

  const handleBulkExport = (options: {
    withWatermark: boolean;
    withSignatures: boolean;
    watermarkEmail?: string;
  }) => {
    const input = {
      documentIds: selection,
      withWatermark: options.withWatermark,
      withSignatures: options.withSignatures,
      ...(options.withWatermark &&
        options.watermarkEmail && { watermarkEmail: options.watermarkEmail }),
    };

    bulkExportDocuments({
      variables: { input },
    }).then(() => {
      clear();
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Documents")}
        description={__("Manage your organization's documents")}
      >
        <div className="flex gap-2">
          <Authorized entity="Document" action="sendSigningNotifications">
          <Button
            icon={IconBell2}
            variant="secondary"
            onClick={handleSendSigningNotifications}
          >
            {__("Send signing notifications")}
          </Button>
          </Authorized>
          <Authorized entity="Organization" action="createDocument">
            <CreateDocumentDialog
              connection={connectionId}
              trigger={<Button icon={IconPlusLarge}>{__("New document")}</Button>}
            />
          </Authorized>
        </div>
      </PageHeader>
      {documents.length > 0 ? (
        <SortableTable {...pagination}>
          <Thead>
            {selection.length === 0 ? (
              <Tr>
                <Th className="w-18">
                  <Checkbox
                    checked={
                      selection.length === documents.length &&
                      documents.length > 0
                    }
                    onChange={() => reset(documents.map((d) => d.id))}
                  />
                </Th>
                <SortableTh field="TITLE" className="min-w-0">
                  {__("Name")}
                </SortableTh>
                <Th className="w-24">{__("Status")}</Th>
                <Th className="w-20">{__("Version")}</Th>
                <SortableTh field="DOCUMENT_TYPE" className="w-28">
                  {__("Type")}
                </SortableTh>
                <Th className="w-32">{__("Classification")}</Th>
                <Th className="w-60">{__("Owner")}</Th>
                <Th className="w-60">{__("Last update")}</Th>
                {canViewSignatures && <Th className="w-20">{__("Signatures")}</Th>}
                {hasAnyAction && <Th className="w-18"></Th>}
              </Tr>
            ) : (
              <Tr>
                <Th colspan={10} compact>
                  <div className="flex justify-between items-center h-8">
                    <div className="flex gap-2 items-center">
                      {sprintf(__("%s documents selected"), selection.length)} -
                      <button
                        onClick={clear}
                        className="flex gap-1 items-center hover:text-txt-primary"
                      >
                        <IconCrossLargeX size={12} />
                        {__("Clear selection")}
                      </button>
                    </div>
                    <div className="flex gap-2 items-center">
                      <Authorized entity="Document" action="updateDocument">
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
                      </Authorized>
                      <Authorized entity="Document" action="bulkRequestSignatures">
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
                      </Authorized>
                      {canViewSignatures ? (
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
                      ) : (
                        <Button
                          variant="secondary"
                          icon={IconArrowDown}
                          className="py-0.5 px-2 text-xs h-6 min-h-6"
                          onClick={() => handleBulkExport({
                            withWatermark: true,
                            withSignatures: false,
                            watermarkEmail: defaultEmail,
                          })}
                          disabled={isBulkExporting}
                        >
                          {__("Export")}
                        </Button>
                      )}
                      <Authorized entity="Document" action="deleteDocument">
                        <Button
                          variant="danger"
                          icon={IconTrashCan}
                          onClick={handleBulkDelete}
                          className="py-0.5 px-2 text-xs h-6 min-h-6"
                        >
                          {__("Delete")}
                        </Button>
                      </Authorized>
                    </div>
                  </div>
                </Th>
              </Tr>
            )}
          </Thead>
          <Tbody>
            {documents.map((document) => (
              <DocumentRow
                checked={selection.includes(document.id)}
                onCheck={() => toggle(document.id)}
                key={document.id}
                document={document}
                organizationId={organization.id}
                connectionId={connectionId}
                hasAnyAction={hasAnyAction}
                canViewSignatures={canViewSignatures}
              />
            ))}
          </Tbody>
        </SortableTable>
      ) : (
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
      )}
    </div>
  );
}

const rowFragment = graphql`
  fragment DocumentsPageRowFragment on Document
  @argumentDefinitions(
    includeSignatures: { type: "Boolean", defaultValue: false }
    useRequestedVersions: { type: "Boolean", defaultValue: false }
  ) {
    id
    title
    description
    documentType
    classification
    updatedAt
    owner {
      id
      fullName
    }
    versions(first: 1) @skip(if: $useRequestedVersions) {
      edges {
        node {
          id
          status
          version
          signatures(first: 1000) @include(if: $includeSignatures) {
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
    requestedVersions(first: 1) @include(if: $useRequestedVersions) {
      edges {
        node {
          id
          status
          version
          signatures(first: 1000) @include(if: $includeSignatures) {
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

function DocumentRow({
  document: documentKey,
  organizationId,
  checked,
  onCheck,
  hasAnyAction,
  canViewSignatures,
}: {
  document: DocumentsPageRowFragment$key;
  organizationId: string;
  connectionId: string;
  checked: boolean;
  onCheck: () => void;
  hasAnyAction: boolean;
  canViewSignatures: boolean;
}) {
  const document = useFragment<DocumentsPageRowFragment$key>(
    rowFragment,
    documentKey
  );
  const lastVersion = document.versions?.edges?.[0]?.node || document.requestedVersions?.edges?.[0]?.node;

  if (!lastVersion) {
    return null;
  }

  const isDraft = lastVersion.status === "DRAFT";
  const { __ } = useTranslate();

  let signatures: any[] = [];
  let signedCount = 0;

  if (canViewSignatures && lastVersion.signatures) {
    signatures = lastVersion.signatures.edges?.map((edge) => edge?.node)?.filter(Boolean) ?? [];
    signedCount = signatures.filter((signature) => signature.state === "SIGNED").length;
  }
  const [deleteDocument] = useDeleteDocumentMutation();
  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        deleteDocument({
          variables: {
            input: { documentId: document.id },
          },
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete the document "%s". This action cannot be undone.'
          ),
          document.title
        ),
      }
    );
  };

  return (
    <Tr to={`/organizations/${organizationId}/documents/${document.id}`}>
      <Td noLink className="w-18">
        <Checkbox checked={checked} onChange={onCheck} />
      </Td>
      <Td className="min-w-0">
        <div className="flex gap-4 items-center">{document.title}</div>
      </Td>
      <Td className="w-24">
        <Badge variant={isDraft ? "neutral" : "success"}>
          {isDraft ? __("Draft") : __("Published")}
        </Badge>
      </Td>
      <Td className="w-20">v{lastVersion.version}</Td>
      <Td className="w-28">
        {getDocumentTypeLabel(__, document.documentType)}
      </Td>
      <Td className="w-32">
        {getDocumentClassificationLabel(__, document.classification)}
      </Td>
      <Td className="w-60">
        <div className="flex gap-2 items-center">
          <Avatar name={document.owner?.fullName ?? ""} />
          {document.owner?.fullName}
        </div>
      </Td>
      <Td className="w-60">{formatDate(document.updatedAt)}</Td>
      {canViewSignatures && (
        <Td className="w-20">
          {signedCount}/{signatures.length}
        </Td>
      )}
      {hasAnyAction && (
        <Td noLink width={50} className="text-end w-18">
          <ActionDropdown>
            <Authorized entity="Document" action="deleteDocument">
              <DropdownItem
                variant="danger"
                icon={IconTrashCan}
                onClick={handleDelete}
              >
                {__("Delete")}
              </DropdownItem>
            </Authorized>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}

export default function DocumentsPage() {
  const organizationId = useOrganizationId();

  const [cacheKey, setCacheKey] = useState(0);

  useEffect(() => {
    fetchPermissions(organizationId).then(() => {
      setCacheKey(prev => prev + 1);
    });
  }, [organizationId]);

  const hasCheckedPermissions = cacheKey > 0;

  if (!hasCheckedPermissions) {
    return <Spinner />;
  }

  return <DocumentsPageContent key={cacheKey} />;
}
