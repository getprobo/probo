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

import type { RiskObligationsPageQuery } from "#/__generated__/core/RiskObligationsPageQuery.graphql";
import { LinkedObligationsCard } from "#/components/obligations/LinkedObligationsCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const riskObligationsPageQuery = graphql`
  query RiskObligationsPageQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        id
        canCreateObligationMapping: permission(
          action: "core:risk:create-obligation-mapping"
        )
        canDeleteObligationMapping: permission(
          action: "core:risk:delete-obligation-mapping"
        )
        obligations(first: 100) @connection(key: "RiskObligationsPage_obligations") {
          __id
          edges {
            node {
              id
              ...LinkedObligationsCardFragment
            }
          }
        }
      }
    }
  }
`;

const attachObligationMutation = graphql`
  mutation RiskObligationsPageCreateMutation(
    $input: CreateRiskObligationMappingInput!
    $connections: [ID!]!
  ) {
    createRiskObligationMapping(input: $input) {
      obligationEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedObligationsCardFragment
        }
      }
    }
  }
`;

const detachObligationMutation = graphql`
  mutation RiskObligationsPageDetachMutation(
    $input: DeleteRiskObligationMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskObligationMapping(input: $input) {
      deletedObligationId @deleteEdge(connections: $connections)
    }
  }
`;

interface RiskObligationsPageProps {
  queryRef: PreloadedQuery<RiskObligationsPageQuery>;
}

export default function RiskObligationsPage(props: RiskObligationsPageProps) {
  const data = usePreloadedQuery<RiskObligationsPageQuery>(riskObligationsPageQuery, props.queryRef);
  if (data.node?.__typename !== "Risk") {
    throw new Error("Risk not found");
  }
  const risk = data.node;
  const connectionId = risk.obligations.__id;
  const obligations = risk.obligations.edges.map(edge => edge.node);

  const readOnly
    = !risk.canCreateObligationMapping && !risk.canDeleteObligationMapping;

  const incrementOptions = {
    id: risk.id,
    node: "obligations(first:0)",
  };
  const [detachObligation, isDetaching] = useMutationWithIncrement(
    detachObligationMutation,
    {
      ...incrementOptions,
      value: -1,
    },
  );
  const [attachObligation, isAttaching] = useMutationWithIncrement(
    attachObligationMutation,
    {
      ...incrementOptions,
      value: 1,
    },
  );
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedObligationsCard
      disabled={isLoading}
      obligations={obligations}
      onAttach={attachObligation}
      onDetach={detachObligation}
      params={{ riskId: risk.id }}
      connectionId={connectionId}
      variant="table"
      readOnly={readOnly}
    />
  );
}
