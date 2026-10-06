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

import type { InternalControlRisksTabFragment$key } from "#/__generated__/core/InternalControlRisksTabFragment.graphql";
import { LinkedRisksCard } from "#/components/risks/LinkedRisksCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const risksFragment = graphql`
  fragment InternalControlRisksTabFragment on InternalControl {
    id
    canCreateRiskInternalControlMapping: permission(
      action: "risk-management:risk:create-internal-control-mapping"
    )
    canDeleteRiskInternalControlMapping: permission(
      action: "risk-management:risk:delete-internal-control-mapping"
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
  mutation InternalControlRisksTabCreateMutation(
    $input: CreateRiskInternalControlMappingInput!
    $connections: [ID!]!
  ) {
    createRiskInternalControlMapping(input: $input) {
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
  mutation InternalControlRisksTabDetachMutation(
    $input: DeleteRiskInternalControlMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskInternalControlMapping(input: $input) {
      deletedRiskId @deleteEdge(connections: $connections)
    }
  }
`;

export default function InternalControlRisksTab() {
  const { internalControlId } = useParams<{ internalControlId: string }>();
  if (!internalControlId) {
    throw new Error("Missing :internalControlId param in route");
  }
  const { internalControl } = useOutletContext<{
    internalControl: InternalControlRisksTabFragment$key;
  }>();
  const data = useFragment(risksFragment, internalControl);
  const connectionId = data.risks.__id;
  const risks = data.risks?.edges?.map(edge => edge.node) ?? [];

  const canLinkRisk = data.canCreateRiskInternalControlMapping;
  const canUnlinkRisk = data.canDeleteRiskInternalControlMapping;
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
      params={{ internalControlId: data.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
