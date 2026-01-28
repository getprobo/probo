import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, IconCheckmark1, IconCrossLargeX, IconPencil, IconTrashCan, Td, Tr } from "@probo/ui";
import { useCallback, useState } from "react";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { CompliancePageAccessListItem_deleteMutation } from "#/__generated__/core/CompliancePageAccessListItem_deleteMutation.graphql";
import type { CompliancePageAccessListItemFragment$key } from "#/__generated__/core/CompliancePageAccessListItemFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { TrustCenterAccessEditDialog } from "#/pages/organizations/trustCenter/TrustCenterAccessTab/TrustCenterAccessEditDialog";

const fragment = graphql`
  fragment CompliancePageAccessListItemFragment on TrustCenterAccess {
    id
    name
    email
    createdAt
    active
    activeCount
    pendingRequestCount
    hasAcceptedNonDisclosureAgreement
    canUpdate: permission(action: "core:trust-center-access:update")
    canDelete: permission(action: "core:trust-center-access:delete")
  }
`;

const deleteCompliancePageAccessMutation = graphql`
  mutation CompliancePageAccessListItem_deleteMutation(
    $input: DeleteTrustCenterAccessInput!
    $connections: [ID!]!
  ) {
    deleteTrustCenterAccess(input: $input) {
      deletedTrustCenterAccessId @deleteEdge(connections: $connections)
    }
  }
`;

export function CompliancePageAccessListItem(props: {
  connectionId: DataID;
  fragmentRef: CompliancePageAccessListItemFragment$key;
  dialogOpen: boolean;
}) {
  const { connectionId, fragmentRef, dialogOpen: initialDialogOpen } = props;

  const { __ } = useTranslate();
  const [dialogOpen, setDialogOpen] = useState<boolean>(initialDialogOpen);

  const access = useFragment<CompliancePageAccessListItemFragment$key>(fragment, fragmentRef);

  const [deleteInvitation, isDeleting] = useMutationWithToasts<CompliancePageAccessListItem_deleteMutation>(
    deleteCompliancePageAccessMutation,
    {
      successMessage: __("Access deleted successfully"),
      errorMessage: __("Failed to delete access"),
    },
  );

  const handleDelete = useCallback(
    async (id: string) => {
      await deleteInvitation({
        variables: {
          input: { id },
          connections: [connectionId],
        },
      });
    },
    [deleteInvitation, connectionId],
  );

  return (
    <>
      <Tr
        key={access.id}
        onClick={() => access.canUpdate && setDialogOpen(true)}
        className="cursor-pointer hover:bg-bg-secondary transition-colors"
      >
        <Td className="font-medium">{access.name}</Td>
        <Td>{access.email}</Td>
        <Td>{formatDate(access.createdAt)}</Td>
        <Td>
          <div className="flex justify-center">
            {access.active
              ? (
                  <IconCheckmark1 size={16} className="text-txt-success" />
                )
              : (
                  <IconCrossLargeX size={16} className="text-txt-danger" />
                )}
          </div>
        </Td>
        <Td className="text-center">{access.activeCount}</Td>
        <Td className="text-center">
          {access.pendingRequestCount > 0 ? access.pendingRequestCount : ""}
        </Td>
        <Td>
          <div className="flex justify-center">
            {access.hasAcceptedNonDisclosureAgreement && (
              <IconCheckmark1 size={16} className="text-txt-success" />
            )}
          </div>
        </Td>
        <Td noLink width={160} className="text-end">
          <div
            className="flex gap-2 justify-end"
            onClick={e => e.stopPropagation()}
          >
            {access.canUpdate && (
              <Button
                variant="secondary"
                onClick={() => setDialogOpen(true)}
                icon={IconPencil}
              />
            )}
            {access.canDelete && (
              <Button
                variant="danger"
                onClick={() => void handleDelete(access.id)}
                disabled={isDeleting}
                icon={IconTrashCan}
              />
            )}
          </div>
        </Td>
      </Tr>

      {access.canUpdate && dialogOpen && (
        <TrustCenterAccessEditDialog
          access={access}
          onClose={() => setDialogOpen(false)}
        />
      )}
    </>
  );
}
