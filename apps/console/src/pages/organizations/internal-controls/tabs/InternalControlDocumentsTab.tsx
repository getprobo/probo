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

import type { InternalControlDocumentsTabFragment$key } from "#/__generated__/core/InternalControlDocumentsTabFragment.graphql";
import { LinkedDocumentsCard } from "#/components/documents/LinkedDocumentsCard";
import { useMutationWithIncrement } from "#/hooks/useMutationWithIncrement";

export const documentsFragment = graphql`
  fragment InternalControlDocumentsTabFragment on InternalControl {
    id
    canCreateDocumentMapping: permission(
      action: "core:internal-control:create-document-mapping"
    )
    canDeleteDocumentMapping: permission(
      action: "core:internal-control:delete-document-mapping"
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
  mutation InternalControlDocumentsTabCreateMutation(
    $input: CreateInternalControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    createInternalControlDocumentMapping(input: $input) {
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
  mutation InternalControlDocumentsTabDetachMutation(
    $input: DeleteInternalControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    deleteInternalControlDocumentMapping(input: $input) {
      deletedDocumentId @deleteEdge(connections: $connections)
    }
  }
`;

export default function InternalControlDocumentsTab() {
  const { internalControlId } = useParams<{ internalControlId: string }>();
  if (!internalControlId) {
    throw new Error("Missing :internalControlId param in route");
  }
  const { internalControl } = useOutletContext<{
    internalControl: InternalControlDocumentsTabFragment$key;
  }>();
  const data = useFragment<InternalControlDocumentsTabFragment$key>(
    documentsFragment,
    internalControl,
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
      params={{ internalControlId }}
      connectionId={connectionId}
      readOnly={readOnly}
    />
  );
}
