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

import { formatError } from "@probo/helpers";
import type { MentionSource } from "@probo/ui";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentReviewMarginMutations_deleteCommentMutation } from "#/__generated__/core/DocumentReviewMarginMutations_deleteCommentMutation.graphql";
import type { DocumentReviewMarginMutations_replyMutation } from "#/__generated__/core/DocumentReviewMarginMutations_replyMutation.graphql";
import type { DocumentReviewMarginMutations_resolveMutation } from "#/__generated__/core/DocumentReviewMarginMutations_resolveMutation.graphql";
import type { DocumentReviewMarginMutations_updateCommentMutation } from "#/__generated__/core/DocumentReviewMarginMutations_updateCommentMutation.graphql";
import type { DocumentReviewMarginThread_thread$key } from "#/__generated__/core/DocumentReviewMarginThread_thread.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import {
  deleteCommentEdge,
  deleteOpenThread,
  insertCommentEdge,
} from "../../_lib/reviewMarginStore";

import { DocumentReviewMarginCard } from "./DocumentReviewMarginCard";
import { DocumentReviewMarginComments } from "./DocumentReviewMarginComments";
import {
  deleteCommentMutation,
  replyMutation,
  resolveMutation,
  updateCommentMutation,
} from "./DocumentReviewMarginMutations";

const mutationFeedback = { errorToast: false };

export const documentReviewMarginThreadFragment = graphql`
  fragment DocumentReviewMarginThread_thread on DocumentCommentThread {
    id
    anchorQuote
    anchorStart
    anchorEnd
    resolvedAt
    suggestionId
    comments(first: 100) @connection(key: "DocumentReviewMargin_comments") {
      edges {
        node {
          id
          content
          createdAt
          author { id fullName }
        }
      }
      pageInfo { hasNextPage endCursor }
    }
  }
`;

interface DocumentReviewMarginThreadProps {
  threadKey: DocumentReviewMarginThread_thread$key;
  documentId: string;
  active: boolean;
  profileId: string | null;
  canComment: boolean;
  mentionSource: MentionSource;
  onActivate: () => void;
  onResolved: () => void;
  onOwnReview: () => void;
  onOwnReviewFailed: () => void;
  onLayout?: () => void;
}

export function DocumentReviewMarginThread({
  threadKey,
  documentId,
  active,
  profileId,
  canComment,
  mentionSource,
  onActivate,
  onResolved,
  onOwnReview,
  onOwnReviewFailed,
  onLayout,
}: DocumentReviewMarginThreadProps) {
  const { t } = useTranslation();
  const toast = useToast();
  const thread = useFragment(documentReviewMarginThreadFragment, threadKey);
  const [replyText, setReplyText] = useState("");
  const [reply, replying] = useMutation<DocumentReviewMarginMutations_replyMutation>(
    replyMutation,
    mutationFeedback,
  );
  const [updateComment] = useMutation<DocumentReviewMarginMutations_updateCommentMutation>(
    updateCommentMutation,
    mutationFeedback,
  );
  const [deleteComment] = useMutation<DocumentReviewMarginMutations_deleteCommentMutation>(
    deleteCommentMutation,
    mutationFeedback,
  );
  const [resolveThread] = useMutation<DocumentReviewMarginMutations_resolveMutation>(
    resolveMutation,
    mutationFeedback,
  );

  const reportError = (error: Error) => {
    toast.add({
      title: t("documentDescriptionPage.errors.title"),
      description: error.message,
      type: "error",
    });
  };

  return (
    <DocumentReviewMarginCard
      active={active}
      label={t("documentDescriptionPage.review.selectComment", {
        name: thread.comments.edges[0]?.node.author.fullName ?? "",
      })}
      onActivate={onActivate}
    >
      <DocumentReviewMarginComments
        thread={thread}
        active={active}
        profileId={profileId}
        canComment={canComment}
        replyText={active ? replyText : ""}
        mentionSource={mentionSource}
        onReplyText={setReplyText}
        replying={replying}
        onReply={(content) => {
          onOwnReview();
          void reply({
            variables: { input: { threadId: thread.id, content } },
            updater: store => insertCommentEdge(store, thread.id),
            onCompleted: (_response, errors) => {
              if (errors?.length) {
                onOwnReviewFailed();
                toast.add({
                  title: t("documentDescriptionPage.errors.title"),
                  description: formatError(t("documentDescriptionPage.review.commentFailed"), errors),
                  type: "error",
                });
                return;
              }
              setReplyText("");
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
        onUpdate={(commentId, content) => {
          onOwnReview();
          void updateComment({
            variables: { input: { commentId, content } },
            onCompleted: (_response, errors) => {
              if (errors?.length) onOwnReviewFailed();
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
        onDelete={(commentId) => {
          onOwnReview();
          void deleteComment({
            variables: { input: { commentId } },
            updater: store => deleteCommentEdge(store, thread.id, commentId),
            onCompleted: (_response, errors) => {
              if (errors?.length) onOwnReviewFailed();
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
        onLayout={onLayout}
        onResolve={() => {
          onOwnReview();
          void resolveThread({
            variables: { input: { threadId: thread.id } },
            updater: store => deleteOpenThread(store, documentId, thread.id),
            onCompleted: (_response, errors) => {
              if (errors?.length) {
                onOwnReviewFailed();
                return;
              }
              onResolved();
            },
            onError: (error) => {
              onOwnReviewFailed();
              reportError(error);
            },
          });
        }}
      />
    </DocumentReviewMarginCard>
  );
}
