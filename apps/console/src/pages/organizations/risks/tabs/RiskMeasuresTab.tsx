import { graphql, useFragment, useMutation } from "react-relay";
import type { RiskMeasuresTabFragment$key } from "./__generated__/RiskMeasuresTabFragment.graphql";
import { useOutletContext, useParams } from "react-router";
import { LinkedMeasuresCard } from "/components/measures/LinkedMeasuresCard";

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
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const { risk } = useOutletContext<{
    risk: RiskMeasuresTabFragment$key & { id: string };
  }>();
  const data = useFragment(measuresFragment, risk);
  const connectionId = data.measures.__id;
  const measures = data.measures?.edges?.map((edge) => edge.node) ?? [];

  const [detachMeasure, isDetaching] = useMutation(detachMeasureMutation);
  const [attachMeasure, isAttaching] = useMutation(attachMeasureMutation);
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedMeasuresCard
      disabled={isLoading}
      hideActions={isSnapshotMode}
      measures={measures}
      onAttach={attachMeasure}
      onDetach={detachMeasure}
      params={{ riskId: data.id }}
      connectionId={connectionId}
    />
  );
}
