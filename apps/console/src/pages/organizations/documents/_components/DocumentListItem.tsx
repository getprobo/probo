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

import { formatDate, getDocumentClassificationLabel, getDocumentTypeLabel, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { ActionDropdown, Badge, Checkbox, DropdownItem, IconTrashCan, Td, Tr, useConfirm } from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { DocumentListItem_deleteMutation } from "#/__generated__/core/DocumentListItem_deleteMutation.graphql";
import type { DocumentListItemFragment$key } from "#/__generated__/core/DocumentListItemFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
  fragment DocumentListItemFragment on Document {
    id
    title
    updatedAt
    canDelete: permission(action: "core:document:delete")
    defaultApprovers {
      id
      fullName
    }
    recentVersions: versions(first: 2 orderBy: { field: CREATED_AT direction: DESC }) {
      edges {
        node {
          id
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
          signatures(first: 0 filter: { activeContract: true }) {
            totalCount
          }
          signedSignatures: signatures(first: 0 filter: { states: [SIGNED], activeContract: true }) {
            totalCount
          }
        }
      }
    }
  }
`;

const deleteDocumentMutation = graphql`
  mutation DocumentListItem_deleteMutation(
    $input: DeleteDocumentInput!
    $connections: [ID!]!
  ) {
    deleteDocument(input: $input) {
      deletedDocumentId @deleteEdge(connections: $connections)
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
  const [deleteDocument] = useMutation<DocumentListItem_deleteMutation>(deleteDocumentMutation);
  const confirm = useConfirm();
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

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve, reject) => {
          deleteDocument({
            variables: {
              connections: [connectionId],
              input: { documentId: document.id },
            },
            onCompleted: () => resolve(),
            onError: err => reject(err),
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

  return (
    <Tr
      to={`/organizations/${organizationId}/documents/${document.id}`}
    >
      <Td noLink className="w-18">
        <Checkbox checked={checked} onChange={onCheck} />
      </Td>
      <Td className="min-w-0">
        <div className="flex gap-4 items-center">{document.title}</div>
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
          <ActionDropdown>
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
        </Td>
      )}
    </Tr>
  );
}
