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

import { TaskListItem as SharedTaskListItem, type TaskPriority, type TaskState } from "@probo/tasks";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskListItem_task$key } from "#/__generated__/core/TaskListItem_task.graphql";
import type { TaskListItemUpdateStateMutation } from "#/__generated__/core/TaskListItemUpdateStateMutation.graphql";
import { UserSelect } from "#/components/form/UserSelect";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { taskDetailsPath } from "#/pages/organizations/tasks/_lib/taskPath";

const noneValue = "__NONE__";

const taskListItemFragment = graphql`
  fragment TaskListItem_task on Task {
    id
    name
    state
    priority
    createdAt
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
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const { canUpdate, ...task } = useFragment(taskListItemFragment, props.taskKey);
  const [updateState, isUpdating] = useMutation<TaskListItemUpdateStateMutation>(
    updateStateMutation,
  );
  const assignedTo = task.assignedTo;

  function updateTask(input: {
    state?: TaskState;
    assignedToId?: string | null;
    priority?: TaskPriority;
    deadline?: string | null;
    timeEstimate?: string | null;
    recurrenceInterval?: string | null;
  }) {
    void updateState({
      variables: {
        input: {
          taskId: task.id,
          ...input,
        },
      },
      onCompleted: () => {
        if (input.state != null) {
          props.onStateChange?.();
        }
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
        createdAt: task.createdAt,
        timeEstimate: task.timeEstimate,
        deadline: task.deadline,
        recurrenceInterval: task.recurrenceInterval,
        externalLink: task.externalLink,
        assignedTo: assignedTo == null
          ? null
          : {
              id: assignedTo.id,
              fullName: assignedTo.fullName,
              emailAddress: assignedTo.emailAddress,
              avatarUrl: assignedTo.avatar?.downloadUrl,
            },
      }}
      canUpdate={canUpdate}
      canUpdateAssignee={canUpdate}
      isUpdating={isUpdating}
      detailsTo={taskDetailsPath(organizationId, task.id)}
      renderAssignee={canUpdate
        ? avatar => (
            <UserSelect
              trigger={avatar}
              value={assignedTo?.id ?? null}
              disabled={isUpdating || props.ignoreClick}
              onValueChange={(assignedToId) => {
                updateTask({ assignedToId });
              }}
              emptyLabel={t("detailsPage.unassigned")}
              emptyValue={noneValue}
              ariaLabel={t("detailsPage.fields.assignedTo")}
              size={1}
              contractEnded={false}
              pinned={assignedTo == null
                ? null
                : {
                    id: assignedTo.id,
                    fullName: assignedTo.fullName,
                    emailAddress: assignedTo.emailAddress,
                    avatarUrl: assignedTo.avatar?.downloadUrl,
                  }}
            />
          )
        : undefined}
      sectionState={props.sectionState}
      canDrag={props.canDrag}
      isDragging={props.isDragging}
      isGhost={props.isGhost}
      ignoreClick={props.ignoreClick}
      onStateChange={(state) => {
        updateTask({ state });
      }}
      onPriorityChange={(priority) => {
        updateTask({ priority });
      }}
      onDeadlineChange={(deadline) => {
        updateTask({
          deadline,
          ...(deadline == null && task.recurrenceInterval != null
            ? { recurrenceInterval: null }
            : {}),
        });
      }}
      onTimeEstimateChange={(timeEstimate) => {
        updateTask({ timeEstimate });
      }}
      onDragStart={props.onDragStart}
      onDragOver={props.onDragOver}
      onDrop={props.onDrop}
      onDragEnd={props.onDragEnd}
    />
  );
}
