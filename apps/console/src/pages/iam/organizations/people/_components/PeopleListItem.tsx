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

import { getAssignableRoles } from "@probo/helpers";
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
    canUpdate: permission(action: "iam:membership-profile:update")
    canInvite: permission(action: "iam:invitation:create")
    canDelete: permission(action: "iam:membership-profile:delete")
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

const archiveUserMutation = graphql`
  mutation PeopleListItem_archiveMutation($input: ArchiveUserInput!) {
    archiveUser(input: $input) {
      archivedProfileId
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

  const isInactive = profile.state === "INACTIVE";

  const canSendActivationMail = isInactive && profile.source !== "SCIM" && profile.canInvite;
  const canArchive = profile.canDelete && profile.source !== "SCIM" && profile.state !== "INACTIVE";
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
  const [archiveUser, isArchiving] = useMutationWithToasts(
    archiveUserMutation,
    {
      successMessage: t("peopleListItem.messages.archived"),
      errorMessage: t("peopleListItem.errors.archive"),
    },
  );
  const [removeUser, isRemoving] = useMutationWithToasts(
    removeUserMutation,
    {
      successMessage: t("peopleListItem.messages.removed"),
      errorMessage: t("peopleListItem.errors.remove"),
    },
  );
  const isMutating = isArchiving || isRemoving;

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
  const handleArchive = () => {
    confirm(
      () => {
        return archiveUser({
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
        message: t("peopleListItem.confirmations.archive", { name: profile.fullName }),
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
        <Badge variant={profile.state === "INACTIVE" ? "neutral" : "success"}>{profile.state}</Badge>
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
            {roleOptions.includes("OWNER") && (
              <Option value="OWNER">{t("peopleListItem.roles.owner")}</Option>
            )}
            {roleOptions.includes("ADMIN") && (
              <Option value="ADMIN">{t("peopleListItem.roles.admin")}</Option>
            )}
            {roleOptions.includes("VIEWER") && (
              <Option value="VIEWER">{t("peopleListItem.roles.viewer")}</Option>
            )}
            {roleOptions.includes("AUDITOR") && (
              <Option value="AUDITOR">{t("peopleListItem.roles.auditor")}</Option>
            )}
            {roleOptions.includes("EMPLOYEE") && (
              <Option value="EMPLOYEE">{t("peopleListItem.roles.employee")}</Option>
            )}
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
      <Td noLink width={160} className="text-end">
        {(canSendActivationMail || canArchive || canRemove) && (
          <ActionDropdown>
            {canSendActivationMail && (
              <DropdownItem
                onClick={handleInvite}
                icon={IconMail}
              >
                {lastInvitation ? t("peopleListItem.actions.resendActivationMail") : t("peopleListItem.actions.sendActivationMail")}
              </DropdownItem>
            )}
            {canArchive && (
              <DropdownItem
                onClick={handleArchive}
                icon={IconArchive}
              >
                {t("peopleListItem.actions.archivePerson")}
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
