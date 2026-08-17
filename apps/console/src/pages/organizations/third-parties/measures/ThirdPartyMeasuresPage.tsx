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

import type { ThirdPartyMeasuresPageAttachMutation } from "#/__generated__/core/ThirdPartyMeasuresPageAttachMutation.graphql";
import type { ThirdPartyMeasuresPageDetachMutation } from "#/__generated__/core/ThirdPartyMeasuresPageDetachMutation.graphql";
import type { ThirdPartyMeasuresPageQuery } from "#/__generated__/core/ThirdPartyMeasuresPageQuery.graphql";
import { LinkedMeasuresCard } from "#/components/measures/LinkedMeasuresCard";
import { useMutation } from "#/lib/relay/useMutation";

export const thirdPartyMeasuresPageQuery = graphql`
  query ThirdPartyMeasuresPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
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

interface ThirdPartyMeasuresPageProps {
  queryRef: PreloadedQuery<ThirdPartyMeasuresPageQuery>;
}

export default function ThirdPartyMeasuresPage(props: ThirdPartyMeasuresPageProps) {
  const data = usePreloadedQuery<ThirdPartyMeasuresPageQuery>(thirdPartyMeasuresPageQuery, props.queryRef);
  if (data.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }
  const thirdParty = data.node;

  const connectionId = thirdParty.measures.__id;
  const measures = thirdParty.measures.edges.map(edge => edge.node);

  const canLink = thirdParty.canCreateMeasureThirdPartyMapping;
  const canUnlink = thirdParty.canDeleteMeasureThirdPartyMapping;
  const readOnly = !canLink && !canUnlink;

  const [detachMeasure, isDetaching] = useMutation<ThirdPartyMeasuresPageDetachMutation>(
    detachMeasureMutation,
  );
  const [attachMeasure, isAttaching] = useMutation<ThirdPartyMeasuresPageAttachMutation>(
    attachMeasureMutation,
  );
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedMeasuresCard
      disabled={isLoading}
      measures={measures}
      onAttach={(args) => {
        void attachMeasure(args).catch(() => undefined);
      }}
      onDetach={(args) => {
        void detachMeasure(args).catch(() => undefined);
      }}
      params={{ thirdPartyId: thirdParty.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
