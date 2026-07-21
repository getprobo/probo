// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { Button, Card, IconTrashCan } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { GeneralSettingsPage_deleteMutation } from "#/__generated__/iam/GeneralSettingsPage_deleteMutation.graphql";
import type { GeneralSettingsPageQuery } from "#/__generated__/iam/GeneralSettingsPageQuery.graphql";
import { DeleteOrganizationDialog } from "#/components/organizations/DeleteOrganizationDialog";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { OrganizationForm } from "./_components/OrganizationForm";

export const generalSettingsPageQuery = graphql`
  query GeneralSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        name @required(action: THROW)
        canDelete: permission(action: "iam:organization:delete")
        ...OrganizationFormFragment
      }
    }
  }
`;

const deleteOrganizationMutation = graphql`
  mutation GeneralSettingsPage_deleteMutation(
    $input: DeleteOrganizationInput!
    $connections: [ID!]!
  ) {
    deleteOrganization(input: $input) {
      deletedOrganizationId @deleteEdge(connections: $connections)
    }
  }
`;

export function GeneralSettingsPage(props: {
  queryRef: PreloadedQuery<GeneralSettingsPageQuery>;
}) {
  const { queryRef } = props;
  const { t } = useTranslation();
  const navigate = useNavigate();

  const { organization } = usePreloadedQuery<GeneralSettingsPageQuery>(
    generalSettingsPageQuery,
    queryRef,
  );
  if (organization.__typename === "%other") {
    throw new Error("Relay node is not an organization");
  }

  const [deleteOrganization, isDeletingOrganization]
    = useMutationWithToasts<GeneralSettingsPage_deleteMutation>(
      deleteOrganizationMutation,
      {
        successMessage: t("generalSettingsPage.messages.deleted"),
        errorMessage: t("generalSettingsPage.errors.delete"),
      },
    );

  const handleDeleteOrganization = () => {
    return deleteOrganization({
      variables: {
        input: {
          organizationId: organization.id,
        },
        connections: [],
      },
      onSuccess: () => {
        void navigate("/", { replace: true });
      },
    });
  };

  return (
    <div className="space-y-6">
      <OrganizationForm fKey={organization} />

      {organization.canDelete && (
        <div className="space-y-4 mt-12">
          <h2 className="text-base font-medium text-red-600">
            {t("generalSettingsPage.dangerZone.title")}
          </h2>
          <Card padded className="border-red-200 flex items-center gap-3">
            <div className="mr-auto">
              <h3 className="text-base font-semibold text-red-700">
                {t("generalSettingsPage.dangerZone.delete.title")}
              </h3>
              <p className="text-sm text-txt-tertiary">
                {t("generalSettingsPage.dangerZone.delete.description")}
                {" "}
                <span className="text-red-600 font-medium">
                  {t("generalSettingsPage.dangerZone.delete.warning")}
                </span>
              </p>
            </div>
            <DeleteOrganizationDialog
              organizationName={organization.name}
              onConfirm={() => void handleDeleteOrganization()}
              isDeleting={isDeletingOrganization}
            >
              <Button
                variant="danger"
                icon={IconTrashCan}
                disabled={isDeletingOrganization}
              >
                {t("generalSettingsPage.dangerZone.delete.action")}
              </Button>
            </DeleteOrganizationDialog>
          </Card>
        </div>
      )}
    </div>
  );
}
