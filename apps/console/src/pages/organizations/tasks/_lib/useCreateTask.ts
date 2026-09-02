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

import { useTranslation } from "react-i18next";
import { graphql, useRelayEnvironment } from "react-relay";

import type { useCreateTaskMutation } from "#/__generated__/core/useCreateTaskMutation.graphql";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { insertTaskEdgeSorted } from "./taskConnectionOrder";
import {
  measureTasksConnectionKey,
  organizationTasksConnectionKey,
  taskConnectionId,
} from "./taskPath";
import type { TaskPriority, TaskState } from "./taskState";

const createTaskMutation = graphql`
  mutation useCreateTaskMutation($input: CreateTaskInput!) {
    createTask(input: $input) {
      taskEdge {
        node {
          ...TasksCard_task
          ...TasksCard_TaskRowFragment
        }
      }
    }
  }
`;

export function useCreateTask() {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const relayEnv = useRelayEnvironment();
  const [commit, isCreating] = useMutation<useCreateTaskMutation>(
    createTaskMutation,
    {
      successMessage: t("createDialog.messages.created"),
      errorToast: t("createDialog.errors.create"),
    },
  );

  async function createTask(
    input: {
      name: string;
      description?: string | null;
      state?: TaskState;
      priority: TaskPriority;
      measureId?: string | null;
    },
    connectionId: string,
  ) {
    const measureId = input.measureId ?? undefined;
    const connections = [...new Set([
      connectionId,
      taskConnectionId(organizationId, organizationTasksConnectionKey),
      ...(measureId
        ? [taskConnectionId(measureId, measureTasksConnectionKey)]
        : []),
    ])];

    await commit({
      variables: {
        input: {
          organizationId,
          name: input.name,
          description: input.description || null,
          state: input.state,
          priority: input.priority,
          measureId,
        },
      },
      updater: store => insertTaskEdgeSorted(store, connections),
    });

    if (measureId) {
      updateStoreCounter(relayEnv, measureId, "tasks(first:0)", 1);
    }
  }

  return [createTask, isCreating] as const;
}
