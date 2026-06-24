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

import type { MeasureDocumentsTabFragment$key } from "#/__generated__/core/MeasureDocumentsTabFragment.graphql";
import { LinkedDocumentsCard } from "#/components/documents/LinkedDocumentsCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const documentsFragment = graphql`
  fragment MeasureDocumentsTabFragment on Measure {
    id
    canCreateDocumentMapping: permission(
      action: "core:measure:create-document-mapping"
    )
    canDeleteDocumentMapping: permission(
      action: "core:measure:delete-document-mapping"
    )
    documents(first: 100) @connection(key: "Measure__documents") {
      __id
      edges {
        node {
          id
          ...LinkedDocumentsCardFragment
        }
      }
    }
  }
`;

const attachDocumentMutation = graphql`
  mutation MeasureDocumentsTabCreateMutation(
    $input: CreateMeasureDocumentMappingInput!
    $connections: [ID!]!
  ) {
    createMeasureDocumentMapping(input: $input) {
      documentEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedDocumentsCardFragment
        }
      }
    }
  }
`;

export const detachDocumentMutation = graphql`
  mutation MeasureDocumentsTabDetachMutation(
    $input: DeleteMeasureDocumentMappingInput!
    $connections: [ID!]!
  ) {
    deleteMeasureDocumentMapping(input: $input) {
      deletedDocumentId @deleteEdge(connections: $connections)
    }
  }
`;

export default function MeasureDocumentsTab() {
  const { measureId } = useParams<{ measureId: string }>();
  if (!measureId) {
    throw new Error("Missing :measureId param in route");
  }
  const { measure } = useOutletContext<{
    measure: MeasureDocumentsTabFragment$key;
  }>();
  const data = useFragment<MeasureDocumentsTabFragment$key>(
    documentsFragment,
    measure,
  );
  const connectionId = data.documents.__id;
  const documents = data.documents?.edges?.map(edge => edge.node) ?? [];

  const canLinkDocument = data.canCreateDocumentMapping;
  const canUnlinkDocument = data.canDeleteDocumentMapping;
  const readOnly = !canLinkDocument && !canUnlinkDocument;

  const incrementOptions = {
    id: data.id,
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
      params={{ measureId }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
