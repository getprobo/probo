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

import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ErrorInfo, type ReactNode, useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskDescriptionSection_task$key } from "#/__generated__/core/TaskDescriptionSection_task.graphql";
import { RichDescriptionEditor } from "#/pages/organizations/_components/RichDescriptionEditor";

import { useUpdateTask } from "../_lib/useUpdateTask";

const taskDescriptionSectionFragment = graphql`
  fragment TaskDescriptionSection_task on Task {
    id
    content
    canUpdate: permission(action: "core:task:update")
  }
`;

interface TaskDescriptionSectionProps {
  taskKey: TaskDescriptionSection_task$key;
  fallback?: ReactNode;
  onError?: (error: unknown, info: ErrorInfo) => void;
}

function TaskDescriptionSectionContent({
  taskKey,
}: {
  taskKey: TaskDescriptionSection_task$key;
}) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskDescriptionSectionFragment, taskKey);
  const [updateTask] = useUpdateTask();
  const save = useCallback(
    async (content: string | null) => {
      await updateTask({
        variables: {
          input: {
            taskId: task.id,
            content,
          },
        },
      });
    },
    [task.id, updateTask],
  );

  return (
    <RichDescriptionEditor
      saved={task.content}
      canUpdate={task.canUpdate}
      ariaLabel={t("detailsPage.fields.description")}
      emptyLabel={t("detailsPage.noDescription")}
      save={save}
    />
  );
}

export function TaskDescriptionSection({
  taskKey,
  fallback,
  onError,
}: TaskDescriptionSectionProps) {
  const { t } = useTranslation("organizations/tasks");

  return (
    <ErrorBoundary
      fallback={fallback ?? (
        <Text size={2} color="faint">
          {t("detailsPage.errors.content")}
        </Text>
      )}
      onError={onError}
    >
      <TaskDescriptionSectionContent taskKey={taskKey} />
    </ErrorBoundary>
  );
}
