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

import { CheckIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { useLayoutEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

export interface DocumentReviewMarginThread {
  id: string;
  anchorQuote: string;
  anchorStart: string;
  anchorEnd: string;
  resolvedAt?: string | null;
  suggestionId?: string | null;
  comments: {
    edges: ReadonlyArray<{
      node: {
        id: string;
        content: string;
        author: { id: string; fullName: string | null | undefined };
      };
    }>;
  };
}

export interface DocumentReviewMarginCommentsProps {
  thread?: DocumentReviewMarginThread | null;
  active: boolean;
  profileId: string | null;
  canComment: boolean;
  replyText: string;
  replying: boolean;
  onReplyText: (value: string) => void;
  onReply: (content: string) => void;
  onUpdate: (commentId: string, content: string) => void;
  onDelete: (commentId: string) => void;
  onResolve?: () => void;
  onLayout?: () => void;
}

export function DocumentReviewMarginComments({
  thread,
  active,
  profileId,
  canComment,
  replyText,
  replying,
  onReplyText,
  onReply,
  onUpdate,
  onDelete,
  onResolve,
  onLayout,
}: DocumentReviewMarginCommentsProps) {
  const { t } = useTranslation();
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingText, setEditingText] = useState("");
  const root = useRef<HTMLDivElement>(null);
  const lastHeight = useRef(0);
  const comments = thread?.comments.edges ?? [];
  const visible = active ? comments : comments.slice(0, 1);
  useLayoutEffect(() => {
    const height = root.current?.offsetHeight ?? 0;
    if (height === lastHeight.current) return;
    lastHeight.current = height;
    onLayout?.();
  });

  return (
    <div ref={root} className="flex flex-col gap-2">
      {onResolve && (
        <div className="flex justify-end">
          <Button
            variant="ghost"
            color="neutral"
            size={1}
            aria-label={t("documentDescriptionPage.review.resolve")}
            iconStart={<CheckIcon />}
            onClick={onResolve}
          />
        </div>
      )}
      {visible.map(({ node }) => (
        <div key={node.id} className="group text-2 text-sand-12">
          <p className="font-medium">{node.author.fullName}</p>
          {editingId === node.id
            ? (
                <form
                  onSubmit={(event) => {
                    event.preventDefault();
                    onUpdate(node.id, editingText);
                    setEditingId(null);
                  }}
                >
                  <textarea
                    className="min-h-12 w-full rounded-2 border border-sand-7 bg-sand-1 p-2 text-2 text-sand-12"
                    value={editingText}
                    onChange={event => setEditingText(event.target.value)}
                  />
                  <Button variant="solid" color="neutral" size={1} type="submit">
                    {t("documentDescriptionPage.review.save")}
                  </Button>
                </form>
              )
            : <p>{node.content}</p>}
          {active && profileId === node.author.id && editingId !== node.id && (
            <div className="mt-1 flex gap-2 text-1 text-sand-11 opacity-0 group-hover:opacity-100">
              <button
                type="button"
                onClick={() => {
                  setEditingId(node.id);
                  setEditingText(node.content);
                }}
              >
                {t("documentDescriptionPage.review.edit")}
              </button>
              <button type="button" onClick={() => onDelete(node.id)}>
                {t("documentDescriptionPage.review.delete")}
              </button>
            </div>
          )}
        </div>
      ))}
      {active && canComment && (
        <form
          className="flex flex-col gap-1"
          onSubmit={(event) => {
            event.preventDefault();
            if (replyText.trim() === "") return;
            onReply(replyText);
          }}
        >
          <textarea
            className="min-h-12 rounded-2 border border-sand-7 bg-sand-1 p-2 text-2 text-sand-12 outline-none focus:border-sand-8"
            value={replyText}
            onChange={event => onReplyText(event.target.value)}
            placeholder={t("documentDescriptionPage.review.replyPlaceholder")}
          />
          <div className="flex justify-end">
            <Button variant="solid" color="indigo" size={1} type="submit" disabled={replying}>
              {t("documentDescriptionPage.review.reply")}
            </Button>
          </div>
        </form>
      )}
    </div>
  );
}
