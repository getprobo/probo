import { formatDate, getDocumentClassificationLabel, getDocumentTypeLabel, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { ActionDropdown, Badge, Checkbox, DropdownItem, IconTrashCan, Td, Tr, useConfirm } from "@probo/ui";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { DocumentListItem_deleteMutation } from "#/__generated__/core/DocumentListItem_deleteMutation.graphql";
import type { DocumentListItemFragment$key } from "#/__generated__/core/DocumentListItemFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
  fragment DocumentListItemFragment on Document {
    id
    title
    documentType
    classification
    updatedAt
    canDelete: permission(action: "core:document:delete")
    recentVersions: versions(first: 2 orderBy: { field: CREATED_AT direction: DESC }) {
      edges {
        node {
          id
          status
          version
          approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
            edges {
              node {
                status
                decisions(first: 20) {
                  totalCount
                  edges {
                    node {
                      approver {
                        fullName
                      }
                    }
                  }
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
  const document = useFragment<DocumentListItemFragment$key>(
    fragment,
    fragmentRef,
  );
  const lastVersion = document.recentVersions.edges[0].node;
  const approverQuorum = lastVersion.approvalQuorums?.edges?.[0]?.node
    ?? document.recentVersions.edges[1]?.node.approvalQuorums?.edges?.[0]?.node;

  const { __ } = useTranslate();

  const statusVariant = {
    DRAFT: "neutral",
    PUBLISHED: "success",
  } as const;

  const statusLabel = {
    DRAFT: __("Draft"),
    PUBLISHED: __("Published"),
  } as const;

  const [deleteDocument] = useMutationWithToasts<DocumentListItem_deleteMutation>(
    deleteDocumentMutation,
    {
      successMessage: __("Document deleted successfully."),
      errorMessage: __("Failed to delete document"),
    },
  );
  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        deleteDocument({
          variables: {
            connections: [connectionId],
            input: { documentId: document.id },
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
        {lastVersion.version}
      </Td>
      <Td className="w-28">
        {getDocumentTypeLabel(__, document.documentType)}
      </Td>
      <Td className="w-32">
        {getDocumentClassificationLabel(__, document.classification)}
      </Td>
      <Td className="w-60">
        {(() => {
          const decisions = approverQuorum?.decisions;
          if (!decisions?.edges.length) return "—";
          const names = decisions.edges.map(e => e.node.approver.fullName).join(", ");
          return decisions.totalCount > 20 ? `${names}...` : names;
        })()}
      </Td>
      <Td className="w-60">{formatDate(document.updatedAt)}</Td>
      <Td className="w-20">
        {(() => {
          const lastQuorum = lastVersion.approvalQuorums?.edges?.[0]?.node;
          return lastQuorum
            ? lastQuorum.status === "REJECTED"
              ? __("Rejected")
              : `${lastQuorum.approvedDecisions.totalCount}/${lastQuorum.decisions.totalCount}`
            : "—";
        })()}
      </Td>
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
