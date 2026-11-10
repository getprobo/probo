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

import { useEffect, useState } from "react";
import { useRelayEnvironment } from "react-relay";
import { fetchQuery, graphql } from "relay-runtime";

import type { useResolvedCommentThreadsQuery } from "#/__generated__/core/useResolvedCommentThreadsQuery.graphql";

export type ResolvedCommentThread = {
  id: string;
  anchorQuote: string;
  comments: {
    edges: ReadonlyArray<{ node: { content: string } }>;
  };
};

const historyQuery = graphql`
  query useResolvedCommentThreadsQuery($documentId: ID!) {
    document: node(id: $documentId) {
      __typename
      ... on Document {
        commentThreads(first: 50, filter: { resolved: true }) {
          edges {
            node {
              id
              anchorQuote
              comments(first: 1) {
                edges { node { content } }
              }
            }
          }
        }
      }
    }
  }
`;

export function useResolvedCommentThreads({
  documentId,
  historyOpen,
  reviewVersion,
}: {
  documentId: string;
  historyOpen: boolean;
  reviewVersion: number;
}) {
  const environment = useRelayEnvironment();
  const [resolvedThreads, setResolvedThreads] = useState<ResolvedCommentThread[]>([]);
  const [invalidNode, setInvalidNode] = useState(false);

  useEffect(() => {
    if (!historyOpen) return;
    setInvalidNode(false);
    const subscription = fetchQuery<useResolvedCommentThreadsQuery>(
      environment,
      historyQuery,
      { documentId: documentId },
      { fetchPolicy: "network-only" },
    ).subscribe({
      next: (data) => {
        if (data.document?.__typename !== "Document") {
          setInvalidNode(true);
          return;
        }
        setResolvedThreads(data.document.commentThreads.edges.map(edge => edge.node));
      },
    });
    return () => subscription.unsubscribe();
  }, [environment, historyOpen, documentId, reviewVersion]);

  if (invalidNode) {
    throw new Error("invalid type for node");
  }

  return {
    resolvedThreads,
    setResolvedThreads,
  };
}
