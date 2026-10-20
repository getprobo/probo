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

import { CheckSquareIcon } from "@phosphor-icons/react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TasksDashboardCard_viewer$key } from "#/__generated__/core/TasksDashboardCard_viewer.graphql";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

import { DashboardCardFrame } from "./DashboardCardFrame";
import { dashboardCard } from "./variants";

const tasksDashboardCardFragment = graphql`
  fragment TasksDashboardCard_viewer on Viewer
  @argumentDefinitions(organizationId: { type: "ID!" })
  @throwOnFieldError {
    inProgressTasks: assignedTasks(
      organizationId: $organizationId
      filter: { state: IN_PROGRESS }
    ) {
      totalCount
    }
    todoTasks: assignedTasks(
      organizationId: $organizationId
      filter: { state: TODO }
    ) {
      totalCount
    }
    backlogTasks: assignedTasks(
      organizationId: $organizationId
      filter: { state: BACKLOG }
    ) {
      totalCount
    }
    doneTasks: assignedTasks(
      organizationId: $organizationId
      filter: { state: DONE }
    ) {
      totalCount
    }
  }
`;

export interface TasksDashboardCardProps {
  viewerKey: TasksDashboardCard_viewer$key;
  wash?: boolean;
}

export function TasksDashboardCard({
  viewerKey,
  wash = false,
}: TasksDashboardCardProps) {
  const { t } = useTranslation();
  const { employeePortalId } = useEmployeePortalRoute();
  const slots = dashboardCard({ wash });
  const viewer = useFragment(tasksDashboardCardFragment, viewerKey);

  const pendingCount
    = viewer.inProgressTasks.totalCount
      + viewer.todoTasks.totalCount
      + viewer.backlogTasks.totalCount;
  const completedCount = viewer.doneTasks.totalCount;

  let bodyState: "empty" | "pending" | "allDone" = "empty";
  if (pendingCount > 0) {
    bodyState = "pending";
  } else if (completedCount > 0) {
    bodyState = "allDone";
  }

  return (
    <DashboardCardFrame
      wash={wash}
      icon={<CheckSquareIcon className={slots.icon()} />}
      title={t("homePage.dashboard.tasks.title")}
      description={t("homePage.dashboard.tasks.description")}
      to={`/${employeePortalId}/tasks`}
      bodyState={bodyState}
      pendingLabel={t("homePage.dashboard.tasks.pendingCount", { count: pendingCount })}
      emptyLabel={t("homePage.dashboard.tasks.empty")}
      actionLabel={t("homePage.dashboard.tasks.action")}
    />
  );
}
