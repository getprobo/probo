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

import type { RiskInternalControlsPageQuery } from "#/__generated__/core/RiskInternalControlsPageQuery.graphql";
import { LinkedInternalControlsCard } from "#/components/internal-controls/LinkedInternalControlsCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const riskInternalControlsPageQuery = graphql`
  query RiskInternalControlsPageQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        id
        canCreateInternalControlMapping: permission(
          action: "risk-management:risk:create-internal-control-mapping"
        )
        canDeleteInternalControlMapping: permission(
          action: "risk-management:risk:delete-internal-control-mapping"
        )
        internalControls(first: 100) @connection(key: "RiskInternalControlsPage_internalControls") {
          __id
          edges {
            node {
              id
              ...LinkedInternalControlsCardFragment
            }
          }
        }
      }
    }
  }
`;

const attachInternalControlMutation = graphql`
  mutation RiskInternalControlsPageCreateMutation(
    $input: CreateRiskInternalControlMappingInput!
    $connections: [ID!]!
  ) {
    createRiskInternalControlMapping(input: $input) {
      internalControlEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedInternalControlsCardFragment
        }
      }
    }
  }
`;

const detachInternalControlMutation = graphql`
  mutation RiskInternalControlsPageDetachMutation(
    $input: DeleteRiskInternalControlMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskInternalControlMapping(input: $input) {
      deletedInternalControlId @deleteEdge(connections: $connections)
    }
  }
`;

interface RiskInternalControlsPageProps {
  queryRef: PreloadedQuery<RiskInternalControlsPageQuery>;
}

export default function RiskInternalControlsPage(props: RiskInternalControlsPageProps) {
  const data = usePreloadedQuery<RiskInternalControlsPageQuery>(riskInternalControlsPageQuery, props.queryRef);
  if (data.node?.__typename !== "Risk") {
    throw new Error("Risk not found");
  }
  const risk = data.node;
  const connectionId = risk.internalControls.__id;
  const internalControls = risk.internalControls.edges.map(edge => edge.node);

  const readOnly = !risk.canCreateInternalControlMapping && !risk.canDeleteInternalControlMapping;

  const incrementOptions = {
    id: risk.id,
    node: "internalControls(first:0)",
  };
  const [detachInternalControl, isDetaching] = useMutationWithIncrement(
    detachInternalControlMutation,
    {
      ...incrementOptions,
      value: -1,
    },
  );
  const [attachInternalControl, isAttaching] = useMutationWithIncrement(
    attachInternalControlMutation,
    {
      ...incrementOptions,
      value: 1,
    },
  );
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedInternalControlsCard
      disabled={isLoading}
      internalControls={internalControls}
      onAttach={attachInternalControl}
      onDetach={detachInternalControl}
      params={{ riskId: risk.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
