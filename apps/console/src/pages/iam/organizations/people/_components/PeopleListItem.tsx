import { getAssignableRoles, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  DropdownItem,
  IconMail,
  IconTrashCan,
  Option,
  Select,
  Td,
  Tr,
  useConfirm,
} from "@probo/ui";
import { clsx } from "clsx";
import { use } from "react";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { PeopleListItem_inviteMutation } from "#/__generated__/iam/PeopleListItem_inviteMutation.graphql";
import type { PeopleListItemFragment$key } from "#/__generated__/iam/PeopleListItemFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

const fragment = graphql`
  fragment PeopleListItemFragment on Profile {
    id
    source
    state
    fullName
    membership @required(action: THROW) {
      id
      role
      canUpdate: permission(action: "iam:membership:update")
    }
    identity @required(action: THROW) {
      email
    }
    lastInvitation: pendingInvitations(first: 1, orderBy: { field: CREATED_AT, direction: DESC })
    @required(action: THROW)
    @connection(key: "PeopleListItem_lastInvitation") {
      __id
      edges {
        node {
          id
          expiresAt
          acceptedAt
          createdAt
        }
      }
    }
    createdAt
    canUpdate: permission(action: "iam:membership-profile:update")
    canInvite: permission(action: "iam:invitation:create")
    canDelete: permission(action: "iam:membership-profile:delete")
  }
`;

const inviteUserMutation = graphql`
  mutation PeopleListItem_inviteMutation(
    $input: InviteUserInput!
    $connections: [ID!]!
  ) {
    inviteUser(input: $input) {
      invitationEdge @prependEdge(connections: $connections) {
        node {
          id
          expiresAt
          acceptedAt
          createdAt
        }
      }
    }
  }
`;

const updateRoleMutation = graphql`
  mutation PeopleListItem_updateRoleMutation($input: UpdateMembershipInput!) {
    updateMembership(input: $input) {
      membership {
        id
        role
      }
    }
  }
`;

const removeUserMutation = graphql`
  mutation PeopleListItem_removeMutation(
    $input: RemoveUserInput!
    $connections: [ID!]!
  ) {
    removeUser(input: $input) {
      deletedProfileId @deleteEdge(connections: $connections)
    }
  }
`;

export function PeopleListItem(props: {
  connectionId: DataID;
  fKey: PeopleListItemFragment$key;
  onRefetch: () => void;
}) {
  const { fKey, connectionId } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();

  const { role } = use(CurrentUser);
  const availableRoles = getAssignableRoles(role);

  const profile = useFragment<PeopleListItemFragment$key>(fragment, fKey);
  const lastInvitation = profile.lastInvitation.edges[0]?.node;

  const isInactive = profile.state === "INACTIVE";

  const [inviteUser]
    = useMutationWithToasts<PeopleListItem_inviteMutation>(inviteUserMutation, {
      successMessage: __("Invitation sent successfully"),
      errorMessage: __("Failed to send invitation"),
    });
  const [updateMembership, isUpdatingRole] = useMutationWithToasts(
    updateRoleMutation,
    {
      successMessage: __("Role updated successfully"),
      errorMessage: __("Failed to update role"),
    },
  );
  const [removeUser, isRemoving] = useMutationWithToasts(
    removeUserMutation,
    {
      successMessage: __("Person removed successfully"),
      errorMessage: __("Failed to remove person"),
    },
  );

  const handleInvite = () => {
    confirm(
      () => {
        return inviteUser({
          variables: {
            input: {
              organizationId,
              profileId: profile.id,
            },
            connections: [profile.lastInvitation.__id],
          },
        });
      },
      {
        label: __("Send"),
        variant: "primary",
        message: sprintf(
          __("Send the activation email to %s?"),
          profile.fullName,
        ),
      },
    );
  };
  const handleUpdateRole = async (role: string) => {
    await updateMembership({
      variables: {
        input: {
          membershipId: profile.membership.id,
          organizationId: organizationId,
          role: role,
        },
      },
    });
  };
  const handleRemove = () => {
    confirm(
      () => {
        return removeUser({
          variables: {
            input: {
              profileId: profile.id,
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
      <Tr to={`/organizations/${organizationId}/people/${profile.id}`}>
        <Td className={clsx(
          isRemoving && "opacity-60 pointer-events-none",
          isInactive && "opacity-50",
        )}
        >
          <span className="font-semibold">{profile.fullName}</span>
        </Td>
        <Td>
          <Badge variant={profile.state === "INACTIVE" ? "neutral" : "success"}>{profile.state}</Badge>
        </Td>
        <Td className={clsx(
          isRemoving && "opacity-60 pointer-events-none",
          isInactive && "opacity-50",
        )}
        >
          <div className="flex items-center gap-2">
            {profile.identity.email}
            <Badge variant="info">{profile.source}</Badge>
          </div>
        </Td>
        <Td
          noLink
          className={clsx(
            "pr-4",
            isRemoving && "opacity-60 pointer-events-none",
            isInactive && "opacity-50",
          )}
        >
          <Select
            disabled={!profile.membership.canUpdate || isUpdatingRole}
            value={profile.membership.role}
            onValueChange={role => void handleUpdateRole(role)}
          >
            {availableRoles.includes("OWNER") && (
              <Option value="OWNER">{__("Owner")}</Option>
            )}
            {availableRoles.includes("ADMIN") && (
              <Option value="ADMIN">{__("Admin")}</Option>
            )}
            {availableRoles.includes("VIEWER") && (
              <Option value="VIEWER">{__("Viewer")}</Option>
            )}
            {availableRoles.includes("AUDITOR") && (
              <Option value="AUDITOR">{__("Auditor")}</Option>
            )}
            {availableRoles.includes("EMPLOYEE") && (
              <Option value="EMPLOYEE">{__("Employee")}</Option>
            )}
          </Select>
        </Td>
        <Td className={clsx(
          isRemoving && "opacity-60 pointer-events-none",
          isInactive && "opacity-50",
        )}
        >
          {new Date(profile.createdAt).toLocaleDateString()}
        </Td>
        <Td noLink width={160} className="text-end">
          <ActionDropdown>
            {isInactive && (
              <DropdownItem
                onClick={handleInvite}
                icon={IconMail}
              >
                {lastInvitation ? __("Resend activation mail") : __("Send activation mail")}
              </DropdownItem>
            )}
            {profile.canDelete && (
              <DropdownItem
                onClick={handleRemove}
                variant="danger"
                icon={IconTrashCan}
              >
                {__("Remove person")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      </Tr>
    </>
  );
}
