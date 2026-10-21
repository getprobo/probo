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

import { type TaskPriority, type TaskState } from "@probo/tasks";
import { useCallback } from "react";
import { useRelayEnvironment } from "react-relay";
import { graphql } from "relay-runtime";

import type { useUpdateTaskMutation } from "#/__generated__/core/useUpdateTaskMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

import {
  reconcileTaskListDeadline,
  resortTaskListEdge,
  taskListStateMoveUpdaters,
} from "./moveTaskListEdge";
import { useTasksFilters } from "./useTasksFilters";

const updateTaskMutation = graphql`
  mutation useUpdateTaskMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TaskDetailsPage_task
      }
    }
  }
`;

export interface UpdateTaskInput {
  name?: string;
  content?: string | null;
  state?: TaskState;
  priority?: TaskPriority;
  deadline?: string | null;
  timeEstimate?: string | null;
  recurrenceInterval?: string | null;
}

export function useUpdateTask(task: {
  id: string;
  state: TaskState;
  deadline?: string | null;
  recurrenceInterval?: string | null;
}) {
  const { organizationId } = useEmployeePortalRoute();
  const { graphqlQuery, deadlineBounds } = useTasksFilters();
  const environment = useRelayEnvironment();
  const [commit, isUpdating] = useMutation<useUpdateTaskMutation>(updateTaskMutation);

  const updateTask = useCallback((input: UpdateTaskInput) => {
    const queries = graphqlQuery == null ? [null] : [graphqlQuery, null];
    const previousDeadline = task.deadline ?? null;
    function applyDeadline(store: Parameters<typeof reconcileTaskListDeadline>[0]) {
      if (input.deadline === undefined) {
        return;
      }
      const node = store.get(task.id);
      if (node == null) {
        return;
      }
      node.setValue(input.deadline, "deadline");
      reconcileTaskListDeadline(store, node, {
        state: task.state,
        organizationId,
        queries,
        deadline: deadlineBounds,
        previousDeadline,
      });
    }
    const move = input.state == null
      ? null
      : taskListStateMoveUpdaters(environment, {
          taskId: task.id,
          nextState: input.state,
          organizationId,
          query: graphqlQuery,
          deadline: deadlineBounds,
        });

    return commit({
      variables: {
        input: {
          taskId: task.id,
          ...input,
        },
      },
      optimisticUpdater(store) {
        applyDeadline(store);
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
        applyDeadline(store);
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
    }, { successMessage: "" });
  }, [
    commit,
    deadlineBounds,
    environment,
    graphqlQuery,
    organizationId,
    task.deadline,
    task.id,
    task.state,
  ]);

  return [updateTask, isUpdating] as const;
}
