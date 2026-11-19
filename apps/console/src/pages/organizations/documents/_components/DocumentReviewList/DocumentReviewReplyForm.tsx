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

import { Form } from "@base-ui/react/form";
import { CommentEditor } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Field } from "@probo/ui/src/v2/form/Field";
import { shortcutAria } from "@probo/ui/src/v2/shortcutAria";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import {
  isRichEditorContentEmpty,
  richEditorContentTextLength,
} from "#/pages/organizations/_lib/richEditorContent";
import { useMentionMembers } from "#/pages/organizations/_lib/useMentionMembers";

import { documentCommentMaxLength } from "../../_lib/documentCommentMaxLength";
import { documentReviewReplyForm } from "../../variants";

interface DocumentReviewReplyFormProps {
  onSubmit: (content: string) => Promise<void>;
}

export function DocumentReviewReplyForm({ onSubmit }: DocumentReviewReplyFormProps) {
  const { t } = useTranslation("organizations/documents");
  const mentionSource = useMentionMembers();
  const [content, setContent] = useState("");
  const [editorKey, setEditorKey] = useState(0);
  const [error, setError] = useState<string>();
  const [submitting, setSubmitting] = useState(false);
  const { root, editor, send } = documentReviewReplyForm();

  function handleSubmit() {
    if (submitting) {
      return;
    }
    if (isRichEditorContentEmpty(content)) {
      setError(t("details.review.contentRequired"));
      return;
    }
    if (richEditorContentTextLength(content) > documentCommentMaxLength) {
      setError(t("details.review.contentTooLong", { max: documentCommentMaxLength }));
      return;
    }

    setSubmitting(true);
    onSubmit(content).then(
      () => {
        setContent("");
        setEditorKey(key => key + 1);
        setError(undefined);
      },
      () => {
        // The mutation already shows an error toast.
      },
    ).finally(() => {
      setSubmitting(false);
    });
  }

  return (
    <Form className={root()} onFormSubmit={handleSubmit}>
      <Field error={error}>
        <CommentEditor
          key={editorKey}
          className={editor()}
          content={content}
          controls={false}
          disabled={submitting}
          mentionSource={mentionSource}
          placeholder={t("details.review.replyPlaceholder")}
          aria-label={t("details.review.reply")}
          onModEnter={handleSubmit}
          toolbar={(
            <Button
              type="submit"
              size={1}
              variant="solid"
              color="neutral"
              highContrast
              className={send()}
              loading={submitting}
              aria-keyshortcuts={submitting ? undefined : shortcutAria("Enter")}
            >
              {t("details.review.send")}
            </Button>
          )}
          onChangeContent={(next) => {
            setContent(next);
            setError(richEditorContentTextLength(next) > documentCommentMaxLength
              ? t("details.review.contentTooLong", { max: documentCommentMaxLength })
              : undefined);
          }}
        />
      </Field>
    </Form>
  );
}
