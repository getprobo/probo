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

import type { RepoCollaborationHandle } from "@probo/collaboration";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import type { Editor } from "@tiptap/core";
import { Suspense } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { DocumentReviewListQuery } from "#/__generated__/core/DocumentReviewListQuery.graphql";

import { DocumentReviewPanel } from "./DocumentReviewPanel";

export const documentReviewListQuery = graphql`
  query DocumentReviewListQuery($documentId: ID!) {
    node(id: $documentId) {
      __typename
      ... on Document {
        ...DocumentReviewPanel_threads
        ...DocumentReviewPanel_suggestions
      }
    }
  }
`;

interface DocumentReviewListProps {
  queryRef: PreloadedQuery<DocumentReviewListQuery> | null;
  editor: Editor | null;
  collaborationHandle: RepoCollaborationHandle;
  // The view switch is portaled here so it sits outside the scroll viewport.
  switchContainer: HTMLElement | null;
  activeKey: string | null;
  onActiveChange: (key: string | null) => void;
}

// The column never blocks the document. While its query loads or after it
// fails, the text stays usable and the column is empty.
export function DocumentReviewList({ queryRef, ...props }: DocumentReviewListProps) {
  if (queryRef == null) {
    return null;
  }

  return (
    <ErrorBoundary fallback={null}>
      <Suspense fallback={null}>
        <DocumentReviewListLoaded queryRef={queryRef} {...props} />
      </Suspense>
    </ErrorBoundary>
  );
}

function DocumentReviewListLoaded({
  queryRef,
  ...props
}: Omit<DocumentReviewListProps, "queryRef"> & {
  queryRef: PreloadedQuery<DocumentReviewListQuery>;
}) {
  const data = usePreloadedQuery<DocumentReviewListQuery>(documentReviewListQuery, queryRef);

  if (data.node.__typename !== "Document") {
    return null;
  }

  return <DocumentReviewPanel documentKey={data.node} {...props} />;
}
