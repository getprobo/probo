import { getAssignableRoles, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  IconTrashCan,
  Option,
  Select,
  Spinner,
  Td,
  Tr,
  useConfirm,
} from "@probo/ui";
import { clsx } from "clsx";
import { use } from "react";
import { useFragment } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

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
    kind
    position
    membership @required(action: THROW) {
      id
      role
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

  const isInactive = profile.state === "INACTIVE";

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
            <Badge variant="info">{profile.source}</Badge>
          </div>
        </Td>
        <Td>{profile.kind}</Td>
        <Td noLink className="pr-4">
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
        <Td>{new Date(profile.createdAt).toLocaleDateString()}</Td>
        <Td>{profile.position}</Td>
        <Td noLink width={160} className="text-end">
          {!isInactive && (
            <div
              className="flex gap-2 justify-end"
              onClick={e => e.stopPropagation()}
            >
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
    </>
  );
}
