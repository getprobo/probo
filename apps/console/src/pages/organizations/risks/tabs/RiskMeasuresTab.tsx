import { graphql, useFragment } from "react-relay";
import { useOutletContext } from "react-router";

import type { RiskGraphNodeQuery$data } from "#/__generated__/core/RiskGraphNodeQuery.graphql";
import type { RiskMeasuresTabFragment$key } from "#/__generated__/core/RiskMeasuresTabFragment.graphql";
import { LinkedMeasuresCard } from "#/components/measures/LinkedMeasuresCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const measuresFragment = graphql`
  fragment RiskMeasuresTabFragment on Risk {
    id
    measures(first: 100) @connection(key: "Risk__measures") {
      __id
      edges {
        node {
          id
          ...LinkedMeasuresCardFragment
        }
      }
    }
  }
`;

const attachMeasureMutation = graphql`
  mutation RiskMeasuresTabCreateMutation(
    $input: CreateRiskMeasureMappingInput!
    $connections: [ID!]!
  ) {
    createRiskMeasureMapping(input: $input) {
      measureEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedMeasuresCardFragment
        }
      }
    }
  }
`;

export const detachMeasureMutation = graphql`
  mutation RiskMeasuresTabDetachMutation(
    $input: DeleteRiskMeasureMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskMeasureMapping(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

export default function RiskMeasuresTab() {
  const { risk } = useOutletContext<{
    risk: RiskGraphNodeQuery$data["node"];
  }>();
  const data = useFragment<RiskMeasuresTabFragment$key>(measuresFragment, risk);
  const connectionId = data.measures.__id;
  const measures = data.measures?.edges?.map(edge => edge.node) ?? [];

  const canLinkMeasure = risk.canCreateMeasureMapping;
  const canUnlinkMeasure = risk.canDeleteMeasureMapping;
  const readOnly = !canLinkMeasure && !canUnlinkMeasure;

  const incrementOptions = {
    id: data.id,
    node: "measures(first:0)",
  };
  const [detachMeasure, isDetaching] = useMutationWithIncrement(
    detachMeasureMutation,
    {
      ...incrementOptions,
      value: -1,
    },
  );
  const [attachMeasure, isAttaching] = useMutationWithIncrement(
    attachMeasureMutation,
    {
      ...incrementOptions,
      value: 1,
    },
  );
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedMeasuresCard
      disabled={isLoading}
      measures={measures}
      onAttach={attachMeasure}
      onDetach={detachMeasure}
      params={{ riskId: data.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
