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

import { ChatCircleIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { createPortal } from "react-dom";
import { useTranslation } from "react-i18next";

import type { DocumentReviewMarginMutations_reopenMutation } from "#/__generated__/core/DocumentReviewMarginMutations_reopenMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { richEditorContentPlainText } from "#/pages/organizations/_lib/richEditorContent";

import {
  type ResolvedCommentThread,
  useResolvedCommentThreads,
} from "../../_lib/useResolvedCommentThreads";
import { reopenThreadEdge } from "../../_lib/reviewMarginStore";

import { reopenMutation } from "./DocumentReviewMarginMutations";

const mutationFeedback = { errorToast: false };

interface DocumentReviewMarginHistoryProps {
  documentId: string;
  toolbarHost: HTMLElement;
  historyOpen: boolean;
  onHistoryOpenChange: (open: boolean) => void;
  reviewVersion: number;
  onOwnReview: () => void;
  onOwnReviewFailed: () => void;
}

export function DocumentReviewMarginHistory({
  documentId,
  toolbarHost,
  historyOpen,
  onHistoryOpenChange,
  reviewVersion,
  onOwnReview,
  onOwnReviewFailed,
}: DocumentReviewMarginHistoryProps) {
  const { t } = useTranslation();
  const toast = useToast();
  const { resolvedThreads, setResolvedThreads } = useResolvedCommentThreads({
    documentId,
    historyOpen,
    reviewVersion,
  });
  const [reopenThread] = useMutation<DocumentReviewMarginMutations_reopenMutation>(
    reopenMutation,
    mutationFeedback,
  );

  return createPortal(
    <Popover open={historyOpen} onOpenChange={onHistoryOpenChange}>
      <PopoverTrigger
        render={(
          <Button
            variant="ghost"
            color="neutral"
            size={2}
            aria-label={t("documentDescriptionPage.review.history")}
            iconStart={<ChatCircleIcon />}
          />
        )}
      />
      <PopoverPopup align="end" side="bottom">
        <div className="flex w-72 flex-col gap-3 p-3">
          {resolvedThreads.length === 0 && (
            <p className="text-2 text-sand-11">{t("documentDescriptionPage.review.resolvedEmpty")}</p>
          )}
          {resolvedThreads.map((thread: ResolvedCommentThread) => (
            <div key={thread.id} className="flex flex-col gap-1">
              <p className="whitespace-pre-wrap text-2 text-sand-12">{thread.anchorQuote}</p>
              <p className="text-1 text-sand-11">
                {richEditorContentPlainText(thread.comments.edges[0]?.node.content ?? "")}
              </p>
              <Button
                variant="ghost"
                color="neutral"
                size={1}
                onClick={() => {
                  onOwnReview();
                  void reopenThread({
                    variables: { input: { threadId: thread.id } },
                    updater: store => reopenThreadEdge(store, documentId, thread.id),
                    onCompleted: (_response, errors) => {
                      if (errors?.length) {
                        onOwnReviewFailed();
                        return;
                      }
                      setResolvedThreads(current => current.filter(item => item.id !== thread.id));
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
                {t("documentDescriptionPage.review.reopen")}
              </Button>
            </div>
          ))}
        </div>
      </PopoverPopup>
    </Popover>,
    toolbarHost,
  );
}
