import { sprintf } from "@probo/helpers";
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
import { clsx } from "clsx";
import { useState } from "react";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { PeopleListItemFragment$key } from "#/__generated__/iam/PeopleListItemFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { EditMemberDialog } from "../../settings/_components/EditMemberDialog";

const fragment = graphql`
  fragment PeopleListItemFragment on Profile {
    id
    fullName
    kind
    position
    membership @required(action: THROW) {
      id
      role
      source
      state
      canUpdate: permission(action: "iam:membership:update")
      canDelete: permission(action: "iam:membership-profile:delete")
    }
    identity @required(action: THROW) {
      email
    }
    createdAt
    canUpdate: permission(action: "iam:membership-profile:update")
  }
`;

const removeMemberMutation = graphql`
  mutation PeopleListItem_removeMutation(
    $input: RemoveMemberInput!
    $connections: [ID!]!
  ) {
    removeMember(input: $input) {
      deletedMembershipId @deleteEdge(connections: $connections)
    }
  }
`;

export function MemberListItem(props: {
  connectionId: DataID;
  fKey: PeopleListItemFragment$key;
  onRefetch: () => void;
}) {
  const { fKey, connectionId } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const [dialogOpen, setDialogOpen] = useState(false);

  const profile = useFragment<PeopleListItemFragment$key>(fragment, fKey);

  const isInactive = profile.membership.state === "INACTIVE";

  const [removeMembership, isRemoving] = useMutationWithToasts(
    removeMemberMutation,
    {
      successMessage: __("Member removed successfully"),
      errorMessage: __("Failed to remove member"),
    },
  );

  const handleRemove = () => {
    confirm(
      () => {
        return removeMembership({
          variables: {
            input: {
              membershipId: profile.membership.id,
              organizationId: organizationId,
            },
            connections: [connectionId],
          },
        });
      },
      {
        message: sprintf(
          __("Are you sure you want to remove %s?"),
          profile.fullName,
        ),
      },
    );
  };

  return (
    <>
      <Tr
        className={clsx(
          isRemoving && "opacity-60 pointer-events-none",
          isInactive && "opacity-50",
        )}
        to={`/organizations/${organizationId}/people/${profile.id}`}
      >
        <Td>
          <div className="flex items-center gap-2">
            <span className="font-semibold">{profile.fullName}</span>
            {isInactive && <Badge variant="neutral">{__("Inactive")}</Badge>}
          </div>
        </Td>
        <Td>
          <div className="flex items-center gap-2">
            {profile.identity.email}
            <Badge variant="info">{profile.membership.source}</Badge>
          </div>
        </Td>
        <Td>{profile.kind}</Td>
        <Td>
          <Badge>{profile.membership.role}</Badge>
        </Td>
        <Td>{new Date(profile.createdAt).toLocaleDateString()}</Td>
        <Td>{profile.position}</Td>
        <Td noLink width={160} className="text-end">
          {!isInactive && (
            <div
              className="flex gap-2 justify-end"
              onClick={e => e.stopPropagation()}
            >
              {profile.membership.canUpdate && (
                <Button
                  variant="secondary"
                  onClick={() => setDialogOpen(true)}
                  disabled={dialogOpen}
                  icon={IconPencil}
                  aria-label={__("Edit role")}
                />
              )}
              {isRemoving
                ? (
                    <Spinner size={16} />
                  )
                : (
                    profile.membership.canDelete && (
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
          profile={profile}
        />
      )}
    </>
  );
}
