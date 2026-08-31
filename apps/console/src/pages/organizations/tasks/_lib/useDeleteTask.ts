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

import type { useDeleteTaskMutation } from "#/__generated__/core/useDeleteTaskMutation.graphql";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import {
  measureTasksConnectionKey,
  organizationTasksConnectionKey,
  taskConnectionId,
} from "./taskPath";

const deleteTaskMutation = graphql`
  mutation useDeleteTaskMutation(
    $input: DeleteTaskInput!
    $connections: [ID!]!
  ) {
    deleteTask(input: $input) {
      deletedTaskId @deleteEdge(connections: $connections)
    }
  }
`;

export function useDeleteTask() {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const relayEnv = useRelayEnvironment();
  const [commit, isDeleting] = useMutation<useDeleteTaskMutation>(
    deleteTaskMutation,
    {
      successMessage: t("detailsPage.messages.deleted"),
      errorToast: t("detailsPage.errors.delete"),
    },
  );

  async function deleteTask(taskId: string, measureId?: string) {
    const connections = [
      taskConnectionId(organizationId, organizationTasksConnectionKey),
      ...(measureId
        ? [taskConnectionId(measureId, measureTasksConnectionKey)]
        : []),
    ];

    await commit({
      variables: {
        input: { taskId },
        connections,
      },
    });

    if (measureId) {
      updateStoreCounter(relayEnv, measureId, "tasks(first:0)", -1);
    }
  }

  return [deleteTask, isDeleting] as const;
}
