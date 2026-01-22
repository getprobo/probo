import { useTranslate } from "@probo/i18n";
import { Button, Card, IconTrashCan } from "@probo/ui";
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
  const { __ } = useTranslate();
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
        successMessage: __("Organization deleted successfully."),
        errorMessage: __("Failed to delete organization"),
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
            {__("Danger Zone")}
          </h2>
          <Card padded className="border-red-200 flex items-center gap-3">
            <div className="mr-auto">
              <h3 className="text-base font-semibold text-red-700">
                {__("Delete Organization")}
              </h3>
              <p className="text-sm text-txt-tertiary">
                {__("Permanently delete this organization and all its data.")}
                {" "}
                <span className="text-red-600 font-medium">
                  {__("This action cannot be undone.")}
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
                {__("Delete Organization")}
              </Button>
            </DeleteOrganizationDialog>
          </Card>
        </div>
      )}
    </div>
  );
}
