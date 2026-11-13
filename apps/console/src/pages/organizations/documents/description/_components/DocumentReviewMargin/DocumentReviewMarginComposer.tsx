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
import { CommentEditor, type MentionSource } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import type { DocumentReviewMarginMutations_createThreadMutation } from "#/__generated__/core/DocumentReviewMarginMutations_createThreadMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import {
  isRichEditorContentEmpty,
  richEditorContentTextLength,
} from "#/pages/organizations/_lib/richEditorContent";

import { insertThreadEdge } from "../../_lib/reviewMarginStore";
import { type CommentDraft } from "../../_lib/useReviewMarginLayout";

import { DocumentReviewMarginCard } from "./DocumentReviewMarginCard";
import { createThreadMutation } from "./DocumentReviewMarginMutations";

const commentMaxLength = 10000;
const mutationFeedback = { errorToast: false };

interface DocumentReviewMarginComposerProps {
  documentId: string;
  draft: CommentDraft;
  mentionSource: MentionSource;
  onOwnReview: () => void;
  onOwnReviewFailed: () => void;
  onCancel: () => void;
  onCreated: () => void;
}

export function DocumentReviewMarginComposer({
  documentId,
  draft,
  mentionSource,
  onOwnReview,
  onOwnReviewFailed,
  onCancel,
  onCreated,
}: DocumentReviewMarginComposerProps) {
  const { t } = useTranslation();
  const toast = useToast();
  const [draftText, setDraftText] = useState("");
  const [draftEditorKey, setDraftEditorKey] = useState(0);
  const [createThread, creatingThread] = useMutation<DocumentReviewMarginMutations_createThreadMutation>(
    createThreadMutation,
    mutationFeedback,
  );

  return (
    <DocumentReviewMarginCard
      active
      label={t("documentDescriptionPage.review.commentPlaceholder")}
    >
      <form
        className="flex flex-col gap-2"
        onSubmit={(event) => {
          event.preventDefault();
          if (isRichEditorContentEmpty(draftText)) return;
          if (richEditorContentTextLength(draftText) > commentMaxLength) return;
          onOwnReview();
          void createThread({
            variables: {
              input: {
                documentId: documentId,
                content: draftText,
                quote: draft.quote,
                anchorStart: draft.anchorStart,
                anchorEnd: draft.anchorEnd,
              },
            },
            updater: store => insertThreadEdge(store, documentId),
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
              setDraftText("");
              setDraftEditorKey(key => key + 1);
              onCreated();
            },
            onError: (error) => {
              onOwnReviewFailed();
              toast.add({
                title: t("documentDescriptionPage.errors.title"),
                description: error.message,
                type: "error",
              });
            },
          });
        }}
      >
        <CommentEditor
          key={draftEditorKey}
          autoFocus
          controls={false}
          className="min-h-16 w-full rounded-2 border border-sand-7 bg-sand-1 p-2 text-2 outline-none focus-within:border-sand-8"
          content={draftText}
          mentionSource={mentionSource}
          placeholder={t("documentDescriptionPage.review.commentPlaceholder")}
          onChangeContent={setDraftText}
          aria-label={t("documentDescriptionPage.review.commentPlaceholder")}
        />
        <div className="flex justify-end gap-2">
          <Button variant="ghost" color="neutral" size={1} onClick={onCancel}>
            {t("documentDescriptionPage.review.cancel")}
          </Button>
          <Button variant="solid" color="indigo" size={1} type="submit" disabled={creatingThread}>
            {t("documentDescriptionPage.review.comment")}
          </Button>
        </div>
      </form>
    </DocumentReviewMarginCard>
  );
}
