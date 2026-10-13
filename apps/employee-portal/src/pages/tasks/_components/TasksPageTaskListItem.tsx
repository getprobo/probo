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

import { TaskListItem, type TaskState } from "@probo/tasks";
import { graphql, useFragment } from "react-relay";

import type { TasksPageTaskListItem_task$key } from "#/__generated__/core/TasksPageTaskListItem_task.graphql";
import type { TasksPageTaskListItemUpdateStateMutation } from "#/__generated__/core/TasksPageTaskListItemUpdateStateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

const tasksPageTaskListItemFragment = graphql`
  fragment TasksPageTaskListItem_task on Task {
    id
    name
    state
    priority
    timeEstimate
    deadline
    recurrenceInterval
    canUpdate: permission(action: "core:task:update")
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

const updateStateMutation = graphql`
  mutation TasksPageTaskListItemUpdateStateMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TasksPage_task
        ...TasksPageTaskListItem_task
      }
    }
  }
`;

export interface TasksPageTaskListItemProps {
  taskKey: TasksPageTaskListItem_task$key;
  sectionState?: TaskState;
}

export function TasksPageTaskListItem({
  taskKey,
  sectionState,
}: TasksPageTaskListItemProps) {
  const { employeePortalId } = useEmployeePortalRoute();
  const { canUpdate, ...task } = useFragment(tasksPageTaskListItemFragment, taskKey);
  const [updateState, isUpdating] = useMutation<TasksPageTaskListItemUpdateStateMutation>(
    updateStateMutation,
  );

  function onStateChange(state: TaskState) {
    void updateState({
      variables: {
        input: {
          taskId: task.id,
          state,
        },
      },
    }).catch(() => {
      // Error feedback is handled by useMutation.
    });
  }

  return (
    <TaskListItem
      task={{
        id: task.id,
        name: task.name,
        state: task.state,
        priority: task.priority,
        timeEstimate: task.timeEstimate,
        deadline: task.deadline,
        recurrenceInterval: task.recurrenceInterval,
        assignedTo: task.assignedTo == null
          ? null
          : {
              id: task.assignedTo.id,
              fullName: task.assignedTo.fullName,
              emailAddress: task.assignedTo.emailAddress,
              avatarUrl: task.assignedTo.avatar?.downloadUrl,
            },
      }}
      canUpdate={canUpdate}
      isUpdating={isUpdating}
      detailsTo={`/${employeePortalId}/tasks/${task.id}`}
      sectionState={sectionState}
      onStateChange={onStateChange}
    />
  );
}
