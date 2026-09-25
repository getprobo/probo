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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Anchor } from "@probo/ui/src/v2/Link/Anchor";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ErrorInfo, type ReactNode, Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useLazyLoadQuery } from "react-relay";

import type { TaskLinearField_task$key } from "#/__generated__/core/TaskLinearField_task.graphql";
import type { TaskLinearFieldQuery } from "#/__generated__/core/TaskLinearFieldQuery.graphql";
import type { TaskLinearFieldUnlinkMutation } from "#/__generated__/core/TaskLinearFieldUnlinkMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { taskLinearField, taskPropertiesSection } from "../variants";

import { TaskLinearPublishField } from "./TaskLinearPublishField";

const taskLinearFieldFragment = graphql`
  fragment TaskLinearField_task on Task {
    id
    canUpdate: permission(action: "core:task:update")
    externalLink {
      identifier
      url
    }
    ...TaskLinearPublishField_task
  }
`;

const taskLinearFieldQuery = graphql`
  query TaskLinearFieldQuery($organizationId: ID!) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        connectors(filter: { providers: [LINEAR_SYNC] }) {
          id
        }
      }
    }
  }
`;

const unlinkMutation = graphql`
  mutation TaskLinearFieldUnlinkMutation($input: UnlinkTaskExternalInput!) {
    unlinkTaskExternal(input: $input) {
      task {
        ...TaskLinearField_task
      }
    }
  }
`;

interface TaskLinearFieldProps {
  taskKey: TaskLinearField_task$key;
  fallback?: ReactNode;
  onError?: (error: unknown, info: ErrorInfo) => void;
}

function TaskLinearFieldContent({ taskKey }: { taskKey: TaskLinearField_task$key }) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const task = useFragment(taskLinearFieldFragment, taskKey);
  const data = useLazyLoadQuery<TaskLinearFieldQuery>(
    taskLinearFieldQuery,
    { organizationId },
  );
  const [unlink, isUnlinking] = useMutation<TaskLinearFieldUnlinkMutation>(
    unlinkMutation,
    {
      successMessage: t("detailsPage.linear.unlinked"),
      errorToast: t("detailsPage.linear.errors.unlink"),
    },
  );

  const organization = data.node?.__typename === "Organization" ? data.node : null;
  const connected = (organization?.connectors.length ?? 0) > 0;
  const { row } = taskPropertiesSection();
  const { link } = taskLinearField();

  if (!connected) {
    return null;
  }

  let value = (
    <Text size={2} color="faint">{t("detailsPage.empty")}</Text>
  );

  if (task.externalLink) {
    value = (
      <div className={link()}>
        <Anchor size={2} href={task.externalLink.url} target="_blank" rel="noreferrer">
          {task.externalLink.identifier}
        </Anchor>
        {task.canUpdate
          ? (
              <Button
                size={1}
                variant="ghost"
                disabled={isUnlinking}
                onClick={() => {
                  void unlink({
                    variables: { input: { taskId: task.id } },
                  });
                }}
              >
                {t("detailsPage.linear.unlink")}
              </Button>
            )
          : null}
      </div>
    );
  } else if (task.canUpdate) {
    value = (
      <Suspense fallback={<SelectSkeleton size={1} className="w-full" />}>
        <TaskLinearPublishField taskKey={task} />
      </Suspense>
    );
  }

  return (
    <div className={row()}>
      <Text size={2} color="faint">{t("detailsPage.fields.linear")}</Text>
      {value}
    </div>
  );
}

export function TaskLinearField({
  taskKey,
  fallback,
  onError,
}: TaskLinearFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const { row } = taskPropertiesSection();

  return (
    <ErrorBoundary
      fallback={fallback ?? (
        <div className={row()}>
          <Text size={2} color="faint">{t("detailsPage.fields.linear")}</Text>
          <Text size={2} color="faint">{t("detailsPage.linear.errors.load")}</Text>
        </div>
      )}
      onError={onError}
    >
      <TaskLinearFieldContent taskKey={taskKey} />
    </ErrorBoundary>
  );
}
