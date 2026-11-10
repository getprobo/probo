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

import { CheckIcon, XIcon } from "@phosphor-icons/react";
import type { MentionSource } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentReviewMarginMutations_acceptMutation } from "#/__generated__/core/DocumentReviewMarginMutations_acceptMutation.graphql";
import type { DocumentReviewMarginMutations_createThreadMutation } from "#/__generated__/core/DocumentReviewMarginMutations_createThreadMutation.graphql";
import type { DocumentReviewMarginMutations_deleteCommentMutation } from "#/__generated__/core/DocumentReviewMarginMutations_deleteCommentMutation.graphql";
import type { DocumentReviewMarginMutations_rejectMutation } from "#/__generated__/core/DocumentReviewMarginMutations_rejectMutation.graphql";
import type { DocumentReviewMarginMutations_replyMutation } from "#/__generated__/core/DocumentReviewMarginMutations_replyMutation.graphql";
import type { DocumentReviewMarginMutations_updateCommentMutation } from "#/__generated__/core/DocumentReviewMarginMutations_updateCommentMutation.graphql";
import type { DocumentReviewMarginSuggestion_suggestion$key } from "#/__generated__/core/DocumentReviewMarginSuggestion_suggestion.graphql";
import type { DocumentReviewMarginThread_thread$key } from "#/__generated__/core/DocumentReviewMarginThread_thread.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import type { SuggestionPreview } from "../../_lib/suggestionPreview";
import {
  closeSuggestion,
  deleteCommentEdge,
  insertCommentEdge,
  insertThreadEdge,
} from "../../_lib/reviewMarginStore";

import { DocumentReviewMarginCard } from "./DocumentReviewMarginCard";
import { DocumentReviewMarginComments } from "./DocumentReviewMarginComments";
import {
  acceptMutation,
  createThreadMutation,
  deleteCommentMutation,
  rejectMutation,
  replyMutation,
  updateCommentMutation,
} from "./DocumentReviewMarginMutations";
import { documentReviewMarginThreadFragment } from "./DocumentReviewMarginThread";

const mutationFeedback = { errorToast: false };

export const documentReviewMarginSuggestionFragment = graphql`
  fragment DocumentReviewMarginSuggestion_suggestion on DocumentSuggestion {
    id
    status
    createdAt
    author { id fullName }
  }
`;

type LocalSuggestion = {
  id: string;
  author: { fullName: string | null | undefined };
};

interface DocumentReviewMarginSuggestionProps {
  documentId: string;
  suggestionKey?: DocumentReviewMarginSuggestion_suggestion$key | null;
  localSuggestion?: LocalSuggestion | null;
  threadKey?: DocumentReviewMarginThread_thread$key | null;
  preview: Pick<SuggestionPreview, "parts">;
  active: boolean;
  canUpdate: boolean;
  canComment: boolean;
  profileId: string | null;
  mentionSource: MentionSource;
  onActivate: () => void;
  onDecided: (suggestionId: string) => void;
  onOwnReview: () => void;
  onOwnReviewFailed: () => void;
  onLayout?: () => void;
}

