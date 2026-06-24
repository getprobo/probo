// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { ComponentProps } from "react";
import { graphql, useRefetchableFragment } from "react-relay";
import { useOutletContext, useParams } from "react-router";

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
    canCreateControlMeasureMapping: permission(
      action: "core:control:create-measure-mapping"
    )
    canDeleteControlMeasureMapping: permission(
      action: "core:control:delete-measure-mapping"
    )
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
    measure: MeasureControlsTabFragment$key;
  }>();
  const { measureId } = useParams<{ measureId: string }>();
  if (!measureId) {
    throw new Error("Missing :measureId param in route");
  }
  // eslint-disable-next-line relay/generated-typescript-types
  const [data, refetch] = useRefetchableFragment(controlsFragment, measure);
  const connectionId = data.controls.__id;
  const controls = data.controls?.edges?.map(edge => edge.node) ?? [];

  const canLinkControl = data.canCreateControlMeasureMapping;
  const canUnlinkControl = data.canDeleteControlMeasureMapping;
  const readOnly = !canLinkControl && !canUnlinkControl;

  const incrementOptions = {
    id: measureId,
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
      controls={controls as ComponentProps<typeof LinkedControlsCard>["controls"]}
      onDetach={detachControl}
      onAttach={attachControl}
      params={{ measureId }}
      connectionId={connectionId}
      refetch={refetch}
      readOnly={readOnly}
    />
  );
}
