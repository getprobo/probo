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
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { TaskFindingsSection_task$key } from "#/__generated__/core/TaskFindingsSection_task.graphql";
import type { TaskFindingsSectionList_task$key } from "#/__generated__/core/TaskFindingsSectionList_task.graphql";
import type { TaskFindingsSectionListQuery } from "#/__generated__/core/TaskFindingsSectionListQuery.graphql";
import type { TaskFindingsSectionListRefetchQuery } from "#/__generated__/core/TaskFindingsSectionListRefetchQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { findingDetailsPath } from "#/pages/organizations/findings/_lib/findingPath";

import { taskFindings } from "../variants";

const taskFindingsPageSize = 50;

const taskFindingsSectionFragment = graphql`
  fragment TaskFindingsSection_task on Task {
    id
    canListFindings: permission(action: "core:finding:list")
  }
`;

const taskFindingsListQuery = graphql`
  query TaskFindingsSectionListQuery($taskId: ID!) {
    node(id: $taskId) {
      __typename
      ... on Task {
        ...TaskFindingsSectionList_task
      }
    }
  }
`;

const taskFindingsListFragment = graphql`
  fragment TaskFindingsSectionList_task on Task
  @refetchable(queryName: "TaskFindingsSectionListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    id
    findings(first: $first, after: $after)
      @connection(key: "TaskFindingsSection_findings") {
      edges {
        node {
          id
          title
          referenceId
        }
      }
    }
  }
`;

interface TaskFindingsSectionProps {
  taskKey: TaskFindingsSection_task$key;
}

export function TaskFindingsSection({ taskKey }: TaskFindingsSectionProps) {
  const task = useFragment(taskFindingsSectionFragment, taskKey);
  if (!task.canListFindings) {
    return null;
  }

  return (
    <Suspense fallback={null}>
      <TaskFindingsLoaded taskId={task.id} />
    </Suspense>
  );
}

function TaskFindingsLoaded({ taskId }: { taskId: string }) {
  const data = useLazyLoadQuery<TaskFindingsSectionListQuery>(
    taskFindingsListQuery,
    { taskId },
  );
  const taskKey = data.node?.__typename === "Task" ? data.node : null;
  if (taskKey == null) {
    return null;
  }

  return <TaskFindingsList taskKey={taskKey} />;
}

function TaskFindingsList({ taskKey }: { taskKey: TaskFindingsSectionList_task$key }) {
  const { t } = useTranslation("organizations/tasks");
  const organizationId = useOrganizationId();
  const {
    data: task,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    TaskFindingsSectionListRefetchQuery,
    TaskFindingsSectionList_task$key
  >(taskFindingsListFragment, taskKey);
  const { root, list, item, title, more } = taskFindings();
  const findings = task.findings.edges.map(edge => edge.node);

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <Heading level={2} size={2} weight="medium">
          {t("detailsPage.findings")}
        </Heading>
        <div className={list()}>
          {findings.length > 0
            ? findings.map(finding => (
                <div key={finding.id} className={item()}>
                  <Link
                    size={1}
                    title={finding.title}
                    className={title()}
                    to={findingDetailsPath(organizationId, finding.id)}
                  >
                    {finding.title}
                  </Link>
                  <Text size={1} color="faint">{finding.referenceId}</Text>
                </div>
              ))
            : (
                <Text size={1} color="faint">{t("detailsPage.noFindings")}</Text>
              )}
          {hasNext && (
            <Button
              variant="ghost"
              color="neutral"
              size={1}
              className={more()}
              loading={isLoadingNext}
              onClick={() => loadNext(taskFindingsPageSize)}
            >
              {t("detailsPage.actions.loadMoreFindings")}
            </Button>
          )}
        </div>
      </div>
    </Card>
  );
}
