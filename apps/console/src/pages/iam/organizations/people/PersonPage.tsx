// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Avatar,
  Badge,
  Breadcrumb,
  Card,
  DropdownItem,
  IconCircleCheck,
  IconCircleX,
  IconTrashCan,
  useConfirm,
} from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { PersonPageQuery } from "#/__generated__/iam/PersonPageQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { PersonFormLoader } from "./_components/PersonForm";

export const personPageQuery = graphql`
  query PersonPageQuery($personId: ID!) {
    person: node(id: $personId) @required(action: THROW) {
      __typename
      ... on Profile {
        id
        fullName
        emailAddress
        source
        state
        canDelete: permission(action: "iam:membership-profile:delete")
        canActivate: permission(action: "iam:membership-profile:activate")
        canDeactivate: permission(action: "iam:membership-profile:deactivate")
        ...PersonFormFragment
      }
    }
  }
`;

const removeUserMutation = graphql`
  mutation PersonPage_removeMutation(
    $input: RemoveUserInput!
  ) {
    removeUser(input: $input) {
      deletedProfileId
    }
  }
`;

const activateUserMutation = graphql`
  mutation PersonPage_activateMutation($input: ActivateUserInput!) {
    activateUser(input: $input) {
      profile {
        id
        state
      }
    }
  }
`;

const deactivateUserMutation = graphql`
  mutation PersonPage_deactivateMutation($input: DeactivateUserInput!) {
    deactivateUser(input: $input) {
      profile {
        id
        state
      }
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
  if (person.__typename !== "Profile") {
    throw new Error("invalid type for node");
  }

  const [removeUser, isRemoving] = useMutationWithToasts(
    removeUserMutation,
    {
      successMessage: __("Person removed successfully"),
      errorMessage: __("Failed to remove person"),
    },
  );
  const [activateUser, isActivating] = useMutationWithToasts(
    activateUserMutation,
    {
      successMessage: __("Person activated successfully"),
      errorMessage: __("Failed to activate person"),
    },
  );
  const [deactivateUser, isDeactivating] = useMutationWithToasts(
    deactivateUserMutation,
    {
      successMessage: __("Person deactivated successfully"),
      errorMessage: __("Failed to deactivate person"),
    },
  );

  const handleRemove = () => {
    confirm(
      () => {
        return removeUser({
          variables: {
            input: {
              profileId: person.id,
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
  const handleDeactivate = () => {
    confirm(
      () => {
        return deactivateUser({
          variables: {
            input: {
              profileId: person.id,
              organizationId: organizationId,
            },
          },
        });
      },
      {
        label: __("Deactivate"),
        message: sprintf(
          __("Deactivate %s? They will keep their profile but lose access until reactivated."),
          person.fullName,
        ),
      },
    );
  };
  const handleActivate = () => {
    confirm(
      () => {
        return activateUser({
          variables: {
            input: {
              profileId: person.id,
              organizationId: organizationId,
            },
          },
        });
      },
      {
        label: __("Activate"),
        variant: "primary",
        message: sprintf(
          __("Reactivate %s? They will regain access to the organization."),
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
            <div className="flex items-center gap-2">
              <span className="text-2xl">{person.fullName}</span>
              <Badge variant="info">{person.source}</Badge>
            </div>
            <div className="text-lg text-txt-secondary">{person.emailAddress}</div>
          </div>
        </div>
        {person.source !== "SCIM" && (() => {
          const isInactive = person.state === "INACTIVE";
          const showActivate = isInactive && person.canActivate;
          const showDeactivate = !isInactive && person.canDeactivate;
          const showDelete = person.canDelete;
          const isMutating = isRemoving || isActivating || isDeactivating;
          if (!showActivate && !showDeactivate && !showDelete) {
            return null;
          }
          return (
            <ActionDropdown variant="secondary">
              {showActivate && (
                <DropdownItem
                  icon={IconCircleCheck}
                  onClick={handleActivate}
                  disabled={isMutating}
                >
                  {__("Activate")}
                </DropdownItem>
              )}
              {showDeactivate && (
                <DropdownItem
                  icon={IconCircleX}
                  onClick={handleDeactivate}
                  disabled={isMutating}
                >
                  {__("Deactivate")}
                </DropdownItem>
              )}
              {showDelete && (
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={handleRemove}
                  disabled={isMutating}
                >
                  {__("Delete")}
                </DropdownItem>
              )}
            </ActionDropdown>
          );
        })()}
      </div>

      <Card padded className="space-y-4">
        <PersonFormLoader fragmentRef={person} />
      </Card>
    </div>
  );
};
