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

import { graphql, useFragment } from "react-relay";
import { useOutletContext } from "react-router";

import type { RiskGraphNodeQuery$data } from "#/__generated__/core/RiskGraphNodeQuery.graphql";
import type { RiskObligationsTabFragment$key } from "#/__generated__/core/RiskObligationsTabFragment.graphql";
import { LinkedObligationsCard } from "#/components/obligations/LinkedObligationsCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const obligationsFragment = graphql`
  fragment RiskObligationsTabFragment on Risk {
    id
    obligations(first: 100) @connection(key: "Risk__obligations") {
      __id
      edges {
        node {
          id
          ...LinkedObligationsCardFragment
        }
      }
    }
  }
`;

const attachObligationMutation = graphql`
  mutation RiskObligationsTabCreateMutation(
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

export const detachObligationMutation = graphql`
  mutation RiskObligationsTabDetachMutation(
    $input: DeleteRiskObligationMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskObligationMapping(input: $input) {
      deletedObligationId @deleteEdge(connections: $connections)
    }
  }
`;

export default function RiskObligationsTab() {
  const { risk } = useOutletContext<{
    risk: RiskGraphNodeQuery$data["node"];
  }>();
  const data = useFragment<RiskObligationsTabFragment$key>(
    obligationsFragment,
    risk,
  );
  const connectionId = data.obligations.__id;
  const obligations = data.obligations?.edges?.map(edge => edge.node) ?? [];

  const canLinkObligation = risk.canCreateObligationMapping;
  const canUnlinkObligation = risk.canDeleteObligationMapping;
  const readOnly = !canLinkObligation && !canUnlinkObligation;

  const incrementOptions = {
    id: data.id,
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
      params={{ riskId: data.id }}
      connectionId={connectionId}
      variant="table"
      readOnly={readOnly}
    />
  );
}
