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
import { useTranslation } from "react-i18next";
import { graphql, useRelayEnvironment } from "react-relay";

import type { useCreateTaskMutation } from "#/__generated__/core/useCreateTaskMutation.graphql";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { adjustTotalCount } from "./moveOrganizationTaskListEdge";
import { insertTaskEdgeSorted } from "./taskConnectionOrder";
import {
  internalControlTasksConnectionKey,
  organizationTaskListConnectionId,
  organizationTaskListFilterVariants,
  taskConnectionId,
} from "./taskPath";
import { taskMatchesFilter, useTasksCardFilters } from "./useTasksCardFilters";

const createTaskMutation = graphql`
  mutation useCreateTaskMutation($input: CreateTaskInput!) {
    createTask(input: $input) {
      taskEdge {
        node {
          id
          ...TasksPageList_task
          ...TaskListItem_task
        }
      }
    }
  }
`;

export function useCreateTask() {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const relayEnv = useRelayEnvironment();
  const { graphqlFilter } = useTasksCardFilters();
  const [commit, isCreating] = useMutation<useCreateTaskMutation>(
    createTaskMutation,
    {
      successMessage: t("newPage.messages.created"),
      errorToast: t("newPage.errors.create"),
    },
  );

  async function createTask(
    input: {
      name: string;
      content?: string | null;
      state?: TaskState;
      priority: TaskPriority;
      internalControlIds?: readonly string[] | null;
      linearTeamId?: string | null;
      assignedToId?: string | null;
      timeEstimate?: string | null;
      deadline?: string | null;
      recurrenceInterval?: string | null;
      asanaProjectId?: string | null;
    },
    connectionId?: string,
  ) {
    const internalControlIds = [...new Set(input.internalControlIds ?? [])];
    const state = input.state ?? "TODO";
    const listFilter = {
      query: graphqlFilter.query,
      assignedToId: graphqlFilter.assignedToId,
      internalControlId: graphqlFilter.internalControlId,
      deadlineOnOrBefore: graphqlFilter.deadlineOnOrBefore,
      deadlineOnOrAfter: graphqlFilter.deadlineOnOrAfter,
      hasDeadline: graphqlFilter.hasDeadline,
      hasAssignee: graphqlFilter.hasAssignee,
    };
    const connections = new Set<string>();
    for (const variant of organizationTaskListFilterVariants(listFilter)) {
      if (taskMatchesFilter(
        {
          name: input.name,
          state,
          internalControlIds,
          assignedToId: input.assignedToId ?? null,
          deadline: input.deadline ?? null,
        },
        { ...variant, state },
      )) {
        connections.add(organizationTaskListConnectionId(organizationId, variant, state));
      }
    }
    for (const internalControlId of internalControlIds) {
      connections.add(taskConnectionId(internalControlId, internalControlTasksConnectionKey));
    }
    if (connectionId != null) {
      connections.add(connectionId);
    }

    const payload = await commit({
      variables: {
        input: {
          organizationId,
          name: input.name,
          content: input.content || null,
          state: input.state,
          priority: input.priority,
          internalControlIds: internalControlIds.length > 0 ? [...internalControlIds] : undefined,
          ...(input.linearTeamId !== undefined
            ? { linearTeamId: input.linearTeamId }
            : {}),
          ...(input.assignedToId != null ? { assignedToId: input.assignedToId } : {}),
          ...(input.timeEstimate != null ? { timeEstimate: input.timeEstimate } : {}),
          ...(input.deadline != null ? { deadline: input.deadline } : {}),
          ...(input.recurrenceInterval != null
            ? { recurrenceInterval: input.recurrenceInterval }
            : {}),
          ...(input.asanaProjectId !== undefined
            ? { asanaProjectId: input.asanaProjectId }
            : {}),
        },
      },
      updater(store) {
        const ids = [...connections];
        insertTaskEdgeSorted(store, ids);
        const createdId = store.getRootField("createTask")
          ?.getLinkedRecord("taskEdge")
          ?.getLinkedRecord("node")
          ?.getDataID();
        if (createdId == null) {
          return;
        }
        for (const id of ids) {
          const connection = store.get(id);
          const present = connection?.getLinkedRecords("edges")?.some(
            edge => edge?.getLinkedRecord("node")?.getDataID() === createdId,
          );
          if (connection != null && present) {
            adjustTotalCount(connection, 1);
          }
        }
      },
    });

    for (const internalControlId of internalControlIds) {
      updateStoreCounter(relayEnv, internalControlId, "tasks(first:0)", 1);
    }

    return payload.createTask.taskEdge.node.id;
  }

  return [createTask, isCreating] as const;
}
