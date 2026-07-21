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
import { dateFormat } from "@probo/i18n";
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
import { useTranslation } from "react-i18next";
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
          signatures(first: 0 filter: { activeContract: true, profileState: ACTIVE }) {
            totalCount
          }
          signedSignatures: signatures(first: 0 filter: { states: [SIGNED], activeContract: true, profileState: ACTIVE }) {
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
  const { t, i18n } = useTranslation();
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
    DRAFT: t("documentListItem.status.draft"),
    PENDING_APPROVAL: t("documentListItem.status.pendingApproval"),
    PUBLISHED: t("documentListItem.status.published"),
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
                  title: t("documentListItem.errors.title"),
                  description: formatError(t("documentListItem.errors.archive"), errors),
                  variant: "error",
                });
              } else {
                toast({
                  title: t("documentListItem.messages.successTitle"),
                  description: t("documentListItem.messages.archived"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({ title: t("documentListItem.errors.title"), description: error.message, variant: "error" });
              resolve();
            },
          });
        }),
      {
        message: t("documentListItem.confirmations.archive", {
          title: lastVersion.title,
        }),
        variant: "danger",
        label: t("documentListItem.actions.archive"),
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
            title: t("documentListItem.errors.title"),
            description: formatError(t("documentListItem.errors.unarchive"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("documentListItem.messages.successTitle"),
          description: t("documentListItem.messages.unarchived"),
          variant: "success",
        });
      },
      onError(error) {
        toast({ title: t("documentListItem.errors.title"), description: error.message, variant: "error" });
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
          {t(`documentListItem.documentTypes.${lastVersion.documentType.toLowerCase()}`)}
        </Td>
        <Td className="w-32">
          {t(`documentListItem.classifications.${lastVersion.classification.toLowerCase()}`)}
        </Td>
        <Td className="w-60">
          {(() => {
            if (lastVersion.status === "PENDING_APPROVAL") {
              const quorum = lastVersion.approvalQuorums?.edges?.[0]?.node;
              if (quorum) {
                if (quorum.status === "REJECTED") {
                  return t("documentListItem.status.rejected");
                }
                return `${quorum.approvedDecisions.totalCount}/${quorum.decisions.totalCount}`;
              }
              return "—";
            }
            if (!document.defaultApprovers.length) return "—";
            return document.defaultApprovers.map(a => a.fullName).join(", ");
          })()}
        </Td>
        <Td className="w-40">{dateFormat(i18n.language, document.updatedAt)}</Td>
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
                    {t("documentListItem.actions.archive")}
                  </DropdownItem>
                )}
                {document.canUnarchive && document.status === "ARCHIVED" && (
                  <DropdownItem
                    icon={IconArchive}
                    disabled={isUnarchiving}
                    onClick={handleUnarchive}
                  >
                    {t("documentListItem.actions.unarchive")}
                  </DropdownItem>
                )}
                {document.canDelete && (
                  <DropdownItem
                    variant="danger"
                    icon={IconTrashCan}
                    onClick={handleDelete}
                  >
                    {t("documentListItem.actions.delete")}
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
