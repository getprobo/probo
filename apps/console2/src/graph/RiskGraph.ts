import { graphql } from "relay-runtime";
import { useTranslate } from "@probo/i18n";
import { useMutationWithToasts } from "../hooks/useMutationWithToasts";
import type { RiskGraphDeleteMutation } from "./__generated__/RiskGraphDeleteMutation.graphql";

const deleteRiskMutation = graphql`
  mutation RiskGraphDeleteMutation(
    $input: DeleteRiskInput!
    $connections: [ID!]!
  ) {
    deleteRisk(input: $input) {
      deletedRiskId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteRiskMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<RiskGraphDeleteMutation>(deleteRiskMutation, {
    successMessage: __("Risk deleted successfully."),
    errorMessage: __("Failed to delete risk. Please try again."),
  });
}
