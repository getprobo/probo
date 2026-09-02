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

import { RichEditor } from "@probo/ui";
import { Field } from "@probo/ui/src/v2/form/Field";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskCommentEditor_taskComment$key } from "#/__generated__/core/TaskCommentEditor_taskComment.graphql";

import { richEditorContentTextLength } from "../_lib/richEditorContent";
import { taskCommentMaxLength } from "../_lib/taskCommentMaxLength";
import { useDebouncedSerializedFieldSave } from "../_lib/useSerializedFieldSave";
import { useUpdateTaskComment } from "../_lib/useUpdateTaskComment";
import { taskCommentEditor } from "../variants";

const taskCommentSaveDelayMs = 1000;

const taskCommentEditorFragment = graphql`
  fragment TaskCommentEditor_taskComment on TaskComment {
    id
    content
  }
`;

interface TaskCommentEditorProps {
  taskCommentKey: TaskCommentEditor_taskComment$key;
}

export function TaskCommentEditor({ taskCommentKey }: TaskCommentEditorProps) {
  const { t } = useTranslation("organizations/tasks");
  const comment = useFragment(taskCommentEditorFragment, taskCommentKey);
  const [updateTaskComment] = useUpdateTaskComment();
  const saved = comment.content;
  const [draft, setDraft] = useState(saved);
  const [savedContent, setSavedContent] = useState(saved);
  const [dirty, setDirty] = useState(false);
  const [editorGeneration, setEditorGeneration] = useState(0);
  const [error, setError] = useState<string | undefined>();

  if (saved !== savedContent) {
    setSavedContent(saved);
    if (!dirty) {
      setDraft(saved);
      setEditorGeneration(generation => generation + 1);
    }
  }

  const persist = useCallback(
    async (value: string) => {
      if (richEditorContentTextLength(value) > taskCommentMaxLength) {
        setError(
          t("detailsPage.comments.errors.contentTooLong", {
            max: taskCommentMaxLength,
          }),
        );
        return;
      }

      try {
        await updateTaskComment(
          {
            variables: {
              input: {
                taskCommentId: comment.id,
                content: value,
              },
            },
          },
        );
        setDraft((current) => {
          if (current === value) {
            setDirty(false);
            setError(undefined);
            return value;
          }
          return current;
        });
      } catch {
        setDraft((current) => {
          if (current === value) {
            setDirty(false);
            setEditorGeneration(generation => generation + 1);
            setError(t("detailsPage.comments.errors.update"));
            return saved;
          }
          return current;
        });
      }
    },
    [comment.id, saved, t, updateTaskComment],
  );
  const persistDebounced = useDebouncedSerializedFieldSave(
    persist,
    taskCommentSaveDelayMs,
  );

  return (
    <Field error={error}>
      <RichEditor
        key={editorGeneration}
        className={taskCommentEditor()}
        content={draft}
        aria-label={t("detailsPage.comments.fields.comment")}
        onChangeContent={(next) => {
          setDirty(true);
          setDraft(next);
          if (richEditorContentTextLength(next) > taskCommentMaxLength) {
            setError(
              t("detailsPage.comments.errors.contentTooLong", {
                max: taskCommentMaxLength,
              }),
            );
            return;
          }

          setError(undefined);
          persistDebounced.schedule(next);
        }}
        onBlur={() => {
          persistDebounced.flush();
        }}
      />
    </Field>
  );
}
