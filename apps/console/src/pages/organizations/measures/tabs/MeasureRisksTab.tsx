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

import { graphql, useFragment } from "react-relay";
import { useOutletContext, useParams } from "react-router";

import type { MeasureRisksTabFragment$key } from "#/__generated__/core/MeasureRisksTabFragment.graphql";
import { LinkedRisksCard } from "#/components/risks/LinkedRisksCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const risksFragment = graphql`
  fragment MeasureRisksTabFragment on Measure {
    id
    canCreateRiskMeasureMapping: permission(
      action: "core:risk:create-measure-mapping"
    )
    canDeleteRiskMeasureMapping: permission(
      action: "core:risk:delete-measure-mapping"
    )
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
  const { measureId } = useParams<{ measureId: string }>();
  if (!measureId) {
    throw new Error("Missing :measureId param in route");
  }
  const { measure } = useOutletContext<{
    measure: MeasureRisksTabFragment$key;
  }>();
  const data = useFragment(risksFragment, measure);
  const connectionId = data.risks.__id;
  const risks = data.risks?.edges?.map(edge => edge.node) ?? [];

  const canLinkRisk = data.canCreateRiskMeasureMapping;
  const canUnlinkRisk = data.canDeleteRiskMeasureMapping;
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
