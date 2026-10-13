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

import {
  TASKS_NAMESPACE,
  TaskDurationField,
  TaskProperties,
  taskRecurrenceDurationUnits,
} from "@probo/tasks";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskPropertiesSection_task$key } from "#/__generated__/core/TaskPropertiesSection_task.graphql";

import { useUpdateTask } from "../_lib/useUpdateTask";

const taskPropertiesSectionFragment = graphql`
  fragment TaskPropertiesSection_task on Task {
    id
    state
    priority
    timeEstimate
    deadline
    recurrenceInterval
    createdAt
    updatedAt
    canUpdate: permission(action: "core:task:update")
  }
`;

interface TaskPropertiesSectionProps {
  taskKey: TaskPropertiesSection_task$key;
}

export function TaskPropertiesSection({ taskKey }: TaskPropertiesSectionProps) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const task = useFragment(taskPropertiesSectionFragment, taskKey);
  const [updateTask, isUpdating] = useUpdateTask();

  function save(
    input: {
      state?: typeof task.state;
      priority?: typeof task.priority;
      timeEstimate?: string | null;
      deadline?: string | null;
      recurrenceInterval?: string | null;
    },
  ) {
    const result = updateTask({
      variables: {
        input: {
          taskId: task.id,
          ...input,
        },
      },
    });
    void result.catch(() => {
      // Error toast is already shown by useMutation.
    });
    return result;
  }

  return (
    <TaskProperties
      task={task}
      canUpdate={task.canUpdate}
      isUpdating={isUpdating}
      timeEstimateField={task.canUpdate
        ? (
            <TaskDurationField
              value={task.timeEstimate ?? null}
              onValueChange={timeEstimate => save({ timeEstimate })}
            />
          )
        : undefined}
      recurrenceField={task.canUpdate
        ? (
            <TaskDurationField
              value={task.recurrenceInterval ?? null}
              disabled={!task.deadline && !task.recurrenceInterval}
              defaultValue="P7D"
              units={taskRecurrenceDurationUnits}
              addLabel={t("actions.addRecurrence")}
              addTitle={!task.deadline
                ? t("hints.recurrenceNeedsDeadline")
                : undefined}
              clearLabel={t("actions.clearRecurrence")}
              amountAriaLabel={t("fields.recurrence")}
              unitAriaLabel={t("fields.recurrenceUnit")}
              onValueChange={recurrenceInterval => save({ recurrenceInterval })}
            />
          )
        : undefined}
      onChange={(change) => {
        void save(change);
      }}
    />
  );
}
