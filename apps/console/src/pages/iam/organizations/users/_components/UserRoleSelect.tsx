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

import { getAssignableRoles, getMembershipRole, getMembershipRoles, type Role } from "@probo/helpers";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { use } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { UserRoleSelect_membership$key } from "#/__generated__/iam/UserRoleSelect_membership.graphql";
import type { UserRoleSelect_updateMutation } from "#/__generated__/iam/UserRoleSelect_updateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { CurrentUser } from "#/providers/CurrentUser";

import { isUsersListRole } from "../_lib/useUsersListFilters";
import { userListItem } from "../variants";

const fragment = graphql`
  fragment UserRoleSelect_membership on Membership {
    id
    role
    canUpdate: permission(action: "iam:membership:update")
  }
`;

const updateMembershipMutation = graphql`
  mutation UserRoleSelect_updateMutation($input: UpdateMembershipInput!) {
    updateMembership(input: $input) {
      membership {
        id
        role
      }
    }
  }
`;

interface UserRoleSelectProps {
  "membershipKey": UserRoleSelect_membership$key;
  "size"?: 1 | 2;
  "id"?: string;
  "aria-describedby"?: string;
  "aria-invalid"?: boolean;
}

export function UserRoleSelect({
  membershipKey,
  size = 1,
  id,
  "aria-describedby": ariaDescribedBy,
  "aria-invalid": ariaInvalid,
}: UserRoleSelectProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { role: viewerRole } = use(CurrentUser);
  const membership = useFragment(fragment, membershipKey);
  const [updateMembership, isUpdating] = useMutation<UserRoleSelect_updateMutation>(
    updateMembershipMutation,
    {
      successMessage: t("userListItem.messages.roleUpdated"),
      errorToast: t("userListItem.errors.updateRole"),
    },
  );
  const { role, roleSelect } = userListItem();
  const assignableRoles = getAssignableRoles(viewerRole);
  const roleOptions = getMembershipRoles(t).filter(({ value }) => (
    assignableRoles.includes(value) || value === membership.role
  ));

  function handleRoleChange(value: string | null) {
    if (value == null || !isUsersListRole(value) || value === membership.role) {
      return;
    }

    void updateMembership({
      variables: {
        input: {
          organizationId,
          membershipId: membership.id,
          role: value,
        },
      },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  const trigger = (
    <Select
      value={membership.role}
      disabled={isUpdating}
      onValueChange={handleRoleChange}
    >
      <SelectTrigger
        id={id}
        size={size}
        aria-label={t("usersList.filters.role")}
        aria-describedby={ariaDescribedBy}
        aria-invalid={ariaInvalid}
      >
        {(value: Role | null) => (
          value != null ? getMembershipRole(t, value) : null
        )}
      </SelectTrigger>
      <SelectPopup align={size === 1 ? "end" : "start"}>
        {roleOptions.map(({ value, label }) => (
          <SelectItem key={value} value={value}>
            {label}
          </SelectItem>
        ))}
      </SelectPopup>
    </Select>
  );

  if (size === 2) {
    if (!membership.canUpdate) {
      return (
        <Text size={2} weight="medium" highContrast>
          {getMembershipRole(t, membership.role)}
        </Text>
      );
    }
    return trigger;
  }

  return (
    <span className={role()}>
      <Text size={1} color="faint">
        {t("usersList.filters.role")}
      </Text>
      {membership.canUpdate
        ? (
            <div className={roleSelect()}>
              {trigger}
            </div>
          )
        : (
            <Text size={1} weight="medium" highContrast>
              {getMembershipRole(t, membership.role)}
            </Text>
          )}
    </span>
  );
}
