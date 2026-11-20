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
import { graphql } from "react-relay";
import { ConnectionHandler } from "relay-runtime";

import type { useCreateDocumentCommentThreadMutation } from "#/__generated__/core/useCreateDocumentCommentThreadMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import type { DocumentThreadConnections } from "./documentThreadConnections";

// The new edge carries every field the review column reads from a thread, so
// it renders in place without a refetch. It is appended to the open threads,
// which run oldest first, and inserted at the top of the history, which runs
// newest first.
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
          suggestionId
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
    { errorToast: t("details.review.errors.reply") },
  );

  async function createSuggestionThread(
    documentId: string,
    suggestionId: string,
    content: string,
    threadConnections: DocumentThreadConnections,
  ) {
    await commit({
      variables: {
        input: { documentId, suggestionId, content },
        connections: [threadConnections.open],
      },
      updater: (store) => {
        const history = store.get(threadConnections.history);
        const edge = store
          .getRootField("createDocumentCommentThread")
          ?.getLinkedRecord("documentCommentThreadEdge");
        if (history == null || edge == null) return;
        const historyEdge = ConnectionHandler.buildConnectionEdge(store, history, edge);
        if (historyEdge != null) {
          ConnectionHandler.insertEdgeBefore(history, historyEdge);
        }
      },
    });
  }

  return [createSuggestionThread, isCreating] as const;
}
