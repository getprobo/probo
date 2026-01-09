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
import { use, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import type { MemberListItemFragment$key } from "/__generated__/iam/MemberListItemFragment.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { sprintf } from "@probo/helpers";
import { EditMemberDialog } from "./EditMemberDialog";
import { CurrentUser } from "/providers/CurrentUser";

const fragment = graphql`
  fragment MemberListItemFragment on Membership {
    id
    role
    source
    state
    profile @required(action: THROW) {
      fullName
    }
    identity @required(action: THROW) {
      email
    }
    createdAt
    canUpdate: permission(action: "iam:membership:update")
    canDelete: permission(action: "iam:membership:delete")
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
  onRefetch: () => void;
}) {
  const { fKey, connectionId } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const [dialogOpen, setDialogOpen] = useState(false);

  const membership = useFragment<MemberListItemFragment$key>(fragment, fKey);
  const { role } = use(CurrentUser);

  const isInactive = membership.state === "INACTIVE";

  // Only OWNER can edit OWNER members
  const canEditThisRole = membership.role === "OWNER" ? role === "OWNER" : true;

  const [removeMembership, isRemoving] = useMutationWithToasts(
    removeMemberMutation,
    {
      successMessage: __("Member removed successfully"),
      errorMessage: __("Failed to remove member"),
    }
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
          membership.profile.fullName
        ),
      }
    );
  };

  return (
    <>
      <Tr
        className={clsx(
          isRemoving && "opacity-60 pointer-events-none",
          isInactive && "opacity-50"
        )}
      >
        <Td>
          <div className="flex items-center gap-2">
            <span className="font-semibold">{membership.profile.fullName}</span>
            {isInactive && <Badge variant="neutral">{__("Inactive")}</Badge>}
          </div>
        </Td>
        <Td>
          <div className="flex items-center gap-2">
            {membership.identity.email}
            <Badge variant="info">{membership.source}</Badge>
          </div>
        </Td>
        <Td>
          <Badge>{membership.role}</Badge>
        </Td>
        <Td>{new Date(membership.createdAt).toLocaleDateString()}</Td>
        <Td noLink width={160} className="text-end">
          {!isInactive && (
            <div
              className="flex gap-2 justify-end"
              onClick={(e) => e.stopPropagation()}
            >
              {membership.canUpdate && canEditThisRole && (
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
                membership.canDelete &&
                canEditThisRole &&
                membership.source !== "SCIM" && (
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
          )}
        </Td>
      </Tr>

      {dialogOpen && (
        <EditMemberDialog
          onClose={() => setDialogOpen(false)}
          membership={membership}
        />
      )}
    </>
  );
}
