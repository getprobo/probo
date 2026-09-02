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
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ErrorInfo, type ReactNode, useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskDescriptionSection_task$key } from "#/__generated__/core/TaskDescriptionSection_task.graphql";

import { isRichEditorContentEmpty } from "../_lib/richEditorContent";
import { useDebouncedSerializedFieldSave } from "../_lib/useSerializedFieldSave";
import { useUpdateTask } from "../_lib/useUpdateTask";
import { taskDescriptionSection } from "../variants";

const taskDescriptionSaveDelayMs = 1000;

const taskDescriptionSectionFragment = graphql`
  fragment TaskDescriptionSection_task on Task {
    id
    content
    canUpdate: permission(action: "core:task:update")
  }
`;

interface TaskDescriptionSectionProps {
  taskKey: TaskDescriptionSection_task$key;
  fallback?: ReactNode;
  onError?: (error: unknown, info: ErrorInfo) => void;
}

function normalizeContent(value: string) {
  return isRichEditorContentEmpty(value) ? "" : value;
}

function TaskDescriptionSectionContent({
  taskKey,
}: {
  taskKey: TaskDescriptionSection_task$key;
}) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskDescriptionSectionFragment, taskKey);
  const [updateTask] = useUpdateTask();
  const saved = task.content;
  const [draft, setDraft] = useState(saved);
  const [savedContent, setSavedContent] = useState(saved);
  const [dirty, setDirty] = useState(false);
  const [editorGeneration, setEditorGeneration] = useState(0);

  if (saved !== savedContent) {
    setSavedContent(saved);
    if (!dirty) {
      setDraft(saved);
      setEditorGeneration(generation => generation + 1);
    }
  }

  const persist = useCallback(
    async (value: string) => {
      const next = normalizeContent(value);

      try {
        await updateTask(
          {
            variables: {
              input: {
                taskId: task.id,
                content: next || null,
              },
            },
          },
        );
        setDraft((current) => {
          if (current === value || normalizeContent(current) === next) {
            setDirty(false);
            return next;
          }
          return current;
        });
      } catch {
        setDraft((current) => {
          if (normalizeContent(current) === next) {
            setDirty(false);
            setEditorGeneration(generation => generation + 1);
            return saved;
          }
          return current;
        });
      }
    },
    [saved, task.id, updateTask],
  );
  const persistDebounced = useDebouncedSerializedFieldSave(
    persist,
    taskDescriptionSaveDelayMs,
  );
  const { root, editor } = taskDescriptionSection();

  return (
    <div className={root()}>
      {task.canUpdate
        ? (
            <RichEditor
              key={editorGeneration}
              className={editor()}
              content={draft}
              aria-label={t("detailsPage.fields.description")}
              onChangeContent={(next) => {
                setDirty(true);
                setDraft(next);
                persistDebounced.schedule(next);
              }}
              onBlur={() => {
                persistDebounced.flush();
              }}
            />
          )
        : isRichEditorContentEmpty(saved)
          ? (
              <Text size={2} color="faint">
                {t("detailsPage.noDescription")}
              </Text>
            )
          : (
              <RichEditor
                className={editor()}
                content={saved}
                disabled
                aria-label={t("detailsPage.fields.description")}
              />
            )}
    </div>
  );
}

export function TaskDescriptionSection({
  taskKey,
  fallback,
  onError,
}: TaskDescriptionSectionProps) {
  const { t } = useTranslation("organizations/tasks");

  return (
    <ErrorBoundary
      fallback={fallback ?? (
        <Text size={2} color="faint">
          {t("detailsPage.errors.content")}
        </Text>
      )}
      onError={onError}
    >
      <TaskDescriptionSectionContent taskKey={taskKey} />
    </ErrorBoundary>
  );
}
