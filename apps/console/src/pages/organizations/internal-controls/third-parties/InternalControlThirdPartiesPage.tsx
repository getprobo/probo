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

import type { InternalControlThirdPartiesPageFragment$key } from "#/__generated__/core/InternalControlThirdPartiesPageFragment.graphql";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

import { LinkedThirdPartiesCard } from "../_components/LinkedThirdPartiesCard";

export const thirdPartiesFragment = graphql`
  fragment InternalControlThirdPartiesPageFragment on InternalControl {
    id
    canCreateInternalControlThirdPartyMapping: permission(
      action: "core:internal-control:create-third-party-mapping"
    )
    canDeleteInternalControlThirdPartyMapping: permission(
      action: "core:internal-control:delete-third-party-mapping"
    )
    thirdParties(first: 100) @connection(key: "InternalControlThirdPartiesPage_thirdParties") {
      __id
      edges {
        node {
          id
          ...LinkedThirdPartiesCardFragment
        }
      }
    }
  }
`;

const attachThirdPartyMutation = graphql`
  mutation InternalControlThirdPartiesPageAttachMutation(
    $input: CreateInternalControlThirdPartyMappingInput!
    $connections: [ID!]!
  ) {
    createInternalControlThirdPartyMapping(input: $input) {
      thirdPartyEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedThirdPartiesCardFragment
        }
      }
    }
  }
`;

const detachThirdPartyMutation = graphql`
  mutation InternalControlThirdPartiesPageDetachMutation(
    $input: DeleteInternalControlThirdPartyMappingInput!
    $connections: [ID!]!
  ) {
    deleteInternalControlThirdPartyMapping(input: $input) {
      deletedThirdPartyId @deleteEdge(connections: $connections)
    }
  }
`;

export default function InternalControlThirdPartiesPage() {
  const { internalControlId } = useParams<{ internalControlId: string }>();
  if (!internalControlId) {
    throw new Error("Missing :internalControlId param in route");
  }
  const { internalControl } = useOutletContext<{
    internalControl: InternalControlThirdPartiesPageFragment$key;
  }>();
  const data = useFragment(thirdPartiesFragment, internalControl);
  const connectionId = data.thirdParties.__id;
  const thirdParties = data.thirdParties?.edges?.map(edge => edge.node) ?? [];

  const canLink = data.canCreateInternalControlThirdPartyMapping;
  const canUnlink = data.canDeleteInternalControlThirdPartyMapping;
  const readOnly = !canLink && !canUnlink;

  const incrementOptions = {
    id: data.id,
    node: "thirdParties(first:0)",
  };
  const [detachThirdParty, isDetaching] = useMutationWithIncrement(
    detachThirdPartyMutation,
    {
      ...incrementOptions,
      value: -1,
    },
  );
  const [attachThirdParty, isAttaching] = useMutationWithIncrement(
    attachThirdPartyMutation,
    {
      ...incrementOptions,
      value: 1,
    },
  );
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedThirdPartiesCard
      disabled={isLoading}
      thirdParties={thirdParties}
      onAttach={attachThirdParty}
      onDetach={detachThirdParty}
      params={{ internalControlId: data.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
