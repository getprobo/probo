import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";

import type { MeasureGraphDeleteMutation } from "#/__generated__/core/MeasureGraphDeleteMutation.graphql";

import { useMutationWithToasts } from "../useMutationWithToasts";

export const measuresQuery = graphql`
  query MeasureGraphListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        canCreateMeasure: permission(action: "core:measure:create")
        ...MeasuresPageFragment
      }
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

export const MeasureConnectionKey = "MeasuresGraphListQuery__measures";

export function useDeleteMeasureMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<MeasureGraphDeleteMutation>(
    deleteMeasureMutation,
    {
      successMessage: __("Measure deleted successfully."),
      errorMessage: __("Failed to delete measure"),
    },
  );
}

export const measureNodeQuery = graphql`
  query MeasureGraphNodeQuery($measureId: ID!) {
    node(id: $measureId) {
      ... on Measure {
        id
        name
        description
        state
        category
        canUpdate: permission(action: "core:measure:update")
        canDelete: permission(action: "core:measure:delete")
        canListTasks: permission(action: "core:task:list")
        evidencesInfos: evidences(first: 0) {
          totalCount
        }
        risksInfos: risks(first: 0) {
          totalCount
        }
        controlsInfos: controls(first: 0) {
          totalCount
        }
        ...MeasureRisksTabFragment
        ...MeasureControlsTabFragment
        ...MeasureFormDialogMeasureFragment
        ...MeasureEvidencesTabFragment
      }
    }
  }
`;

const measureUpdateMutation = graphql`
  mutation MeasureGraphUpdateMutation($input: UpdateMeasureInput!) {
    updateMeasure(input: $input) {
      measure {
        ...MeasureFormDialogMeasureFragment
      }
    }
  }
`;

export const useUpdateMeasure = () => {
  const { __ } = useTranslate();

  return useMutationWithToasts(measureUpdateMutation, {
    successMessage: __("Measure updated successfully."),
    errorMessage: __("Failed to update measure"),
  });
};
