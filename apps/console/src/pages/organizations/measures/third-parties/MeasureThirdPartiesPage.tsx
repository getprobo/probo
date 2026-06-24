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

import type { MeasureThirdPartiesPageFragment$key } from "#/__generated__/core/MeasureThirdPartiesPageFragment.graphql";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

import { LinkedThirdPartiesCard } from "../_components/LinkedThirdPartiesCard";

export const thirdPartiesFragment = graphql`
  fragment MeasureThirdPartiesPageFragment on Measure {
    id
    canCreateMeasureThirdPartyMapping: permission(
      action: "core:measure:create-third-party-mapping"
    )
    canDeleteMeasureThirdPartyMapping: permission(
      action: "core:measure:delete-third-party-mapping"
    )
    thirdParties(first: 100) @connection(key: "MeasureThirdPartiesPage_thirdParties") {
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
  mutation MeasureThirdPartiesPageAttachMutation(
    $input: CreateMeasureThirdPartyMappingInput!
    $connections: [ID!]!
  ) {
    createMeasureThirdPartyMapping(input: $input) {
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
  mutation MeasureThirdPartiesPageDetachMutation(
    $input: DeleteMeasureThirdPartyMappingInput!
    $connections: [ID!]!
  ) {
    deleteMeasureThirdPartyMapping(input: $input) {
      deletedThirdPartyId @deleteEdge(connections: $connections)
    }
  }
`;

export default function MeasureThirdPartiesPage() {
  const { measureId } = useParams<{ measureId: string }>();
  if (!measureId) {
    throw new Error("Missing :measureId param in route");
  }
  const { measure } = useOutletContext<{
    measure: MeasureThirdPartiesPageFragment$key;
  }>();
  const data = useFragment(thirdPartiesFragment, measure);
  const connectionId = data.thirdParties.__id;
  const thirdParties = data.thirdParties?.edges?.map(edge => edge.node) ?? [];

  const canLink = data.canCreateMeasureThirdPartyMapping;
  const canUnlink = data.canDeleteMeasureThirdPartyMapping;
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
      params={{ measureId: data.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
