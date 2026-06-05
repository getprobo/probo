// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
  const data = usePreloadedQuery(riskMeasuresPageQuery, props.queryRef);
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
