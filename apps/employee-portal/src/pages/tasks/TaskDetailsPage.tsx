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

import { isOverdueTask, TaskDetails } from "@probo/tasks";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { useLocation } from "react-router";

import type { TaskDetailsPage_task$key } from "#/__generated__/core/TaskDetailsPage_task.graphql";
import type { TaskDetailsPageQuery } from "#/__generated__/core/TaskDetailsPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { PageHeader } from "#/pages/_components/PageHeader";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

import { TaskDescriptionSection } from "./_components/TaskDescriptionSection";
import { TaskEngagementSection } from "./_components/TaskEngagementSection";
import { TaskNameField } from "./_components/TaskNameField";
import { TaskPropertiesSection } from "./_components/TaskPropertiesSection";

export const taskDetailsPageFragment = graphql`
  fragment TaskDetailsPage_task on Task {
    name
    state
    deadline
    ...TaskNameField_task
    ...TaskDescriptionSection_task
    ...TaskEngagementSection_task
    ...TaskPropertiesSection_task
  }
`;

export const taskDetailsPageQuery = graphql`
  query TaskDetailsPageQuery($taskId: ID!) {
    node(id: $taskId) {
      __typename
      ... on Task {
        ...TaskDetailsPage_task
      }
    }
  }
`;

interface TaskDetailsPageProps {
  queryRef: PreloadedQuery<TaskDetailsPageQuery>;
}

export function TaskDetailsPage({ queryRef }: TaskDetailsPageProps) {
  const data = usePreloadedQuery<TaskDetailsPageQuery>(
    taskDetailsPageQuery,
    queryRef,
  );
  const { t } = useTranslation("tasks");
  const { t: tApp } = useTranslation();
  const { employeePortalId } = useEmployeePortalRoute();
  const { search } = useLocation();
  const taskKey = data.node?.__typename === "Task" ? data.node : null;
  const task = useFragment<TaskDetailsPage_task$key>(
    taskDetailsPageFragment,
    taskKey,
  );

  if (!task) {
    throw new NotFoundError(t("detail.notFound"));
  }

  const listTo = `/${employeePortalId}/tasks${search}`;

  return (
    <main className="mx-auto flex w-full max-w-5xl flex-col gap-10 px-8 pt-8 pb-32">
      <PageHeader
        homeLabel={tApp("homePage.breadcrumb")}
        parent={{
          label: t("breadcrumb"),
          to: listTo,
        }}
        currentLabel={task.name}
        hideTitle
      />
      <TaskDetails
        name={<TaskNameField taskKey={task} />}
        overdue={isOverdueTask(task.deadline, task.state)}
        description={<TaskDescriptionSection taskKey={task} />}
        engagement={<TaskEngagementSection taskKey={task} />}
        properties={<TaskPropertiesSection taskKey={task} />}
      />
    </main>
  );
}
