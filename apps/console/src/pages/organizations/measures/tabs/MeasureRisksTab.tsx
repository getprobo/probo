import { graphql, useFragment } from "react-relay";
import { useOutletContext } from "react-router";

import type { MeasureRisksTabFragment$key } from "#/__generated__/core/MeasureRisksTabFragment.graphql";
import { LinkedRisksCard } from "#/components/risks/LinkedRisksCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const risksFragment = graphql`
  fragment MeasureRisksTabFragment on Measure {
    id
    risks(first: 100) @connection(key: "Measure__risks") {
      __id
      edges {
        node {
          id
          canCreateMeasureMapping: permission(
            action: "core:risk:create-measure-mapping"
          )
          canDeleteMeasureMapping: permission(
            action: "core:risk:delete-measure-mapping"
          )
          ...LinkedRisksCardFragment
        }
      }
    }
  }
`;

const attachRiskMutation = graphql`
  mutation MeasureRisksTabCreateMutation(
    $input: CreateRiskMeasureMappingInput!
    $connections: [ID!]!
  ) {
    createRiskMeasureMapping(input: $input) {
      riskEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedRisksCardFragment
        }
      }
    }
  }
`;

export const detachRiskMutation = graphql`
  mutation MeasureRisksTabDetachMutation(
    $input: DeleteRiskMeasureMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskMeasureMapping(input: $input) {
      deletedRiskId @deleteEdge(connections: $connections)
    }
  }
`;

export default function MeasureRisksTab() {
  const { measure } = useOutletContext<{
    measure: MeasureRisksTabFragment$key & { id: string };
  }>();
  const data = useFragment(risksFragment, measure);
  const connectionId = data.risks.__id;
  const risks = data.risks?.edges?.map(edge => edge.node) ?? [];

  const canLinkRisk = risks.some(
    ({ canCreateMeasureMapping }) => canCreateMeasureMapping,
  );
  const canUnlinkRisk = risks.some(
    ({ canDeleteMeasureMapping }) => canDeleteMeasureMapping,
  );
  const readOnly = !canLinkRisk && !canUnlinkRisk;

  const incrementOptions = {
    id: data.id,
    node: "risks(first:0)",
  };
  const [detachRisk, isDetaching] = useMutationWithIncrement(
    detachRiskMutation,
    {
      ...incrementOptions,
      value: -1,
    },
  );
  const [attachRisk, isAttaching] = useMutationWithIncrement(
    attachRiskMutation,
    {
      ...incrementOptions,
      value: 1,
    },
  );
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedRisksCard
      disabled={isLoading}
      risks={risks}
      onAttach={attachRisk}
      onDetach={detachRisk}
      params={{ measureId: data.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