export function DocumentReviewMarginSuggestion({
  documentId,
  suggestionKey,
  localSuggestion,
  threadKey,
  preview,
  active,
  canUpdate,
  canComment,
  profileId,
  mentionSource,
  onActivate,
  onDecided,
  onOwnReview,
  onOwnReviewFailed,
  onLayout,
}: DocumentReviewMarginSuggestionProps) {
  const { t } = useTranslation();
  const toast = useToast();
  const loaded = useFragment(documentReviewMarginSuggestionFragment, suggestionKey ?? null);
  const thread = useFragment(documentReviewMarginThreadFragment, threadKey ?? null);
  const suggestion = loaded ?? localSuggestion;
  const [replyText, setReplyText] = useState("");
  const [createThread] = useMutation<DocumentReviewMarginMutations_createThreadMutation>(
    createThreadMutation,
    mutationFeedback,
  );
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
  const [acceptSuggestion] = useMutation<DocumentReviewMarginMutations_acceptMutation>(
    acceptMutation,
    mutationFeedback,
  );
  const [rejectSuggestion] = useMutation<DocumentReviewMarginMutations_rejectMutation>(
    rejectMutation,
    mutationFeedback,
  );

  if (!suggestion) return null;

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
      label={t("documentDescriptionPage.review.selectSuggestion", {
        name: suggestion.author.fullName ?? "",
      })}
      onActivate={onActivate}
    >
      <div className="flex items-start justify-between gap-2">
        <p className="text-2 font-medium text-sand-12">{suggestion.author.fullName}</p>
        {canUpdate && !suggestion.id.startsWith("local:") && (
          <div className="flex gap-1">
            <Button
              variant="ghost"
              color="green"
              size={1}
              aria-label={t("documentDescriptionPage.review.accept")}
              iconStart={<CheckIcon />}
              onClick={() => {
                onOwnReview();
                void acceptSuggestion({
                  variables: { input: { suggestionId: suggestion.id } },
                  updater: store => closeSuggestion(store, documentId, suggestion.id),
                  onCompleted: (_response, errors) => {
                    if (errors?.length) {
                      onOwnReviewFailed();
                      return;
                    }
                    onDecided(suggestion.id);
                  },
                  onError: (error) => {
                    onOwnReviewFailed();
                    reportError(error);
                  },
                });
              }}
            />
            <Button
              variant="ghost"
              color="red"
              size={1}
              aria-label={t("documentDescriptionPage.review.reject")}
              iconStart={<XIcon />}
              onClick={() => {
                onOwnReview();
                void rejectSuggestion({
                  variables: { input: { suggestionId: suggestion.id } },
                  updater: store => closeSuggestion(store, documentId, suggestion.id),
                  onCompleted: (_response, errors) => {
                    if (errors?.length) {
                      onOwnReviewFailed();
                      return;
                    }
                    onDecided(suggestion.id);
                  },
                  onError: (error) => {
                    onOwnReviewFailed();
                    reportError(error);
                  },
                });
              }}
            />
          </div>
        )}
      </div>
      <p className="whitespace-pre-wrap text-2 text-sand-12">
        {preview.parts.map((part, index) =>
          part.kind === "deleted"
            ? (
                <span key={index} className="text-red-11 line-through">{part.text}</span>
              )
            : part.kind === "inserted"
              ? (
                  <span key={index} className="text-green-11">{part.text}</span>
                )
              : (
                  <span key={index}>{part.text}</span>
                ),
        )}
        {preview.parts.length === 0 && thread?.anchorQuote}
      </p>
      <DocumentReviewMarginComments
        thread={thread}
        active={active}
        profileId={profileId}
        canComment={canComment && !suggestion.id.startsWith("local:")}
        replyText={active ? replyText : ""}
        mentionSource={mentionSource}
        onReplyText={setReplyText}
        replying={replying}
        onLayout={onLayout}
        onReply={(content) => {
          if (thread) {
            onOwnReview();
            void reply({
              variables: { input: { threadId: thread.id, content } },
              updater: store => insertCommentEdge(store, thread.id),
              onCompleted: (_response, errors) => {
                if (errors?.length) {
                  onOwnReviewFailed();
                  return;
                }
                setReplyText("");
              },
              onError: (error) => {
                onOwnReviewFailed();
                reportError(error);
              },
            });
            return;
          }
          onOwnReview();
          void createThread({
            variables: { input: { documentId: documentId, suggestionId: suggestion.id, content } },
            updater: store => insertThreadEdge(store, documentId),
            onCompleted: (_response, errors) => {
              if (errors?.length) {
                onOwnReviewFailed();
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
          if (!thread) return;
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
      />
    </DocumentReviewMarginCard>
  );
}
