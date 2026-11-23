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

import { useTranslation } from "react-i18next";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { useCreateDocumentCommentThreadMutation } from "#/__generated__/core/useCreateDocumentCommentThreadMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import type { DocumentCommentDraft } from "./documentCommentDraft";
import type { DocumentThreadConnections } from "./documentThreadConnections";

// The thread carries every field the review column reads, so its card
// appears without a refetch. It is appended to the open threads, which run
// oldest first, and inserted at the top of the history, which runs newest
// first.
const createDocumentCommentThreadMutation = graphql`
  mutation useCreateDocumentCommentThreadMutation(
    $input: CreateDocumentCommentThreadInput!
    $connections: [ID!]!
  ) {
    createDocumentCommentThread(input: $input) {
      documentCommentThreadEdge @appendEdge(connections: $connections) {
        cursor
        node {
          id
          anchorStart
          anchorEnd
          suggestion {
            id
            status
          }
          ...DocumentCommentThreadListItem_thread
          ...DocumentSuggestionListItem_thread
          ...DocumentCommentHistoryItem_thread
        }
      }
    }
  }
`;

export function useCreateDocumentCommentThread() {
  const { t } = useTranslation("organizations/documents");
  const [commit, isCreating] = useMutation<useCreateDocumentCommentThreadMutation>(
    createDocumentCommentThreadMutation,
    { errorToast: t("details.review.errors.comment") },
  );

  async function createDocumentCommentThread(
    documentId: string,
    draft: DocumentCommentDraft,
    content: string,
    connections: DocumentThreadConnections,
  ) {
    const response = await commit({
      variables: {
        input: {
          documentId,
          content,
          quote: draft.quote,
          anchorStart: draft.anchorStart,
          anchorEnd: draft.anchorEnd,
        },
        connections: [connections.open],
      },
      updater: (store) => {
        const history = store.get(connections.history);
        const edge = store
          .getRootField("createDocumentCommentThread")
          ?.getLinkedRecord("documentCommentThreadEdge");
        const thread = edge?.getLinkedRecord("node");
        if (history == null || thread == null) return;
        const historyEdge = ConnectionHandler.createEdge(store, history, thread, "DocumentCommentThreadEdge");
        ConnectionHandler.insertEdgeBefore(history, historyEdge);
      },
    });
    return response.createDocumentCommentThread.documentCommentThreadEdge.node.id;
  }

  return [createDocumentCommentThread, isCreating] as const;
}
