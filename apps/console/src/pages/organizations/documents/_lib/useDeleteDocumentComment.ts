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

import type { useDeleteDocumentCommentMutation } from "#/__generated__/core/useDeleteDocumentCommentMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

// Deleting the last comment also deletes its thread on the server, so the
// thread edge leaves the review column in the same response.
const deleteDocumentCommentMutation = graphql`
  mutation useDeleteDocumentCommentMutation(
    $input: DeleteDocumentCommentInput!
    $commentConnections: [ID!]!
    $threadConnections: [ID!]!
  ) {
    deleteDocumentComment(input: $input) {
      deletedDocumentCommentId @deleteEdge(connections: $commentConnections)
      deletedDocumentCommentThreadId @deleteEdge(connections: $threadConnections)
    }
  }
`;

export function useDeleteDocumentComment() {
  const { t } = useTranslation("organizations/documents");
  const [commit, isDeleting] = useMutation<useDeleteDocumentCommentMutation>(
    deleteDocumentCommentMutation,
    { errorToast: t("details.review.errors.delete") },
  );

  async function deleteDocumentComment(
    commentId: string,
    commentsConnectionId: string,
    threadsConnectionId: string,
  ) {
    await commit({
      variables: {
        input: { commentId },
        commentConnections: [commentsConnectionId],
        threadConnections: [threadsConnectionId],
      },
    });
  }

  return [deleteDocumentComment, isDeleting] as const;
}
