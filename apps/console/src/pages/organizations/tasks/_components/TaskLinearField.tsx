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

import { Linear } from "@probo/ui/src/Atoms/ThirdParties/Linear";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ErrorInfo, type ReactNode, Suspense } from "react";
import { Trans, useTranslation } from "react-i18next";
import { graphql, useFragment, useLazyLoadQuery } from "react-relay";

import type { TaskLinearField_task$key } from "#/__generated__/core/TaskLinearField_task.graphql";
import type { TaskLinearFieldQuery } from "#/__generated__/core/TaskLinearFieldQuery.graphql";
import type { TaskLinearFieldUnlinkMutation } from "#/__generated__/core/TaskLinearFieldUnlinkMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { taskLinearField } from "../variants";

import { TaskLinearPublishField } from "./TaskLinearPublishField";

const taskLinearFieldFragment = graphql`
  fragment TaskLinearField_task on Task {
    id
    canUpdate: permission(action: "core:task:update")
    externalLink {
      provider
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
        ...TaskListItem_task
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
  const { root, title, titleIcon, link, issue } = taskLinearField();

  if (!connected || task.externalLink?.provider === "ASANA") {
    return null;
  }

  let value = (
    <Text size={1} color="faint">{t("detailsPage.empty")}</Text>
  );

  if (task.externalLink) {
    const externalLink = task.externalLink;
    value = (
      <div className={link()}>
        <Text size={2}>
          <Trans
            ns="organizations/tasks"
            i18nKey="detailsPage.linear.linkedIssue"
            values={{ issue: externalLink.identifier }}
            components={{
              issue: (
                <a
                  className={issue()}
                  href={externalLink.url}
                  target="_blank"
                  rel="noreferrer"
                />
              ),
            }}
          />
        </Text>
        {task.canUpdate
          ? (
              <Button
                className="w-full"
                color="linear"
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
    <Card variant="soft" size={2}>
      <div className={root()}>
        <div className={title()}>
          <Linear className={titleIcon()} aria-hidden />
          <Heading level={2} size={2} weight="medium">
            {task.externalLink
              ? t("detailsPage.fields.linear")
              : t("detailsPage.linear.linkHint")}
          </Heading>
        </div>
        {value}
      </div>
    </Card>
  );
}

export function TaskLinearField({
  taskKey,
  fallback,
  onError,
}: TaskLinearFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const { root, title, titleIcon } = taskLinearField();

  return (
    <ErrorBoundary
      fallback={fallback ?? (
        <Card variant="soft" size={2}>
          <div className={root()}>
            <div className={title()}>
              <Linear className={titleIcon()} aria-hidden />
              <Heading level={2} size={2} weight="medium">
                {t("detailsPage.fields.linear")}
              </Heading>
            </div>
            <Text size={1} color="faint">{t("detailsPage.linear.errors.load")}</Text>
          </div>
        </Card>
      )}
      onError={onError}
    >
      <TaskLinearFieldContent taskKey={taskKey} />
    </ErrorBoundary>
  );
}
