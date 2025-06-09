import { graphql, useFragment, useMutation } from "react-relay";
import type { MeasureRisksTabFragment$key } from "./__generated__/MeasureRisksTabFragment.graphql";
import { useOutletContext } from "react-router";
import { LinkedRisksCard } from "/components/risks/LinkedRisksCard";

const risksFragment = graphql`
  fragment MeasureRisksTabFragment on Measure {
    id
    risks(first: 100) @connection(key: "Measure__risks") {
      __id
      edges {
        node {
          id
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
  const risks = data.risks?.edges?.map((edge) => edge.node) ?? [];

  const [detachRisk, isDetaching] = useMutation(detachRiskMutation);
  const [attachRisk, isAttaching] = useMutation(attachRiskMutation);
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedRisksCard
      disabled={isLoading}
      risks={risks}
      onAttach={attachRisk}
      onDetach={detachRisk}
      params={{ measureId: data.id }}
      connectionId={connectionId}
    />
  );
}
