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
import { graphql, type UseMutationConfig, useRelayEnvironment } from "react-relay";

import type { useUpdateTaskMutation } from "#/__generated__/core/useUpdateTaskMutation.graphql";
import { type MutationFeedback, useMutation } from "#/lib/relay/useMutation";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

import { taskListStateMoveUpdaters } from "./moveTaskListEdge";
import { useTasksFilters } from "./useTasksFilters";

const updateTaskMutation = graphql`
  mutation useUpdateTaskMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TaskDetailsPage_task
        ...TasksPageList_task
        ...TasksPageTaskListItem_task
      }
    }
  }
`;

export function useUpdateTask() {
  const { t } = useTranslation("tasks");
  const { organizationId } = useEmployeePortalRoute();
  const { graphqlQuery } = useTasksFilters();
  const environment = useRelayEnvironment();
  const [commit, isUpdating] = useMutation<useUpdateTaskMutation>(
    updateTaskMutation,
    {
      successMessage: t("detail.messages.updated"),
      errorToast: t("detail.errors.update"),
    },
  );

  function updateTask(
    config: UseMutationConfig<useUpdateTaskMutation>,
    feedback?: MutationFeedback,
  ) {
    const move = taskListStateMoveUpdaters(environment, {
      taskId: config.variables.input.taskId,
      nextState: config.variables.input.state,
      organizationId,
      query: graphqlQuery,
    });

    return commit({
      ...config,
      optimisticUpdater(store, data) {
        move.optimisticUpdater(store);
        config.optimisticUpdater?.(store, data);
      },
      updater(store, data) {
        move.updater(store);
        config.updater?.(store, data);
      },
    }, feedback);
  }

  return [updateTask, isUpdating] as const;
}
