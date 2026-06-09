// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  const data = usePreloadedQuery(riskObligationsPageQuery, props.queryRef);
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
