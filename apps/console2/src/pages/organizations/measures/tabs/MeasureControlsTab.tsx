import { graphql, useFragment, useMutation } from "react-relay";
import { useOutletContext } from "react-router";
import { LinkedControlsCard } from "/components/controls/LinkedControlsCard";
import type { MeasureControlsTabFragment$key } from "./__generated__/MeasureControlsTabFragment.graphql";

const ControlsFragment = graphql`
  fragment MeasureControlsTabFragment on Measure {
    id
    controls(first: 100)
      @connection(key: "MeasureControlsTabFragment_controls") {
      __id
      edges {
        node {
          id
          ...LinkedControlsCardFragment
        }
      }
    }
  }
`;

export const detachControlMutation = graphql`
  mutation MeasureControlsTabDetachMutation(
    $input: DeleteControlMeasureMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlMeasureMapping(input: $input) {
      deletedControlId @deleteEdge(connections: $connections)
    }
  }
`;

export default function MeasureControlsTab() {
  const { measure } = useOutletContext<{
    measure: MeasureControlsTabFragment$key & { id: string };
  }>();
  const data = useFragment(ControlsFragment, measure);
  const connectionId = data.controls.__id;
  const controls = data.controls?.edges?.map((edge) => edge.node) ?? [];

  const [detachControl, isDetaching] = useMutation(detachControlMutation);

  return (
    <LinkedControlsCard
      disabled={isDetaching}
      controls={controls}
      onDetach={detachControl}
      params={{ measureId: data.id }}
      connectionId={connectionId}
    />
  );
}
