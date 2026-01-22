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
