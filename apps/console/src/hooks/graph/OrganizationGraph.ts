import { graphql } from "relay-runtime";
import { useTranslate } from "@probo/i18n";
import { useMutationWithToasts } from "../useMutationWithToasts";
import type { OrganizationGraphDeleteMutation } from "./__generated__/OrganizationGraphDeleteMutation.graphql";

export const organizationViewQuery = graphql`
  query OrganizationGraph_ViewQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        id
        name
        ...SettingsPageFragment
        ...SettingsPageMembershipsFragment
        ...SettingsPageInvitationsFragment
      }
    }
  }
`;

const deleteOrganizationMutation = graphql`
  mutation OrganizationGraphDeleteMutation(
    $input: DeleteOrganizationInput!
    $connections: [ID!]!
  ) {
    deleteOrganization(input: $input) {
      deletedOrganizationId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteOrganizationMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<OrganizationGraphDeleteMutation>(
    deleteOrganizationMutation,
    {
      successMessage: __("Organization deleted successfully."),
      errorMessage: __("Failed to delete organization. Please try again."),
    }
  );
}
