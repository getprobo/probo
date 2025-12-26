import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  IconPencil,
  IconTrashCan,
  Spinner,
  Td,
  Tr,
  useConfirm,
} from "@probo/ui";
import clsx from "clsx";
import { useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import type { MemberListItemFragment$key } from "./__generated__/MemberListItemFragment.graphql";
import type { MemberListItem_permissionsFragment$key } from "./__generated__/MemberListItem_permissionsFragment.graphql";
import type { MemberListItem_currentRoleFragment$key } from "./__generated__/MemberListItem_currentRoleFragment.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { sprintf } from "@probo/helpers";
import { EditMemberDialog } from "./EditMemberDialog";

const fragment = graphql`
  fragment MemberListItemFragment on Membership {
    id
    role
    profile @required(action: THROW) {
      fullName
    }
    identity @required(action: THROW) {
      email
    }
    createdAt
  }
`;

const currentRoleFragment = graphql`
  fragment MemberListItem_currentRoleFragment on Organization {
    viewerMembership @required(action: THROW) {
      role
    }
  }
`;

const permissionsFragment = graphql`
  fragment MemberListItem_permissionsFragment on Identity
  @argumentDefinitions(organizationId: { type: "ID!" }) {
    canUpdateMembership: permission(
      action: "iam:membership:update"
      id: $organizationId
    )
    canDeleteMembership: permission(
      action: "iam:membership:delete"
      id: $organizationId
    )
  }
`;

const removeMemberMutation = graphql`
  mutation MemberListItem_removeMutation(
    $input: RemoveMemberInput!
    $connections: [ID!]!
  ) {
    removeMember(input: $input) {
      deletedMembershipId @deleteEdge(connections: $connections)
    }
  }
`;

export function MemberListItem(props: {
  connectionId: string;
  fKey: MemberListItemFragment$key;
  permissionsFKey: MemberListItem_permissionsFragment$key;
  viewerFKey: MemberListItem_currentRoleFragment$key;
  onRefetch: () => void;
}) {
  const { fKey, connectionId, permissionsFKey, viewerFKey } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const [dialogOpen, setDialogOpen] = useState(false);

  const membership = useFragment<MemberListItemFragment$key>(fragment, fKey);
  const { viewerMembership } =
    useFragment<MemberListItem_currentRoleFragment$key>(
      currentRoleFragment,
      viewerFKey,
    );
  const permissions = useFragment<MemberListItem_permissionsFragment$key>(
    permissionsFragment,
    permissionsFKey,
  );

  // Only OWNER can edit OWNER members
  const canEditThisRole =
    membership.role === "OWNER" ? viewerMembership.role === "OWNER" : true;

  const [removeMembership, isRemoving] = useMutationWithToasts(
    removeMemberMutation,
    {
      successMessage: __("Member removed successfully"),
      errorMessage: __("Failed to remove member"),
    },
  );

  const handleRemove = async () => {
    confirm(
      () => {
        return removeMembership({
          variables: {
            input: {
              membershipId: membership.id,
              organizationId: organizationId,
            },
            connections: [connectionId],
          },
        });
      },
      {
        message: sprintf(
          __("Are you sure you want to remove %s?"),
          membership.profile.fullName,
        ),
      },
    );
  };

  return (
    <>
      <Tr className={clsx(isRemoving && "opacity-60 pointer-events-none")}>
        <Td>
          <div className="font-semibold">{membership.profile.fullName}</div>
        </Td>
        <Td>
          <div className="flex items-center gap-2">
            {membership.identity.email}
            {/* FIXME: put back */}
            {/* {membership.authMethod === "SAML" && (
              <Badge variant="info">SAML</Badge>
            )} */}
          </div>
        </Td>
        <Td>
          <Badge>{membership.role}</Badge>
        </Td>
        <Td>{new Date(membership.createdAt).toLocaleDateString()}</Td>
        <Td noLink width={160} className="text-end">
          <div
            className="flex gap-2 justify-end"
            onClick={(e) => e.stopPropagation()}
          >
            {permissions.canUpdateMembership && canEditThisRole && (
              <Button
                variant="secondary"
                onClick={() => setDialogOpen(true)}
                disabled={dialogOpen}
                icon={IconPencil}
                aria-label={__("Edit role")}
              />
            )}
            {isRemoving ? (
              <Spinner size={16} />
            ) : (
              permissions.canDeleteMembership &&
              canEditThisRole && (
                <Button
                  variant="danger"
                  onClick={handleRemove}
                  disabled={isRemoving}
                  icon={IconTrashCan}
                  aria-label={__("Remove member")}
                />
              )
            )}
          </div>
        </Td>
      </Tr>

      {dialogOpen && (
        <EditMemberDialog
          currentRole={viewerMembership.role}
          onClose={() => setDialogOpen(false)}
          membership={membership}
        />
      )}
    </>
  );
}
