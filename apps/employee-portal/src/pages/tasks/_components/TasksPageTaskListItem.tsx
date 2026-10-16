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

import { TaskListItem, type TaskPriority, type TaskState } from "@probo/tasks";
import { memo, useState } from "react";
import { graphql, useFragment, useRelayEnvironment } from "react-relay";

import type { TasksPageTaskListItem_task$key } from "#/__generated__/core/TasksPageTaskListItem_task.graphql";
import type { TasksPageTaskListItemUpdateStateMutation } from "#/__generated__/core/TasksPageTaskListItemUpdateStateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

import { resortTaskListEdge, taskListStateMoveUpdaters } from "../_lib/moveTaskListEdge";
import { useTasksFilters } from "../_lib/useTasksFilters";

import { TasksPageTaskDeleteDialog } from "./TasksPageTaskDeleteDialog";

const tasksPageTaskListItemFragment = graphql`
  fragment TasksPageTaskListItem_task on Task {
    id
    name
    state
    priority
    rank
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
  }
`;

const updateStateMutation = graphql`
  mutation TasksPageTaskListItemUpdateStateMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TasksPageList_task
        ...TasksPageTaskListItem_task
      }
    }
  }
`;

export interface TasksPageTaskListItemProps {
  taskKey: TasksPageTaskListItem_task$key;
  sectionState?: TaskState;
}

export const TasksPageTaskListItem = memo(function TasksPageTaskListItem({
  taskKey,
  sectionState,
}: TasksPageTaskListItemProps) {
  const { organizationId } = useEmployeePortalRoute();
  const { graphqlQuery, deadlineBounds } = useTasksFilters();
  const environment = useRelayEnvironment();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const task = useFragment(tasksPageTaskListItemFragment, taskKey);
  const { canUpdate, canDelete } = task;
  const [updateState, isUpdating] = useMutation<TasksPageTaskListItemUpdateStateMutation>(
    updateStateMutation,
  );

  function updateTask(input: {
    state?: TaskState;
    priority?: TaskPriority;
    deadline?: string | null;
    timeEstimate?: string | null;
    recurrenceInterval?: string | null;
  }) {
    const queries = graphqlQuery == null ? [null] : [graphqlQuery, null];
    const move = input.state == null
      ? null
      : taskListStateMoveUpdaters(environment, {
          taskId: task.id,
          nextState: input.state,
          organizationId,
          query: graphqlQuery,
          deadline: deadlineBounds,
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
            resortTaskListEdge(store, node, {
              state: task.state,
              organizationId,
              queries,
              deadline: deadlineBounds,
            });
          }
        }
      },
      updater(store) {
        move?.updater(store);
        if (input.state == null && input.priority != null) {
          const node = store.get(task.id);
          if (node != null) {
            resortTaskListEdge(store, node, {
              state: task.state,
              organizationId,
              queries,
              deadline: deadlineBounds,
            });
          }
        }
      },
    }).catch(() => {
      // Error feedback is handled by useMutation.
    });
  }

  return (
    <>
      <TaskListItem
        task={{
          id: task.id,
          name: task.name,
          state: task.state,
          priority: task.priority,
          createdAt: task.createdAt,
          timeEstimate: task.timeEstimate,
          deadline: task.deadline,
          recurrenceInterval: task.recurrenceInterval,
          externalLink: task.externalLink == null
            ? null
            : {
                identifier: task.externalLink.identifier,
                url: task.externalLink.url,
              },
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
        sectionState={sectionState}
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
        onDelete={canDelete
          ? () => {
              setDeleteOpen(true);
            }
          : undefined}
      />
      {canDelete && (
        <TasksPageTaskDeleteDialog
          taskId={task.id}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
        />
      )}
    </>
  );
});
