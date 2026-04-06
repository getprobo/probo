// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
