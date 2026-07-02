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
