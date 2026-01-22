import { graphql, useRefetchableFragment } from "react-relay";
import { useOutletContext } from "react-router";

import type { MeasureControlsTabFragment$key } from "#/__generated__/core/MeasureControlsTabFragment.graphql";
import { LinkedControlsCard } from "#/components/controls/LinkedControlsCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

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
          canCreateMeasureMapping: permission(
            action: "core:control:create-measure-mapping"
          )
          canDeleteMeasureMapping: permission(
            action: "core:control:delete-measure-mapping"
          )
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
  const controls = data.controls?.edges?.map(edge => edge.node) ?? [];

  const canLinkControl = controls.some(
    ({ canCreateMeasureMapping }) => canCreateMeasureMapping,
  );
  const canUnlinkControl = controls.some(
    ({ canDeleteMeasureMapping }) => canDeleteMeasureMapping,
  );
  const readOnly = !canLinkControl && !canUnlinkControl;

  const incrementOptions = {
    id: data.id,
    node: "controls(first:0)",
  };
  const [detachControl, isDetaching] = useMutationWithIncrement(
    detachControlMutation,
    {
      ...incrementOptions,
      value: -1,
    },
  );
  const [attachControl, isAttaching] = useMutationWithIncrement(
    attachControlMutation,
    {
      ...incrementOptions,
      value: 1,
    },
  );
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
      readOnly={readOnly}
    />
  );
}
