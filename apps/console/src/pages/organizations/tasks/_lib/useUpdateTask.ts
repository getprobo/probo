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
import { type UseMutationConfig, useRelayEnvironment } from "react-relay";
import { graphql } from "relay-runtime";

import type { useUpdateTaskMutation } from "#/__generated__/core/useUpdateTaskMutation.graphql";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { moveTaskNodeSorted } from "./taskConnectionOrder";
import {
  measureTasksConnectionKey,
  organizationTasksConnectionKey,
  taskConnectionId,
} from "./taskPath";

function linkedRecordId(record: unknown, field: string) {
  if (record == null || typeof record !== "object") {
    return undefined;
  }

  const value = (record as Record<string, unknown>)[field];
  if (value == null || typeof value !== "object" || !("__ref" in value)) {
    return undefined;
  }

  return typeof value.__ref === "string" ? value.__ref : undefined;
}

const updateTaskMutation = graphql`
  mutation useUpdateTaskMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TaskDetailsPage_task
        ...TasksCard_task
        ...TasksCard_TaskRowFragment
      }
    }
  }
`;

export function useUpdateTask() {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const relayEnv = useRelayEnvironment();
  const [commit, isUpdating] = useMutation<useUpdateTaskMutation>(
    updateTaskMutation,
    {
      successMessage: t("detailsPage.messages.updated"),
      errorToast: t("detailsPage.errors.update"),
    },
  );

  function updateTask(config: UseMutationConfig<useUpdateTaskMutation>) {
    const previousMeasureId = linkedRecordId(
      relayEnv.getStore().getSource().get(config.variables.input.taskId),
      "measure",
    );
    const inputMeasureId = config.variables.input.measureId;
    const measureChanged = inputMeasureId !== undefined;
    const nextMeasureId = measureChanged
      ? inputMeasureId ?? undefined
      : previousMeasureId;

    return commit({
      ...config,
      updater: (store, data) => {
        const node = store.getRootField("updateTask")?.getLinkedRecord("task");
        if (node) {
          moveTaskNodeSorted(store, node, {
            organizationConnectionId: taskConnectionId(
              organizationId,
              organizationTasksConnectionKey,
            ),
            previousMeasureConnectionId: measureChanged && previousMeasureId
              ? taskConnectionId(previousMeasureId, measureTasksConnectionKey)
              : undefined,
            nextMeasureConnectionId: nextMeasureId
              ? taskConnectionId(nextMeasureId, measureTasksConnectionKey)
              : undefined,
            createIfMissing: measureChanged,
          });
        }
        config.updater?.(store, data);
      },
    }).then((result) => {
      if (measureChanged && previousMeasureId !== nextMeasureId) {
        if (previousMeasureId) {
          updateStoreCounter(relayEnv, previousMeasureId, "tasks(first:0)", -1);
        }
        if (nextMeasureId) {
          updateStoreCounter(relayEnv, nextMeasureId, "tasks(first:0)", 1);
        }
      }
      return result;
    });
  }

  return [updateTask, isUpdating] as const;
}
