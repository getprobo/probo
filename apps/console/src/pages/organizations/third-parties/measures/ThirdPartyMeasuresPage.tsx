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

import type { ThirdPartyMeasuresPageFragment$key } from "#/__generated__/core/ThirdPartyMeasuresPageFragment.graphql";
import { LinkedMeasuresCard } from "#/components/measures/LinkedMeasuresCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const measuresFragment = graphql`
  fragment ThirdPartyMeasuresPageFragment on ThirdParty {
    id
    canCreateMeasureThirdPartyMapping: permission(
      action: "core:measure:create-third-party-mapping"
    )
    canDeleteMeasureThirdPartyMapping: permission(
      action: "core:measure:delete-third-party-mapping"
    )
    measures(first: 100) @connection(key: "ThirdPartyMeasuresPage_measures") {
      __id
      edges {
        node {
          id
          ...LinkedMeasuresCardFragment
        }
      }
    }
  }
`;

const attachMeasureMutation = graphql`
  mutation ThirdPartyMeasuresPageAttachMutation(
    $input: CreateMeasureThirdPartyMappingInput!
    $connections: [ID!]!
  ) {
    createMeasureThirdPartyMapping(input: $input) {
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
  mutation ThirdPartyMeasuresPageDetachMutation(
    $input: DeleteMeasureThirdPartyMappingInput!
    $connections: [ID!]!
  ) {
    deleteMeasureThirdPartyMapping(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

export default function ThirdPartyMeasuresPage() {
  const { thirdPartyId } = useParams<{ thirdPartyId: string }>();
  if (!thirdPartyId) {
    throw new Error("Missing :thirdPartyId param in route");
  }
  const { thirdParty } = useOutletContext<{
    thirdParty: ThirdPartyMeasuresPageFragment$key;
  }>();
  const data = useFragment(measuresFragment, thirdParty);
  const connectionId = data.measures.__id;
  const measures = data.measures?.edges?.map(edge => edge.node) ?? [];

  const canLink = data.canCreateMeasureThirdPartyMapping;
  const canUnlink = data.canDeleteMeasureThirdPartyMapping;
  const readOnly = !canLink && !canUnlink;

  const incrementOptions = {
    id: data.id,
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
      params={{ thirdPartyId: data.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
