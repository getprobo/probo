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

import { UserCircleDashedIcon } from "@phosphor-icons/react";
import { taskListItem, TaskListItemAssignee, TaskProperties, TASKS_NAMESPACE } from "@probo/tasks";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskPropertiesSection_task$key } from "#/__generated__/core/TaskPropertiesSection_task.graphql";
import { TaskAssigneeMenu } from "#/components/tasks/TaskAssigneeMenu";

import { useUpdateTask } from "../_lib/useUpdateTask";

const noneValue = "__NONE__";

const taskPropertiesSectionFragment = graphql`
  fragment TaskPropertiesSection_task on Task {
    id
    state
    priority
    timeEstimate
    deadline
    recurrenceInterval
    canUpdate: permission(action: "core:task:update")
    canAssign: permission(action: "core:task:assign")
    assignedTo {
      id
      fullName
      emailAddress
      avatar {
        downloadUrl
      }
    }
  }
`;

interface TaskPropertiesSectionProps {
  taskKey: TaskPropertiesSection_task$key;
}

export function TaskPropertiesSection({ taskKey }: TaskPropertiesSectionProps) {
  const { t } = useTranslation("organizations/tasks");
  const { t: tTasks } = useTranslation(TASKS_NAMESPACE);
  const task = useFragment(taskPropertiesSectionFragment, taskKey);
  const [updateTask, isUpdating] = useUpdateTask();
  const assignedTo = task.assignedTo;
  const { stateOption } = taskListItem();

  function save(input: {
    state?: typeof task.state;
    priority?: typeof task.priority;
    assignedToId?: string | null;
    timeEstimate?: string | null;
    deadline?: string | null;
    recurrenceInterval?: string | null;
  }) {
    const result = updateTask({
      variables: {
        input: {
          taskId: task.id,
          ...input,
        },
      },
    }, { successMessage: "" });
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
      assignee={(
        <TaskListItemAssignee
          assignee={assignedTo == null
            ? null
            : {
                id: assignedTo.id,
                fullName: assignedTo.fullName,
                emailAddress: assignedTo.emailAddress,
                avatarUrl: assignedTo.avatar?.downloadUrl,
              }}
          canUpdate={task.canAssign}
          isUpdating={isUpdating}
          showLabel
          placeholder={tTasks("placeholders.assignee")}
          className={stateOption()}
          renderAssigneeMenu={task.canAssign
            ? () => (
                <TaskAssigneeMenu
                  value={assignedTo?.id ?? null}
                  emptyLabel={t("detailsPage.unassigned")}
                  emptyValue={noneValue}
                  emptyIcon={<UserCircleDashedIcon />}
                  disabled={isUpdating}
                  onValueChange={(assignedToId) => {
                    void save({ assignedToId });
                  }}
                />
              )
            : undefined}
        />
      )}
      onChange={change => save(change)}
    />
  );
}
