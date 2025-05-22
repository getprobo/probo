import { graphql } from "relay-runtime";
import { useTranslate } from "@probo/i18n";
import { useMutationWithToasts } from "../hooks/useMutationWithToasts";
import type { RisksDeleteMutation } from "./__generated__/RisksDeleteMutation.graphql";

const deleteRiskMutation = graphql`
  mutation RisksDeleteMutation($input: DeleteRiskInput!, $connections: [ID!]!) {
    deleteRisk(input: $input) {
      deletedRiskId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteRiskMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<RisksDeleteMutation>(deleteRiskMutation, {
    successMessage: __("Risk deleted successfully."),
    errorMessage: __("Failed to delete risk. Please try again."),
  });
}
