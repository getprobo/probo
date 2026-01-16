import {
  Button,
  IconCheckmark1,
  IconCrossLargeX,
  IconPencil,
  IconTrashCan,
  Td,
  Tr,
} from "@probo/ui";
import { formatDate } from "@probo/helpers";
import { useCallback, useState } from "react";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { deleteTrustCenterAccessMutation } from "/hooks/graph/TrustCenterAccessGraph";
import { useTranslate } from "@probo/i18n";
import { TrustCenterAccessEditDialog } from "./TrustCenterAccessEditDialog";
import type { TrustCenterAccessGraph_accesses$data } from "/__generated__/core/TrustCenterAccessGraph_accesses.graphql";
import type { NodeOf } from "/types";

interface TrustCenterAccessItemProps {
  access: NodeOf<TrustCenterAccessGraph_accesses$data["accesses"]>;
  connectionId?: string;
  dialogOpen: boolean;
}

export function TrustCenterAccessItem(props: TrustCenterAccessItemProps) {
  const { access, connectionId, dialogOpen: initialDialogOpen } = props;

  const { __ } = useTranslate();
  const [dialogOpen, setDialogOpen] = useState<boolean>(initialDialogOpen);

  const [deleteInvitation, isDeleting] = useMutationWithToasts(
    deleteTrustCenterAccessMutation,
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
          connections: connectionId ? [connectionId] : [],
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
            {access.active ? (
              <IconCheckmark1 size={16} className="text-txt-success" />
            ) : (
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
            onClick={(e) => e.stopPropagation()}
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
                onClick={() => handleDelete(access.id)}
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
