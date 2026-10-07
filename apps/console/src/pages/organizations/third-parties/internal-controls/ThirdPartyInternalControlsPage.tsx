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

import type { ThirdPartyInternalControlsPageAttachMutation } from "#/__generated__/core/ThirdPartyInternalControlsPageAttachMutation.graphql";
import type { ThirdPartyInternalControlsPageDetachMutation } from "#/__generated__/core/ThirdPartyInternalControlsPageDetachMutation.graphql";
import type { ThirdPartyInternalControlsPageQuery } from "#/__generated__/core/ThirdPartyInternalControlsPageQuery.graphql";
import { LinkedInternalControlsCard } from "#/components/internal-controls/LinkedInternalControlsCard";
import { useMutation } from "#/lib/relay/useMutation";

export const thirdPartyInternalControlsPageQuery = graphql`
  query ThirdPartyInternalControlsPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        id
        canCreateInternalControlThirdPartyMapping: permission(
          action: "core:internal-control:create-third-party-mapping"
        )
        canDeleteInternalControlThirdPartyMapping: permission(
          action: "core:internal-control:delete-third-party-mapping"
        )
        internalControls(first: 100) @connection(key: "ThirdPartyInternalControlsPage_internalControls") {
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
  mutation ThirdPartyInternalControlsPageAttachMutation(
    $input: CreateInternalControlThirdPartyMappingInput!
    $connections: [ID!]!
  ) {
    createInternalControlThirdPartyMapping(input: $input) {
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
  mutation ThirdPartyInternalControlsPageDetachMutation(
    $input: DeleteInternalControlThirdPartyMappingInput!
    $connections: [ID!]!
  ) {
    deleteInternalControlThirdPartyMapping(input: $input) {
      deletedInternalControlId @deleteEdge(connections: $connections)
    }
  }
`;

interface ThirdPartyInternalControlsPageProps {
  queryRef: PreloadedQuery<ThirdPartyInternalControlsPageQuery>;
}

export default function ThirdPartyInternalControlsPage(props: ThirdPartyInternalControlsPageProps) {
  const data = usePreloadedQuery<ThirdPartyInternalControlsPageQuery>(
    thirdPartyInternalControlsPageQuery,
    props.queryRef,
  );
  if (data.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }
  const thirdParty = data.node;

  const connectionId = thirdParty.internalControls.__id;
  const internalControls = thirdParty.internalControls.edges.map(edge => edge.node);

  const canLink = thirdParty.canCreateInternalControlThirdPartyMapping;
  const canUnlink = thirdParty.canDeleteInternalControlThirdPartyMapping;
  const readOnly = !canLink && !canUnlink;

  const [detachInternalControl, isDetaching] = useMutation<ThirdPartyInternalControlsPageDetachMutation>(
    detachInternalControlMutation,
  );
  const [attachInternalControl, isAttaching] = useMutation<ThirdPartyInternalControlsPageAttachMutation>(
    attachInternalControlMutation,
  );
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedInternalControlsCard
      disabled={isLoading}
      internalControls={internalControls}
      onAttach={(args) => {
        void attachInternalControl(args).catch(() => undefined);
      }}
      onDetach={(args) => {
        void detachInternalControl(args).catch(() => undefined);
      }}
      params={{ thirdPartyId: thirdParty.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
