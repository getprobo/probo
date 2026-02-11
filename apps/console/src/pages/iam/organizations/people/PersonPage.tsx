import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { ActionDropdown, Avatar, Breadcrumb, DropdownItem, IconTrashCan, useConfirm } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { PersonPageQuery } from "#/__generated__/iam/PersonPageQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { PersonForm } from "./_components/PersonForm";

export const personPageQuery = graphql`
  query PersonPageQuery($personId: ID!) {
    person: node(id: $personId) @required(action: THROW) {
      __typename
      ... on MembershipProfile {
        fullName
        membershipId
        identity @required(action: THROW) {
          email
        }
        canDelete: permission(action: "iam:membership-profile:delete")
        ...PersonFormFragment
      }
    }
  }
`;

const removeMemberMutation = graphql`
  mutation PersonPage_removeMutation(
    $input: RemoveMemberInput!
  ) {
    removeMember(input: $input) {
      deletedMembershipId
    }
  }
`;

export function PersonPage(props: { queryRef: PreloadedQuery<PersonPageQuery> }) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const confirm = useConfirm();
  const navigate = useNavigate();

  const { person } = usePreloadedQuery<PersonPageQuery>(personPageQuery, queryRef);
  if (person.__typename !== "MembershipProfile") {
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
              membershipId: person.membershipId,
              organizationId: organizationId,
            },
          },
          onCompleted: () => {
            void navigate(`/organizations/${organizationId}/people`);
          },
        });
      },
      {
        message: sprintf(
          __("Are you sure you want to remove %s?"),
          person.fullName,
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
            to: `/organizations/${organizationId}/people`,
          },
          {
            label: person.fullName,
          },
        ]}
      />
      <div className="flex justify-between">
        <div className="flex items-center gap-6">
          <Avatar name={person.fullName} size="xl" />
          <div>

            <div className="text-2xl">{person.fullName}</div>
            <div className="text-lg text-txt-secondary">{person.identity.email}</div>
          </div>
        </div>
        {person.canDelete && (
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

      <PersonForm fragmentRef={person} />
    </div>
  );
};
