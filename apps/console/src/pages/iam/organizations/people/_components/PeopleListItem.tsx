// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { getAssignableRoles, getMembershipRoles } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  DropdownItem,
  IconArchive,
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
import { useTranslation } from "react-i18next";
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
    emailAddress
    membership @required(action: THROW) {
      id
      role
      canUpdate: permission(action: "iam:membership:update", attributes: { target_role: "VIEWER" })
    }
    lastInvitation: pendingInvitations(first: 1, orderBy: { field: CREATED_AT, direction: DESC })
    @required(action: THROW)
    @connection(key: "PeopleListItem_lastInvitation") {
      __id
      edges {
        node {
          id
          createdAt
        }
      }
    }
    createdAt
    contractEndDate
    canUpdate: permission(action: "iam:membership-profile:update")
    canInvite: permission(action: "iam:invitation:create")
    canDeactivate: permission(action: "iam:membership-profile:deactivate")
    canRemoveMember: permission(action: "iam:membership:delete")
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

const deactivateUserMutation = graphql`
  mutation PeopleListItem_deactivateMutation($input: DeactivateUserInput!) {
    deactivateUser(input: $input) {
      success
    }
  }
`;

export function PeopleListItem(props: {
  connectionId: DataID;
  fKey: PeopleListItemFragment$key;
  onRefetch: () => void;
}) {
  const { fKey, connectionId, onRefetch } = props;

  const organizationId = useOrganizationId();
  const { t, i18n } = useTranslation();
  const confirm = useConfirm();

  const { role } = use(CurrentUser);
  const availableRoles = getAssignableRoles(role);

  const profile = useFragment<PeopleListItemFragment$key>(fragment, fKey);
  const lastInvitation = profile.lastInvitation.edges[0]?.node;

  const roleOptions = availableRoles.includes(profile.membership.role)
    ? availableRoles
    : [...availableRoles, profile.membership.role];

  const isActive = profile.state === "ACTIVE";
  const isInactive = profile.state === "DEACTIVATED";

  const canSendActivationMail = !isActive && profile.source !== "SCIM" && profile.canInvite;
  const canDeactivate = profile.canDeactivate && profile.source !== "SCIM" && profile.state !== "DEACTIVATED";
  const canRemove = profile.canRemoveMember && profile.source !== "SCIM";

  const [inviteUser]
    = useMutationWithToasts<PeopleListItem_inviteMutation>(inviteUserMutation, {
      successMessage: t("peopleListItem.messages.invitationSent"),
      errorMessage: t("peopleListItem.errors.sendInvitation"),
    });
  const [updateMembership, isUpdatingRole] = useMutationWithToasts(
    updateRoleMutation,
    {
      successMessage: t("peopleListItem.messages.roleUpdated"),
      errorMessage: t("peopleListItem.errors.updateRole"),
    },
  );
  const [deactivateUser, isDeactivating] = useMutationWithToasts(
    deactivateUserMutation,
    {
      successMessage: t("peopleListItem.messages.deactivated"),
      errorMessage: t("peopleListItem.errors.deactivate"),
    },
  );
  const [removeUser, isRemoving] = useMutationWithToasts(
    removeUserMutation,
    {
      successMessage: t("peopleListItem.messages.removed"),
      errorMessage: t("peopleListItem.errors.remove"),
    },
  );
  const isMutating = isDeactivating || isRemoving;

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
          updater: (store) => {
            // Inviting a deactivated user marks the profile pending server-side;
            // the payload only returns the invitation edge, so update state here.
            store.get(profile.id)?.setValue("PENDING", "state");
          },
        });
      },
      {
        label: t("peopleListItem.actions.send"),
        variant: "primary",
        message: t("peopleListItem.confirmations.sendActivationEmail", { name: profile.fullName }),
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
  const handleDeactivate = () => {
    confirm(
      () => {
        return deactivateUser({
          variables: {
            input: {
              profileId: profile.id,
              organizationId: organizationId,
            },
          },
          onCompleted: () => {
            onRefetch();
          },
        });
      },
      {
        message: t("peopleListItem.confirmations.deactivate", { name: profile.fullName }),
      },
    );
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
          onCompleted: () => {
            onRefetch();
          },
        });
      },
      {
        message: t("peopleListItem.confirmations.remove", { name: profile.fullName }),
      },
    );
  };

  return (
    <Tr to={`/organizations/${organizationId}/people/${profile.id}`}>
      <Td className={clsx(
        isMutating && "opacity-60 pointer-events-none",
        isInactive && "opacity-50",
      )}
      >
        <span className="font-semibold">{profile.fullName}</span>
      </Td>
      <Td>
        <Badge variant={isActive ? "success" : "neutral"}>{profile.state}</Badge>
      </Td>
      <Td className={clsx(
        isMutating && "opacity-60 pointer-events-none",
        isInactive && "opacity-50",
      )}
      >
        <div className="flex items-center gap-2">
          {profile.emailAddress}
          <Badge variant="info">{profile.source}</Badge>
        </div>
      </Td>
      {availableRoles.length > 0 && (
        <Td
          noLink
          className={clsx(
            "pr-4",
            isMutating && "opacity-60 pointer-events-none",
            isInactive && "opacity-50",
          )}
        >
          <Select
            disabled={!profile.membership.canUpdate || isUpdatingRole}
            value={profile.membership.role}
            onValueChange={role => void handleUpdateRole(role)}
          >
            {getMembershipRoles(t)
              .filter(({ value }) => roleOptions.includes(value))
              .map(({ value, label }) => (
                <Option key={value} value={value}>
                  {label}
                </Option>
              ))}
          </Select>
        </Td>
      )}
      <Td className={clsx(
        isMutating && "opacity-60 pointer-events-none",
        isInactive && "opacity-50",
      )}
      >
        {dateFormat(i18n.language, profile.createdAt)}
      </Td>
      <Td className={clsx(
        isMutating && "opacity-60 pointer-events-none",
        isInactive && "opacity-50",
      )}
      >
        {profile.contractEndDate
          ? (
              <time dateTime={profile.contractEndDate}>
                {dateFormat(i18n.language, profile.contractEndDate)}
              </time>
            )
          : (
              <span className="text-txt-tertiary">—</span>
            )}
      </Td>
      <Td noLink width={160} className="text-end">
        {(canSendActivationMail || canDeactivate || canRemove) && (
          <ActionDropdown>
            {canSendActivationMail && (
              <DropdownItem
                onClick={handleInvite}
                icon={IconMail}
              >
                {lastInvitation ? t("peopleListItem.actions.resendActivationMail") : t("peopleListItem.actions.sendActivationMail")}
              </DropdownItem>
            )}
            {canDeactivate && (
              <DropdownItem
                onClick={handleDeactivate}
                icon={IconArchive}
              >
                {t("peopleListItem.actions.deactivatePerson")}
              </DropdownItem>
            )}
            {canRemove && (
              <DropdownItem
                onClick={handleRemove}
                variant="danger"
                icon={IconTrashCan}
              >
                {t("peopleListItem.actions.removePerson")}
              </DropdownItem>
            )}
          </ActionDropdown>
        )}
      </Td>
    </Tr>
  );
}
