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

import { TaskListItem as SharedTaskListItem, type TaskState } from "@probo/tasks";
import { graphql, useFragment } from "react-relay";

import type { TaskListItem_task$key } from "#/__generated__/core/TaskListItem_task.graphql";
import type { TaskListItemUpdateStateMutation } from "#/__generated__/core/TaskListItemUpdateStateMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { taskDetailsPath } from "#/pages/organizations/tasks/_lib/taskPath";

const taskListItemFragment = graphql`
  fragment TaskListItem_task on Task {
    id
    name
    state
    priority
    timeEstimate
    deadline
    recurrenceInterval
    canUpdate: permission(action: "core:task:update")
    externalLink {
      identifier
      url
    }
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
  mutation TaskListItemUpdateStateMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TasksCard_task
        ...TaskListItem_task
        ...TaskDetailsPage_task
      }
    }
  }
`;

interface TaskListItemProps {
  taskKey: TaskListItem_task$key;
  sectionState?: TaskState;
  canDrag?: boolean;
  isDragging?: boolean;
  isGhost?: boolean;
  ignoreClick: boolean;
  onDragStart?: () => void;
  onDragOver?: (event: React.DragEvent) => void;
  onDrop?: () => void;
  onDragEnd?: () => void;
  onStateChange?: () => void;
}

export function TaskListItem(props: TaskListItemProps) {
  const organizationId = useOrganizationId();
  const { canUpdate, ...task } = useFragment(taskListItemFragment, props.taskKey);
  const [updateState, isUpdating] = useMutation<TaskListItemUpdateStateMutation>(
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
      onCompleted: () => {
        props.onStateChange?.();
      },
    }).catch(() => {
      // Error feedback is handled by useMutation.
    });
  }

  return (
    <SharedTaskListItem
      task={{
        id: task.id,
        name: task.name,
        state: task.state,
        priority: task.priority,
        timeEstimate: task.timeEstimate,
        deadline: task.deadline,
        recurrenceInterval: task.recurrenceInterval,
        externalLink: task.externalLink,
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
      detailsTo={taskDetailsPath(organizationId, task.id)}
      assigneeTo={task.assignedTo
        ? `/organizations/${organizationId}/settings/users/${task.assignedTo.id}`
        : undefined}
      sectionState={props.sectionState}
      canDrag={props.canDrag}
      isDragging={props.isDragging}
      isGhost={props.isGhost}
      ignoreClick={props.ignoreClick}
      onStateChange={onStateChange}
      onDragStart={props.onDragStart}
      onDragOver={props.onDragOver}
      onDrop={props.onDrop}
      onDragEnd={props.onDragEnd}
    />
  );
}
