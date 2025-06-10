import {
  graphql,
  useMutation,
  useRefetchableFragment,
} from "react-relay";
import { useOutletContext } from "react-router";
import { LinkedControlsCard } from "/components/controls/LinkedControlsCard";
import type { MeasureControlsTabFragment$key } from "./__generated__/MeasureControlsTabFragment.graphql";

export const controlsFragment = graphql`
  fragment MeasureControlsTabFragment on Measure
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey" }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    order: { type: "ControlOrder", defaultValue: null }
    filter: { type: "ControlFilter", defaultValue: null }
  )
  @refetchable(queryName: "MeasureControlsTabControlsQuery") {
    id
    controls(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "MeasureControlsTab_controls") {
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

export const attachControlMutation = graphql`
  mutation MeasureControlsTabAttachMutation(
    $input: CreateControlMeasureMappingInput!
    $connections: [ID!]!
  ) {
    createControlMeasureMapping(input: $input) {
      controlEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedControlsCardFragment
        }
      }
    }
  }
`;

export default function MeasureControlsTab() {
  const { measure } = useOutletContext<{
    measure: MeasureControlsTabFragment$key & { id: string };
  }>();
  const [data, refetch] = useRefetchableFragment(controlsFragment, measure);
  const connectionId = data.controls.__id;
  const controls = data.controls?.edges?.map((edge) => edge.node) ?? [];

  const [detachControl, isDetaching] = useMutation(detachControlMutation);
  const [attachControl, isAttaching] = useMutation(attachControlMutation);
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedControlsCard
      disabled={isLoading}
      controls={controls}
      onDetach={detachControl}
      onAttach={attachControl}
      params={{ measureId: data.id }}
      connectionId={connectionId}
      refetch={refetch}
    />
  );
}
