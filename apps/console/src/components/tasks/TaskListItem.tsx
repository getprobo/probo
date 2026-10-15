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
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useRelayEnvironment } from "react-relay";

import type { TaskListItem_task$key } from "#/__generated__/core/TaskListItem_task.graphql";
import type { TaskListItemUpdateStateMutation } from "#/__generated__/core/TaskListItemUpdateStateMutation.graphql";
import { TaskAssigneeMenu } from "#/components/tasks/TaskAssigneeMenu";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { TaskDeleteDialog } from "#/pages/organizations/tasks/_components/TaskDeleteDialog";
import {
  resortOrganizationTaskListEdge,
  taskListStateMoveUpdaters,
} from "#/pages/organizations/tasks/_lib/moveOrganizationTaskListEdge";
import {
  organizationTaskListFilterVariants,
  taskDetailsPath,
} from "#/pages/organizations/tasks/_lib/taskPath";
import { useTasksCardFilters } from "#/pages/organizations/tasks/_lib/useTasksCardFilters";

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
    canDelete: permission(action: "core:task:delete")
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
    ...TaskDeleteDialog_task
  }
`;

const updateStateMutation = graphql`
  mutation TaskListItemUpdateStateMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TasksPageList_task
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
  const environment = useRelayEnvironment();
  const { graphqlFilter } = useTasksCardFilters();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const task = useFragment(taskListItemFragment, props.taskKey);
  const { canUpdate, canDelete } = task;
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
    const listFilter = {
      query: graphqlFilter.query,
      assignedToId: graphqlFilter.assignedToId,
      internalControlId: graphqlFilter.internalControlId,
    };
    const filters = organizationTaskListFilterVariants(listFilter);
    const move = input.state == null
      ? null
      : taskListStateMoveUpdaters(environment, {
          taskId: task.id,
          nextState: input.state,
          organizationId,
          filter: listFilter,
        });
    void updateState({
      variables: {
        input: {
          taskId: task.id,
          ...input,
        },
      },
      optimisticUpdater(store) {
        move?.optimisticUpdater(store);
        if (input.state == null && input.priority != null) {
          const node = store.get(task.id);
          if (node != null) {
            node.setValue(input.priority, "priority");
            resortOrganizationTaskListEdge(
              store,
              node,
              organizationId,
              task.state,
              filters,
            );
          }
        }
      },
      updater(store) {
        move?.updater(store);
        if (input.state == null && input.priority != null) {
          const node = store.get(task.id);
          if (node != null) {
            resortOrganizationTaskListEdge(
              store,
              node,
              organizationId,
              task.state,
              filters,
            );
          }
        }
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

  const assigneeSelectProps = {
    value: assignedTo?.id ?? null,
    disabled: isUpdating || props.ignoreClick,
    onValueChange: (assignedToId: string | null) => {
      updateTask({ assignedToId });
    },
    emptyLabel: t("detailsPage.unassigned"),
    emptyValue: noneValue,
    ariaLabel: t("detailsPage.fields.assignedTo"),
    size: 1 as const,
    contractEnded: false,
    pinned: assignedTo == null
      ? null
      : {
          id: assignedTo.id,
          fullName: assignedTo.fullName,
          emailAddress: assignedTo.emailAddress,
          avatarUrl: assignedTo.avatar?.downloadUrl,
        },
  };

  return (
    <>
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
        renderAssigneeMenu={canUpdate
          ? () => (
              <TaskAssigneeMenu
                value={assigneeSelectProps.value}
                emptyLabel={assigneeSelectProps.emptyLabel}
                emptyValue={assigneeSelectProps.emptyValue}
                disabled={assigneeSelectProps.disabled}
                onValueChange={assigneeSelectProps.onValueChange}
              />
            )
          : undefined}
        onDelete={canDelete
          ? () => {
              setDeleteOpen(true);
            }
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
      {canDelete && (
        <TaskDeleteDialog
          taskKey={task}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
        />
      )}
    </>
  );
}
