// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Suspense } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { DocumentCommentHistoryQuery } from "#/__generated__/core/DocumentCommentHistoryQuery.graphql";

import type { DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import type { SuggestionAnchor } from "../../_lib/reviewAnchors";
import type { ReviewHistoryTarget } from "../../_lib/reviewHistoryTarget";

import { DocumentCommentHistoryList } from "./DocumentCommentHistoryList";

export const documentCommentHistoryQuery = graphql`
  query DocumentCommentHistoryQuery($documentId: ID!) {
    node(id: $documentId) {
      __typename
      ... on Document {
        ...DocumentCommentHistoryList_document
      }
    }
  }
`;

interface DocumentCommentHistoryProps {
  queryRef: PreloadedQuery<DocumentCommentHistoryQuery> | null | undefined;
  threadConnections: DocumentThreadConnections;
  liveSuggestions: ReadonlyMap<string, SuggestionAnchor>;
  activeKey: string | null;
  onActivate: (target: ReviewHistoryTarget) => void;
  onPressInside: () => void;
}

// The history never blocks the document. While its query loads or after it
// fails, the text stays usable and the list is empty.
export function DocumentCommentHistory({ queryRef, ...props }: DocumentCommentHistoryProps) {
  if (queryRef == null) {
    return null;
  }

  return (
    <ErrorBoundary fallback={null}>
      <Suspense fallback={null}>
        <DocumentCommentHistoryLoaded queryRef={queryRef} {...props} />
      </Suspense>
    </ErrorBoundary>
  );
}

function DocumentCommentHistoryLoaded({
  queryRef,
  ...props
}: Omit<DocumentCommentHistoryProps, "queryRef"> & {
  queryRef: PreloadedQuery<DocumentCommentHistoryQuery>;
}) {
  const data = usePreloadedQuery<DocumentCommentHistoryQuery>(documentCommentHistoryQuery, queryRef);

  if (data.node.__typename !== "Document") {
    return null;
  }

  return <DocumentCommentHistoryList documentKey={data.node} {...props} />;
}
