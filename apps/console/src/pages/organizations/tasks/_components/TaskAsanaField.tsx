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

import { Asana } from "@probo/ui/src/Atoms/ThirdParties/Asana";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ErrorInfo, type ReactNode, Suspense } from "react";
import { Trans, useTranslation } from "react-i18next";
import { graphql, useFragment, useLazyLoadQuery } from "react-relay";

import type { TaskAsanaField_task$key } from "#/__generated__/core/TaskAsanaField_task.graphql";
import type { TaskAsanaFieldQuery } from "#/__generated__/core/TaskAsanaFieldQuery.graphql";
import type { TaskAsanaFieldUnlinkMutation } from "#/__generated__/core/TaskAsanaFieldUnlinkMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { taskAsanaField } from "../variants";

import { TaskAsanaPublishField } from "./TaskAsanaPublishField";

const taskAsanaFieldFragment = graphql`
  fragment TaskAsanaField_task on Task {
    id
    canUpdate: permission(action: "core:task:update")
    externalLink {
      provider
      url
    }
    ...TaskAsanaPublishField_task
  }
`;

const taskAsanaFieldQuery = graphql`
  query TaskAsanaFieldQuery($organizationId: ID!) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        connectors(filter: { providers: [ASANA] }) {
          id
        }
      }
    }
  }
`;

const unlinkMutation = graphql`
  mutation TaskAsanaFieldUnlinkMutation($input: UnlinkTaskExternalInput!) {
    unlinkTaskExternal(input: $input) {
      task {
        ...TaskAsanaField_task
        ...TaskListItem_task
      }
    }
  }
`;

interface TaskAsanaFieldProps {
  taskKey: TaskAsanaField_task$key;
  fallback?: ReactNode;
  onError?: (error: unknown, info: ErrorInfo) => void;
}

function TaskAsanaFieldContent({ taskKey }: { taskKey: TaskAsanaField_task$key }) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const task = useFragment(taskAsanaFieldFragment, taskKey);
  const data = useLazyLoadQuery<TaskAsanaFieldQuery>(
    taskAsanaFieldQuery,
    { organizationId },
  );
  const [unlink, isUnlinking] = useMutation<TaskAsanaFieldUnlinkMutation>(
    unlinkMutation,
    {
      successMessage: t("detailsPage.asana.unlinked"),
      errorToast: t("detailsPage.asana.errors.unlink"),
    },
  );

  const organization = data.node?.__typename === "Organization" ? data.node : null;
  const connected = (organization?.connectors.length ?? 0) > 0;
  const { root, title, titleIcon, link, issue, action, skeleton } = taskAsanaField();
  const linked = task.externalLink?.provider === "ASANA";

  if (!connected || (task.externalLink != null && !linked)) {
    return null;
  }

  let value = (
    <Text size={1} color="faint">{t("detailsPage.empty")}</Text>
  );

  if (linked && task.externalLink) {
    const externalLink = task.externalLink;
    value = (
      <div className={link()}>
        <Text size={2}>
          <Trans
            ns="organizations/tasks"
            i18nKey="detailsPage.asana.linkedTask"
            components={{
              task: (
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
                className={action()}
                color="neutral"
                disabled={isUnlinking}
                onClick={() => {
                  void unlink({
                    variables: { input: { taskId: task.id } },
                  });
                }}
              >
                {t("detailsPage.asana.unlink")}
              </Button>
            )
          : null}
      </div>
    );
  } else if (task.canUpdate) {
    value = (
      <Suspense fallback={<SelectSkeleton size={1} className={skeleton()} />}>
        <TaskAsanaPublishField taskKey={task} />
      </Suspense>
    );
  }

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <div className={title()}>
          <Asana className={titleIcon()} aria-hidden />
          <Heading level={2} size={2} weight="medium">
            {linked
              ? t("detailsPage.fields.asana")
              : t("detailsPage.asana.linkHint")}
          </Heading>
        </div>
        {value}
      </div>
    </Card>
  );
}

export function TaskAsanaField({
  taskKey,
  fallback,
  onError,
}: TaskAsanaFieldProps) {
  const { t } = useTranslation("organizations/tasks");
  const { root, title, titleIcon } = taskAsanaField();

  return (
    <ErrorBoundary
      fallback={fallback ?? (
        <Card variant="soft" size={2}>
          <div className={root()}>
            <div className={title()}>
              <Asana className={titleIcon()} aria-hidden />
              <Heading level={2} size={2} weight="medium">
                {t("detailsPage.fields.asana")}
              </Heading>
            </div>
            <Text size={1} color="faint">{t("detailsPage.asana.errors.load")}</Text>
          </div>
        </Card>
      )}
      onError={onError}
    >
      <TaskAsanaFieldContent taskKey={taskKey} />
    </ErrorBoundary>
  );
}
