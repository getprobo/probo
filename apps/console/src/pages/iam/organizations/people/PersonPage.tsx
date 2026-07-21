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

import { ActionDropdown, Avatar, Badge, Breadcrumb, Card, DropdownItem, IconArchive, IconTrashCan, useConfirm } from "@probo/ui";
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
        canDelete: permission(action: "iam:membership-profile:delete")
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

const archiveUserMutation = graphql`
  mutation PersonPage_archiveMutation(
    $input: ArchiveUserInput!
  ) {
    archiveUser(input: $input) {
      archivedProfileId
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

  const [archiveUser, isArchiving] = useMutationWithToasts(
    archiveUserMutation,
    {
      successMessage: t("personPage.messages.archived"),
      errorMessage: t("personPage.errors.archive"),
    },
  );
  const [removeUser, isRemoving] = useMutationWithToasts(
    removeUserMutation,
    {
      successMessage: t("personPage.messages.removed"),
      errorMessage: t("personPage.errors.remove"),
    },
  );
  const isMutating = isArchiving || isRemoving;

  const handleArchive = () => {
    confirm(
      () => {
        return archiveUser({
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
        message: t("personPage.confirmations.archive", { name: person.fullName }),
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
            void navigate(`/organizations/${organizationId}/people`);
          },
        });
      },
      {
        message: t("personPage.confirmations.remove", { name: person.fullName }),
      },
    );
  };

  const canArchive = person.canDelete && person.source !== "SCIM" && person.state !== "INACTIVE";
  const canRemove = person.canRemoveMember && person.source !== "SCIM";

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t("personPage.breadcrumb.people"),
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
        {(canArchive || canRemove) && (
          <ActionDropdown variant="secondary">
            {canArchive && (
              <DropdownItem
                icon={IconArchive}
                onClick={handleArchive}
                disabled={isMutating}
              >
                {t("personPage.actions.archive")}
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
