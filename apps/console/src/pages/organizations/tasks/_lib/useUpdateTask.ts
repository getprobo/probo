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
import { type MutationFeedback, useMutation } from "#/lib/relay/useMutation";

import { moveTaskNodeSorted, syncTaskInternalControlIds } from "./taskConnectionOrder";
import {
  internalControlTasksConnectionKey,
  organizationTasksConnectionKey,
  taskConnectionId,
} from "./taskPath";

const updateTaskMutation = graphql`
  mutation useUpdateTaskMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TaskDetailsPage_task
        ...TasksCard_task
        ...TaskListItem_task
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

  function updateTask(
    config: UseMutationConfig<useUpdateTaskMutation> & {
      previousInternalControlIds?: readonly string[];
    },
    feedback?: MutationFeedback,
  ) {
    const { previousInternalControlIds = [], ...relayConfig } = config;
    const inputInternalControlIds = relayConfig.variables.input.internalControlIds;
    const internalControlsChanged = inputInternalControlIds !== undefined;
    const nextInternalControlIds = internalControlsChanged
      ? inputInternalControlIds ?? []
      : previousInternalControlIds;
    const previousSet = new Set(previousInternalControlIds);
    const nextSet = new Set(nextInternalControlIds);

    return commit({
      ...relayConfig,
      updater: (store, data) => {
        const payload = store.getRootField("updateTask");
        const node = payload?.getLinkedRecord("task");
        if (node) {
          moveTaskNodeSorted(store, node, {
            organizationConnectionId: taskConnectionId(
              organizationId,
              organizationTasksConnectionKey,
            ),
            previousInternalControlConnectionIds: internalControlsChanged
              ? previousInternalControlIds.map(internalControlId =>
                  taskConnectionId(internalControlId, internalControlTasksConnectionKey),
                )
              : undefined,
            nextInternalControlConnectionIds: nextInternalControlIds.map(internalControlId =>
              taskConnectionId(internalControlId, internalControlTasksConnectionKey),
            ),
            createIfMissing: internalControlsChanged,
          });
          if (internalControlsChanged) {
            syncTaskInternalControlIds(
              store,
              node,
              previousInternalControlIds,
              nextInternalControlIds,
            );
          }
        }
        relayConfig.updater?.(store, data);
      },
    }, feedback).then((result) => {
      if (internalControlsChanged) {
        for (const internalControlId of previousSet) {
          if (!nextSet.has(internalControlId)) {
            updateStoreCounter(relayEnv, internalControlId, "tasks(first:0)", -1);
          }
        }
        for (const internalControlId of nextSet) {
          if (!previousSet.has(internalControlId)) {
            updateStoreCounter(relayEnv, internalControlId, "tasks(first:0)", 1);
          }
        }
      }
      return result;
    });
  }

  return [updateTask, isUpdating] as const;
}
