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

import { usePageTitle } from "@probo/hooks";
import { ActionDropdown, Avatar, Badge, Card, DropdownItem, IconArchive, IconTrashCan, useConfirm } from "@probo/ui";
import { useTranslation } from "react-i18next";
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
        avatar {
          downloadUrl
        }
        canDeactivate: permission(action: "iam:membership-profile:deactivate")
        canRemoveMember: permission(action: "iam:membership:delete")
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

const deactivateUserMutation = graphql`
  mutation PersonPage_deactivateMutation(
    $input: DeactivateUserInput!
  ) {
    deactivateUser(input: $input) {
      success
    }
  }
`;

export function PersonPage(props: { queryRef: PreloadedQuery<PersonPageQuery> }) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const confirm = useConfirm();
  const navigate = useNavigate();

  const { person } = usePreloadedQuery<PersonPageQuery>(personPageQuery, queryRef);
  if (person.__typename !== "Profile") {
    throw new Error("invalid type for node");
  }

  usePageTitle(person.fullName);

  const [deactivateUser, isDeactivating] = useMutationWithToasts(
    deactivateUserMutation,
    {
      successMessage: t("personPage.messages.deactivated"),
      errorMessage: t("personPage.errors.deactivate"),
    },
  );
  const [removeUser, isRemoving] = useMutationWithToasts(
    removeUserMutation,
    {
      successMessage: t("personPage.messages.removed"),
      errorMessage: t("personPage.errors.remove"),
    },
  );
  const isMutating = isDeactivating || isRemoving;

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
          onCompleted: () => {
            void navigate(`/organizations/${organizationId}/settings/people`);
          },
        });
      },
      {
        message: t("personPage.confirmations.deactivate", { name: person.fullName }),
      },
    );
  };

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
            void navigate(`/organizations/${organizationId}/settings/people`);
          },
        });
      },
      {
        message: t("personPage.confirmations.remove", { name: person.fullName }),
      },
    );
  };

  const canDeactivate = person.canDeactivate && person.source !== "SCIM" && person.state !== "DEACTIVATED";
  const canRemove = person.canRemoveMember && person.source !== "SCIM";

  return (
    <div className="space-y-6">
      <div className="flex justify-between">
        <div className="flex items-center gap-6">
          <Avatar name={person.fullName} src={person.avatar?.downloadUrl} size="xl" />
          <div>
            <div className="flex items-center gap-2">
              <span className="text-2xl">{person.fullName}</span>
              <Badge variant="info">{person.source}</Badge>
            </div>
            <div className="text-lg text-txt-secondary">{person.emailAddress}</div>
          </div>
        </div>
        {(canDeactivate || canRemove) && (
          <ActionDropdown variant="secondary">
            {canDeactivate && (
              <DropdownItem
                icon={IconArchive}
                onClick={handleDeactivate}
                disabled={isMutating}
              >
                {t("personPage.actions.deactivate")}
              </DropdownItem>
            )}
            {canRemove && (
              <DropdownItem
                variant="danger"
                icon={IconTrashCan}
                onClick={handleRemove}
                disabled={isMutating}
              >
                {t("personPage.actions.remove")}
              </DropdownItem>
            )}
          </ActionDropdown>
        )}
      </div>

      <Card padded className="space-y-4">
        <PersonFormLoader fragmentRef={person} />
      </Card>
    </div>
  );
};
