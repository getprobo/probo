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

import { graphql } from "relay-runtime";

export const createThreadMutation = graphql`
  mutation DocumentReviewMarginMutations_createThreadMutation($input: CreateDocumentCommentThreadInput!) {
    createDocumentCommentThread(input: $input) {
      documentCommentThreadEdge {
        cursor
        node { ...DocumentReviewMarginThread_thread }
      }
    }
  }
`;

export const replyMutation = graphql`
  mutation DocumentReviewMarginMutations_replyMutation($input: CreateDocumentCommentInput!) {
    createDocumentComment(input: $input) {
      documentCommentEdge {
        cursor
        node { id content createdAt author { id fullName } }
      }
    }
  }
`;

export const updateCommentMutation = graphql`
  mutation DocumentReviewMarginMutations_updateCommentMutation($input: UpdateDocumentCommentInput!) {
    updateDocumentComment(input: $input) {
      documentComment { id content }
    }
  }
`;

export const deleteCommentMutation = graphql`
  mutation DocumentReviewMarginMutations_deleteCommentMutation($input: DeleteDocumentCommentInput!) {
    deleteDocumentComment(input: $input) {
      deletedDocumentCommentId
    }
  }
`;

export const resolveMutation = graphql`
  mutation DocumentReviewMarginMutations_resolveMutation($input: ResolveDocumentCommentThreadInput!) {
    resolveDocumentCommentThread(input: $input) {
      documentCommentThread { id resolvedAt }
    }
  }
`;

export const reopenMutation = graphql`
  mutation DocumentReviewMarginMutations_reopenMutation($input: ReopenDocumentCommentThreadInput!) {
    reopenDocumentCommentThread(input: $input) {
      documentCommentThread { ...DocumentReviewMarginThread_thread }
    }
  }
`;

export const acceptMutation = graphql`
  mutation DocumentReviewMarginMutations_acceptMutation($input: AcceptDocumentSuggestionInput!) {
    acceptDocumentSuggestion(input: $input) {
      documentSuggestion { id status }
    }
  }
`;

export const rejectMutation = graphql`
  mutation DocumentReviewMarginMutations_rejectMutation($input: RejectDocumentSuggestionInput!) {
    rejectDocumentSuggestion(input: $input) {
      documentSuggestion { id status }
    }
  }
`;
