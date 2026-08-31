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

import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskNameField_task$key } from "#/__generated__/core/TaskNameField_task.graphql";

import { useDebouncedSerializedFieldSave } from "../_lib/useSerializedFieldSave";
import { useUpdateTask } from "../_lib/useUpdateTask";
import { taskNameField } from "../variants";

const taskNameMaxLength = 1000;
const taskNameSaveDelayMs = 1000;

const taskNameFieldFragment = graphql`
  fragment TaskNameField_task on Task {
    id
    name
    canUpdate: permission(action: "core:task:update")
  }
`;

interface TaskNameFieldProps {
  taskKey: TaskNameField_task$key;
}

export function TaskNameField({ taskKey }: TaskNameFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskNameFieldFragment, taskKey);
  const [updateTask] = useUpdateTask();
  const [draft, setDraft] = useState(task.name);
  const [savedName, setSavedName] = useState(task.name);
  const [failedSave, setFailedSave] = useState<{
    next: string;
    nameAtStart: string;
  } | null>(null);
  const { root, input } = taskNameField();

  if (task.name !== savedName) {
    setSavedName(task.name);
    if (draft === savedName) {
      setDraft(task.name);
    }
  }

  if (failedSave) {
    setFailedSave(null);
    if (draft.trim() === failedSave.next && task.name === failedSave.nameAtStart) {
      setDraft(task.name);
    }
  }

  const persist = useCallback(
    async (value: string) => {
      const next = value.trim();
      const nameAtStart = task.name;
      if (!next || next === nameAtStart) {
        setDraft(current => (current === value ? nameAtStart : current));
        return;
      }

      try {
        await updateTask({
          variables: {
            input: {
              taskId: task.id,
              name: next,
            },
          },
        });
        setDraft(current => (current.trim() === next ? next : current));
      } catch {
        setFailedSave({ next, nameAtStart });
      }
    },
    [task.id, task.name, updateTask],
  );
  const persistDebounced = useDebouncedSerializedFieldSave(persist, taskNameSaveDelayMs);

  if (!task.canUpdate) {
    return (
      <Heading level={1} size={6} weight="medium" highContrast className="min-w-0 truncate">
        {task.name}
      </Heading>
    );
  }

  return (
    <div className={root()}>
      <input
        className={input()}
        value={draft}
        required
        maxLength={taskNameMaxLength}
        aria-label={t("detailsPage.fields.name")}
        onChange={(event) => {
          const next = event.currentTarget.value;
          setDraft(next);
          persistDebounced.schedule(next);
        }}
        onBlur={() => {
          persistDebounced.flush();
        }}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            event.preventDefault();
            persistDebounced.flush();
            event.currentTarget.blur();
          }
        }}
      />
    </div>
  );
}
