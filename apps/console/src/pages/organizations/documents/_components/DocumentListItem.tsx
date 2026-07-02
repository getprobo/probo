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

import {
  formatDate,
  formatError,
  getDocumentClassificationLabel,
  getDocumentTypeLabel,
  sprintf,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Checkbox,
  DropdownItem,
  IconArchive,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useRef } from "react";
import { useFragment, useMutation } from "react-relay";
import { ConnectionHandler, type DataID, graphql } from "relay-runtime";

import type { DocumentListItem_archiveMutation } from "#/__generated__/core/DocumentListItem_archiveMutation.graphql";
import type { DocumentListItem_unarchiveMutation } from "#/__generated__/core/DocumentListItem_unarchiveMutation.graphql";
import type { DocumentListItemFragment$key } from "#/__generated__/core/DocumentListItemFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { DeleteDocumentDialog, type DeleteDocumentDialogRef } from "./DeleteDocumentDialog";

const fragment = graphql`
  fragment DocumentListItemFragment on Document {
    id
    status
    updatedAt
    canArchive: permission(action: "core:document:archive")
    canDelete: permission(action: "core:document:delete")
    canUnarchive: permission(action: "core:document:unarchive")
    defaultApprovers {
      id
      fullName
    }
    recentVersions: versions(first: 2 orderBy: { field: CREATED_AT direction: DESC }) {
      edges {
        node {
          id
          title
          status
          major
          minor
          documentType
          classification
          approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
            edges {
              node {
                status
                decisions(first: 0) {
                  totalCount
                }
                approvedDecisions: decisions(first: 0 filter: { states: [APPROVED] }) {
                  totalCount
                }
              }
            }
          }
          signatures(first: 0 filter: { activeContract: true, state: ACTIVE }) {
            totalCount
          }
          signedSignatures: signatures(first: 0 filter: { states: [SIGNED], activeContract: true, state: ACTIVE }) {
            totalCount
          }
        }
      }
    }
  }
`;

const archiveDocumentMutation = graphql`
  mutation DocumentListItem_archiveMutation(
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
  mutation DocumentListItem_unarchiveMutation(
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

export function DocumentListItem(props: {
  fragmentRef: DocumentListItemFragment$key;
  connectionId: DataID;
  checked: boolean;
  onCheck: () => void;
  hasAnyAction: boolean;
}) {
  const {
    connectionId,
    fragmentRef,
    checked,
    onCheck,
    hasAnyAction,
  } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [archiveDocument, isArchiving] = useMutation<DocumentListItem_archiveMutation>(archiveDocumentMutation);
  const [unarchiveDocument, isUnarchiving] = useMutation<DocumentListItem_unarchiveMutation>(unarchiveDocumentMutation);
  const confirm = useConfirm();
  const deleteDialogRef = useRef<DeleteDocumentDialogRef>(null);
  const document = useFragment<DocumentListItemFragment$key>(
    fragment,
    fragmentRef,
  );

  const lastVersionEdge = document.recentVersions.edges[0];
  if (!lastVersionEdge) return null;
  const lastVersion = lastVersionEdge.node;

  const statusVariant = {
    DRAFT: "neutral",
    PENDING_APPROVAL: "warning",
    PUBLISHED: "success",
  } as const;

  const statusLabel = {
    DRAFT: __("Draft"),
    PENDING_APPROVAL: __("Pending approval"),
    PUBLISHED: __("Published"),
  } as const;

  const handleArchive = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          archiveDocument({
            variables: { input: { documentId: document.id } },
            updater(store) {
              const conn = store.get(connectionId);
              if (conn) {
                ConnectionHandler.deleteNode(conn, document.id);
              }
            },
            onCompleted(_, errors) {
              if (errors?.length) {
                toast({
                  title: __("Error"),
                  description: formatError(__("Failed to archive document"), errors),
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Document archived successfully."),
                  variant: "success",
                });
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
          lastVersion.title,
        ),
        variant: "danger",
        label: __("Archive"),
      },
    );
  };

  const handleDelete = () => {
    deleteDialogRef.current?.open();
  };

  const handleUnarchive = () => {
    unarchiveDocument({
      variables: { input: { documentId: document.id } },
      updater(store) {
        const conn = store.get(connectionId);
        if (conn) {
          ConnectionHandler.deleteNode(conn, document.id);
        }
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to unarchive document"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Document unarchived successfully."),
          variant: "success",
        });
      },
      onError(error) {
        toast({ title: __("Error"), description: error.message, variant: "error" });
      },
    });
  };

  const hasRowAction
    = (document.canArchive && document.status === "ACTIVE")
      || (document.canUnarchive && document.status === "ARCHIVED")
      || document.canDelete;

  return (
    <>
      <Tr
        to={`/organizations/${organizationId}/documents/${document.id}`}
      >
        <Td noLink className="w-18">
          <Checkbox checked={checked} onChange={onCheck} />
        </Td>
        <Td className="min-w-0">
          <div className="flex gap-4 items-center">{lastVersion.title}</div>
        </Td>
        <Td className="w-24">
          <Badge variant={statusVariant[lastVersion.status]}>
            {statusLabel[lastVersion.status]}
          </Badge>
        </Td>
        <Td className="w-20">
          v
          {lastVersion.major}
          .
          {lastVersion.minor}
        </Td>
        <Td className="w-28">
          {getDocumentTypeLabel(__, lastVersion.documentType)}
        </Td>
        <Td className="w-32">
          {getDocumentClassificationLabel(__, lastVersion.classification)}
        </Td>
        <Td className="w-60">
          {(() => {
            if (lastVersion.status === "PENDING_APPROVAL") {
              const quorum = lastVersion.approvalQuorums?.edges?.[0]?.node;
              if (quorum) {
                if (quorum.status === "REJECTED") return __("Rejected");
                return `${quorum.approvedDecisions.totalCount}/${quorum.decisions.totalCount}`;
              }
              return "—";
            }
            if (!document.defaultApprovers.length) return "—";
            return document.defaultApprovers.map(a => a.fullName).join(", ");
          })()}
        </Td>
        <Td className="w-40">{formatDate(document.updatedAt)}</Td>
        <Td className="w-20">
          {lastVersion.signedSignatures.totalCount}
          /
          {lastVersion.signatures.totalCount}
        </Td>
        {hasAnyAction && (
          <Td noLink width={50} className="text-end w-18">
            {hasRowAction && (
              <ActionDropdown>
                {document.canArchive && document.status === "ACTIVE" && (
                  <DropdownItem
                    icon={IconArchive}
                    disabled={isArchiving}
                    onClick={handleArchive}
                  >
                    {__("Archive")}
                  </DropdownItem>
                )}
                {document.canUnarchive && document.status === "ARCHIVED" && (
                  <DropdownItem
                    icon={IconArchive}
                    disabled={isUnarchiving}
                    onClick={handleUnarchive}
                  >
                    {__("Unarchive")}
                  </DropdownItem>
                )}
                {document.canDelete && (
                  <DropdownItem
                    variant="danger"
                    icon={IconTrashCan}
                    onClick={handleDelete}
                  >
                    {__("Delete")}
                  </DropdownItem>
                )}
              </ActionDropdown>
            )}
          </Td>
        )}
      </Tr>
      <DeleteDocumentDialog
        ref={deleteDialogRef}
        documentId={document.id}
        documentTitle={lastVersion.title}
        connections={[connectionId]}
      />
    </>
  );
}
