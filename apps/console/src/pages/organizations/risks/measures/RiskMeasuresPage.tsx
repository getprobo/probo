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

import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { RiskMeasuresPageQuery } from "#/__generated__/core/RiskMeasuresPageQuery.graphql";
import { LinkedMeasuresCard } from "#/components/measures/LinkedMeasuresCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const riskMeasuresPageQuery = graphql`
  query RiskMeasuresPageQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        id
        canCreateMeasureMapping: permission(
          action: "core:risk:create-measure-mapping"
        )
        canDeleteMeasureMapping: permission(
          action: "core:risk:delete-measure-mapping"
        )
        measures(first: 100) @connection(key: "RiskMeasuresPage_measures") {
          __id
          edges {
            node {
              id
              ...LinkedMeasuresCardFragment
            }
          }
        }
      }
    }
  }
`;

const attachMeasureMutation = graphql`
  mutation RiskMeasuresPageCreateMutation(
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

const detachMeasureMutation = graphql`
  mutation RiskMeasuresPageDetachMutation(
    $input: DeleteRiskMeasureMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskMeasureMapping(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

interface RiskMeasuresPageProps {
  queryRef: PreloadedQuery<RiskMeasuresPageQuery>;
}

export default function RiskMeasuresPage(props: RiskMeasuresPageProps) {
  const data = usePreloadedQuery<RiskMeasuresPageQuery>(riskMeasuresPageQuery, props.queryRef);
  if (data.node?.__typename !== "Risk") {
    throw new Error("Risk not found");
  }
  const risk = data.node;
  const connectionId = risk.measures.__id;
  const measures = risk.measures.edges.map(edge => edge.node);

  const readOnly = !risk.canCreateMeasureMapping && !risk.canDeleteMeasureMapping;

  const incrementOptions = {
    id: risk.id,
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
      params={{ riskId: risk.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
