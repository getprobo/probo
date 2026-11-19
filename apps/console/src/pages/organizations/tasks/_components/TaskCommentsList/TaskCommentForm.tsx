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
import { taskCommentForm } from "@probo/tasks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Field } from "@probo/ui/src/v2/form/Field";
import { shortcutAria } from "@probo/ui/src/v2/shortcutAria";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useParams } from "react-router";

import { RichEditor } from "#/pages/organizations/_components/RichEditor";
import {
  isRichEditorContentEmpty,
  richEditorContentTextLength,
} from "#/pages/organizations/_lib/richEditorContent";
import { useMentionMembers } from "#/pages/organizations/_lib/useMentionMembers";

import { taskCommentMaxLength } from "../../_lib/taskCommentMaxLength";
import { useCreateTaskComment } from "../../_lib/useCreateTaskComment";
import { useTaskCommentDraft } from "../../_lib/useTaskCommentDraft";

export function TaskCommentForm() {
  const { taskId } = useParams<{ taskId: string }>();

  // A new task id remounts the composer so its draft is read and flushed alone.
  return <TaskCommentComposer key={taskId ?? ""} />;
}

function TaskCommentComposer() {
  const { t } = useTranslation("organizations/tasks");
  const [createTaskComment, isCreating] = useCreateTaskComment();
  const mentionSource = useMentionMembers();
  const draft = useTaskCommentDraft();
  const [uploading, setUploading] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { root, editor } = taskCommentForm();

  function handleSubmit() {
    if (uploading || isCreating) {
      return;
    }

    if (isRichEditorContentEmpty(draft.content)) {
      setErrors({ content: t("detailsPage.comments.errors.contentRequired") });
      return;
    }

    if (richEditorContentTextLength(draft.content) > taskCommentMaxLength) {
      setErrors({
        content: t("detailsPage.comments.errors.contentTooLong", {
          max: taskCommentMaxLength,
        }),
      });
      return;
    }

    void createTaskComment(draft.content).then(
      () => {
        draft.clear();
        setErrors({});
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Card variant="soft" size={2}>
      <Form className={root()} onFormSubmit={handleSubmit}>
        <Field error={errors.content}>
          <RichEditor
            key={draft.editorKey}
            controls={false}
            attachments
            className={editor()}
            content={draft.content}
            disabled={isCreating}
            placeholder={t("detailsPage.comments.fields.leaveAComment")}
            aria-label={t("detailsPage.comments.fields.leaveAComment")}
            mentionSource={mentionSource}
            onModEnter={handleSubmit}
            toolbar={(
              <Button
                type="submit"
                size={1}
                variant="solid"
                color="neutral"
                highContrast
                disabled={uploading}
                loading={isCreating}
                aria-keyshortcuts={
                  uploading || isCreating ? undefined : shortcutAria("Enter")
                }
              >
                {t("detailsPage.comments.actions.send")}
              </Button>
            )}
            onChangeContent={(next) => {
              draft.setContent(next);
              if (richEditorContentTextLength(next) > taskCommentMaxLength) {
                setErrors({
                  content: t("detailsPage.comments.errors.contentTooLong", {
                    max: taskCommentMaxLength,
                  }),
                });
                return;
              }

              setErrors({});
            }}
            onUploadingChange={setUploading}
          />
        </Field>
      </Form>
    </Card>
  );
}
