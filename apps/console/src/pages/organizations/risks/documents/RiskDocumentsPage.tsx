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

import type { RiskDocumentsPageQuery } from "#/__generated__/core/RiskDocumentsPageQuery.graphql";
import { LinkedDocumentsCard } from "#/components/documents/LinkedDocumentsCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const riskDocumentsPageQuery = graphql`
  query RiskDocumentsPageQuery($riskId: ID!) {
    node(id: $riskId) {
      __typename
      ... on Risk {
        id
        canCreateDocumentMapping: permission(
          action: "core:risk:create-document-mapping"
        )
        canDeleteDocumentMapping: permission(
          action: "core:risk:delete-document-mapping"
        )
        documents(first: 100) @connection(key: "RiskDocumentsPage_documents") {
          __id
          edges {
            node {
              id
              ...LinkedDocumentsCardFragment
            }
          }
        }
      }
    }
  }
`;

const attachDocumentMutation = graphql`
  mutation RiskDocumentsPageCreateMutation(
    $input: CreateRiskDocumentMappingInput!
    $connections: [ID!]!
  ) {
    createRiskDocumentMapping(input: $input) {
      documentEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedDocumentsCardFragment
        }
      }
    }
  }
`;

const detachDocumentMutation = graphql`
  mutation RiskDocumentsPageDetachMutation(
    $input: DeleteRiskDocumentMappingInput!
    $connections: [ID!]!
  ) {
    deleteRiskDocumentMapping(input: $input) {
      deletedDocumentId @deleteEdge(connections: $connections)
    }
  }
`;

interface RiskDocumentsPageProps {
  queryRef: PreloadedQuery<RiskDocumentsPageQuery>;
}

export default function RiskDocumentsPage(props: RiskDocumentsPageProps) {
  const data = usePreloadedQuery<RiskDocumentsPageQuery>(riskDocumentsPageQuery, props.queryRef);
  if (data.node?.__typename !== "Risk") {
    throw new Error("Risk not found");
  }
  const risk = data.node;
  const connectionId = risk.documents.__id;
  const documents = risk.documents.edges.map(edge => edge.node);

  const readOnly = !risk.canCreateDocumentMapping && !risk.canDeleteDocumentMapping;

  const incrementOptions = {
    id: risk.id,
    node: "documents(first:0)",
  };
  const [detachDocument, isDetaching] = useMutationWithIncrement(
    detachDocumentMutation,
    {
      ...incrementOptions,
      value: -1,
    },
  );
  const [attachDocument, isAttaching] = useMutationWithIncrement(
    attachDocumentMutation,
    {
      ...incrementOptions,
      value: 1,
    },
  );
  const isLoading = isDetaching || isAttaching;

  return (
    <LinkedDocumentsCard
      disabled={isLoading}
      documents={documents}
      onAttach={attachDocument}
      onDetach={detachDocument}
      params={{ riskId: risk.id }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
