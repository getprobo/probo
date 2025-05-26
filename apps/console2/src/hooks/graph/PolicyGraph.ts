import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "../useMutationWithToasts";
import type { PolicyGraphDeleteMutation } from "./__generated__/PolicyGraphDeleteMutation.graphql";

export const policiesQuery = graphql`
  query PolicyGraphListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...PoliciesPageListFragment
    }
  }
`;

const deletePolicyMutation = graphql`
  mutation PolicyGraphDeleteMutation(
    $input: DeletePolicyInput!
    $connections: [ID!]!
  ) {
    deletePolicy(input: $input) {
      deletedPolicyId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeletePolicyMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<PolicyGraphDeleteMutation>(
    deletePolicyMutation,
    {
      successMessage: __("Policy deleted successfully."),
      errorMessage: __("Failed to delete policy. Please try again."),
    }
  );
}

export const policyNodeQuery = graphql`
  query PolicyGraphNodeQuery($policyId: ID!) {
    node(id: $policyId) {
      ... on Policy {
        ...PolicyPagePolicyFragment
      }
    }
  }
`;
