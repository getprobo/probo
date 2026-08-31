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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskDescriptionSection_task$key } from "#/__generated__/core/TaskDescriptionSection_task.graphql";

import { useDebouncedSerializedFieldSave } from "../_lib/useSerializedFieldSave";
import { useUpdateTask } from "../_lib/useUpdateTask";
import { taskDescriptionSection } from "../variants";

const taskDescriptionMaxLength = 5000;
const taskDescriptionSaveDelayMs = 1000;

const taskDescriptionSectionFragment = graphql`
  fragment TaskDescriptionSection_task on Task {
    id
    description
    canUpdate: permission(action: "core:task:update")
  }
`;

interface TaskDescriptionSectionProps {
  taskKey: TaskDescriptionSection_task$key;
}

export function TaskDescriptionSection({ taskKey }: TaskDescriptionSectionProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskDescriptionSectionFragment, taskKey);
  const [updateTask] = useUpdateTask();
  const saved = task.description ?? "";
  const [draft, setDraft] = useState(saved);
  const [savedDescription, setSavedDescription] = useState(saved);
  const [dirty, setDirty] = useState(false);

  if (saved !== savedDescription) {
    setSavedDescription(saved);
    if (!dirty) {
      setDraft(saved);
    }
  }

  const persist = useCallback(
    async (value: string) => {
      const next = value.trim();

      try {
        await updateTask({
          variables: {
            input: {
              taskId: task.id,
              description: next || null,
            },
          },
        });
        setDraft((current) => {
          if (current === value || current.trim() === next) {
            setDirty(false);
            return next;
          }
          return current;
        });
      } catch {
        setDraft((current) => {
          if (current.trim() === next) {
            setDirty(false);
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
  const { root, textarea } = taskDescriptionSection();

  return (
    <div className={root()}>
      {task.canUpdate
        ? (
            <textarea
              className={textarea()}
              rows={1}
              value={draft}
              maxLength={taskDescriptionMaxLength}
              placeholder={t("detailsPage.descriptionPlaceholder")}
              aria-label={t("detailsPage.fields.description")}
              onChange={(event) => {
                const next = event.currentTarget.value;
                setDirty(true);
                setDraft(next);
                persistDebounced.schedule(next);
              }}
              onBlur={() => {
                persistDebounced.flush();
              }}
            />
          )
        : task.description
          ? (
              <Text size={2} className="whitespace-pre-wrap wrap-break-word">
                {task.description}
              </Text>
            )
          : (
              <Text size={2} color="faint">
                {t("detailsPage.noDescription")}
              </Text>
            )}
    </div>
  );
}
