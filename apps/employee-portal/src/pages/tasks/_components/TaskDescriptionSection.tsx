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

import { TASKS_NAMESPACE, taskDescription } from "@probo/tasks";
import { RichEditor } from "@probo/ui";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ErrorInfo, type ReactNode, useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskDescriptionSection_task$key } from "#/__generated__/core/TaskDescriptionSection_task.graphql";

import { isRichEditorContentEmpty } from "../_lib/richEditorContent";
import { useDebouncedSerializedFieldSave } from "../_lib/useDebouncedSerializedFieldSave";
import { useUpdateTask } from "../_lib/useUpdateTask";

const taskDescriptionSectionFragment = graphql`
  fragment TaskDescriptionSection_task on Task {
    id
    content
    canUpdate: permission(action: "core:task:update")
  }
`;

const defaultSaveDelayMs = 1000;

function normalizeContent(value: string) {
  return isRichEditorContentEmpty(value) ? "" : value;
}

interface TaskDescriptionSectionProps {
  taskKey: TaskDescriptionSection_task$key;
  fallback?: ReactNode;
  onError?: (error: unknown, info: ErrorInfo) => void;
}

function TaskDescriptionSectionContent({
  taskKey,
}: {
  taskKey: TaskDescriptionSection_task$key;
}) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const { t: tPage } = useTranslation("tasks");
  const task = useFragment(taskDescriptionSectionFragment, taskKey);
  const [updateTask] = useUpdateTask();
  const saveRef = useRef(async (content: string | null) => {
    await updateTask({
      variables: {
        input: {
          taskId: task.id,
          content,
        },
      },
    });
  });
  const saved = task.content;
  const savedRef = useRef(saved);
  const [draft, setDraft] = useState(saved);
  const [savedContent, setSavedContent] = useState(saved);
  const [dirty, setDirty] = useState(false);
  const [editorGeneration, setEditorGeneration] = useState(0);
  const { root, editor } = taskDescription();

  useEffect(() => {
    saveRef.current = async (content: string | null) => {
      await updateTask({
        variables: {
          input: {
            taskId: task.id,
            content,
          },
        },
      });
    };
  }, [task.id, updateTask]);

  useEffect(() => {
    savedRef.current = saved;
  }, [saved]);

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
        await saveRef.current(next || null);
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
            return savedRef.current;
          }
          return current;
        });
      }
    },
    [],
  );
  const persistDebounced = useDebouncedSerializedFieldSave(persist, defaultSaveDelayMs);

  return (
    <div className={root()}>
      {task.canUpdate
        ? (
            <RichEditor
              key={editorGeneration}
              className={editor()}
              content={draft}
              aria-label={t("fields.description")}
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
                {tPage("detail.noDescription")}
              </Text>
            )
          : (
              <RichEditor
                className={editor()}
                content={saved}
                disabled
                aria-label={t("fields.description")}
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
  const { t } = useTranslation("tasks");

  return (
    <ErrorBoundary
      fallback={fallback ?? (
        <Text size={2} color="faint">
          {t("detail.errors.content")}
        </Text>
      )}
      onError={onError}
    >
      <TaskDescriptionSectionContent taskKey={taskKey} />
    </ErrorBoundary>
  );
}
