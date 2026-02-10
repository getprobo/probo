import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { ActionDropdown, DropdownItem, IconArchive, IconCheckmark1, IconPencil, IconRotateCw, Td, Tr } from "@probo/ui";
import { useCallback, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageAccessListItemFragment$key } from "#/__generated__/core/CompliancePageAccessListItemFragment.graphql";
import type { TrustCenterAccessGraphUpdateMutation } from "#/__generated__/core/TrustCenterAccessGraphUpdateMutation.graphql";
import { updateTrustCenterAccessMutation } from "#/hooks/graph/TrustCenterAccessGraph";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { TrustCenterAccessEditDialog } from "#/pages/organizations/trustCenter/TrustCenterAccessTab/TrustCenterAccessEditDialog";

const fragment = graphql`
  fragment CompliancePageAccessListItemFragment on TrustCenterAccess {
    id
    name
    email
    createdAt
    state
    activeCount
    pendingRequestCount
    hasAcceptedNonDisclosureAgreement
    canUpdate: permission(action: "core:trust-center-access:update")
  }
`;

export function CompliancePageAccessListItem(props: {
  fragmentRef: CompliancePageAccessListItemFragment$key;
  dialogOpen: boolean;
}) {
  const { fragmentRef, dialogOpen: initialDialogOpen } = props;

  const { __ } = useTranslate();
  const [dialogOpen, setDialogOpen] = useState<boolean>(initialDialogOpen);

  const access = useFragment<CompliancePageAccessListItemFragment$key>(fragment, fragmentRef);

  const isActive = access.state === "ACTIVE";

  const [updateAccess, isUpdating] = useMutationWithToasts<TrustCenterAccessGraphUpdateMutation>(
    updateTrustCenterAccessMutation,
    {
      successMessage: isActive
        ? __("Access deactivated successfully")
        : __("Access activated successfully"),
      errorMessage: isActive
        ? __("Failed to deactivate access")
        : __("Failed to activate access"),
    },
  );

  const handleToggleState = useCallback(() => {
    void updateAccess({
      variables: {
        input: {
          id: access.id,
          name: access.name,
          state: isActive ? "INACTIVE" : "ACTIVE",
        },
      },
    });
  }, [updateAccess, access.id, access.name, isActive]);

  return (
    <>
      <Tr
        key={access.id}
        onClick={() => access.canUpdate && isActive && setDialogOpen(true)}
        className={`cursor-pointer hover:bg-bg-secondary transition-colors${!isActive ? " opacity-50" : ""}`}
      >
        <Td className="font-medium">{access.name}</Td>
        <Td>{access.email}</Td>
        <Td>{formatDate(access.createdAt)}</Td>
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
              <ActionDropdown>
                {isActive && (
                  <DropdownItem
                    icon={IconPencil}
                    onClick={() => setDialogOpen(true)}
                  >
                    {__("Edit")}
                  </DropdownItem>
                )}
                <DropdownItem
                  icon={isActive ? IconArchive : IconRotateCw}
                  onClick={handleToggleState}
                  disabled={isUpdating}
                  variant={isActive ? "danger" : "primary"}
                >
                  {isActive ? __("Deactivate") : __("Activate")}
                </DropdownItem>
              </ActionDropdown>
            )}
          </div>
        </Td>
      </Tr>

      {access.canUpdate && isActive && dialogOpen && (
        <TrustCenterAccessEditDialog
          access={access}
          onClose={() => setDialogOpen(false)}
        />
      )}
    </>
  );
}
