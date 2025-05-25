import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "../useMutationWithToasts";
import type { MeasureGraphDeleteMutation } from "./__generated__/MeasureGraphDeleteMutation.graphql";

export const measuresQuery = graphql`
  query MeasureGraphListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ...MeasuresPageFragment
    }
  }
`;

const deleteMeasureMutation = graphql`
  mutation MeasureGraphDeleteMutation(
    $input: DeleteMeasureInput!
    $connections: [ID!]!
  ) {
    deleteMeasure(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteMeasureMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<MeasureGraphDeleteMutation>(
    deleteMeasureMutation,
    {
      successMessage: __("Measure deleted successfully."),
      errorMessage: __("Failed to delete measure. Please try again."),
    }
  );
}
