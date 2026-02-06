import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { ActionDropdown, Avatar, Breadcrumb, DropdownItem, IconTrashCan, useConfirm } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { UserPageQuery } from "#/__generated__/iam/UserPageQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { UserForm } from "./_components/UserForm";

export const userPageQuery = graphql`
  query UserPageQuery($userId: ID!) {
    user: node(id: $userId) @required(action: THROW) {
      __typename
      ... on MembershipProfile {
        fullName
        membershipId
        identity @required(action: THROW) {
          email
        }
        canDelete: permission(action: "iam:membership-profile:delete")
        ...UserFormFragment
      }
    }
  }
`;

const removeMemberMutation = graphql`
  mutation UserPage_removeMutation(
    $input: RemoveMemberInput!
  ) {
    removeMember(input: $input) {
      deletedMembershipId
    }
  }
`;

export function UserPage(props: { queryRef: PreloadedQuery<UserPageQuery> }) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const navigate = useNavigate();

  const { user } = usePreloadedQuery<UserPageQuery>(userPageQuery, queryRef);
  if (user.__typename !== "MembershipProfile") {
    throw new Error("invalid type for node");
  }

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
              membershipId: user.membershipId,
              organizationId: organizationId,
            },
          },
          onCompleted: () => {
            void navigate(`/organizations/${organizationId}/users`);
          },
        });
      },
      {
        message: sprintf(
          __("Are you sure you want to remove %s?"),
          user.fullName,
        ),
      },
    );
  };

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("People"),
            to: `/organizations/${organizationId}/users`,
          },
          {
            label: user.fullName,
          },
        ]}
      />
      <div className="flex justify-between">
        <div className="flex items-center gap-6">
          <Avatar name={user.fullName} size="xl" />
          <div>

            <div className="text-2xl">{user.fullName}</div>
            <div className="text-lg text-txt-secondary">{user.identity.email}</div>
          </div>
        </div>
        {user.canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={handleRemove}
              disabled={isRemoving}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <UserForm fragmentRef={user} />
    </div>
  );
};
