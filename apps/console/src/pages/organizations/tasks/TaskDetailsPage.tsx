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

import { CaretLeftIcon, TrashIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { TaskDetails } from "@probo/tasks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { useLocation, useNavigate } from "react-router";

import type { TaskDetailsPage_task$key } from "#/__generated__/core/TaskDetailsPage_task.graphql";
import type { TaskDetailsPageQuery } from "#/__generated__/core/TaskDetailsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { TaskDeleteDialog } from "./_components/TaskDeleteDialog";
import { TaskDescriptionSection } from "./_components/TaskDescriptionSection";
import { TaskEngagementSection } from "./_components/TaskEngagementSection";
import { TaskInternalControlsSection } from "./_components/TaskInternalControlsSection";
import { TaskLinearField } from "./_components/TaskLinearField";
import { TaskNameField } from "./_components/TaskNameField";
import { TaskPropertiesSection } from "./_components/TaskPropertiesSection";
import { taskListPath } from "./_lib/taskPath";

export const taskDetailsPageFragment = graphql`
  fragment TaskDetailsPage_task on Task {
    id
    name
    canDelete: permission(action: "core:task:delete")
    organization {
      id
    }
    ...TaskNameField_task
    ...TaskDescriptionSection_task
    ...TaskEngagementSection_task
    ...TaskPropertiesSection_task
    ...TaskInternalControlsSection_task
    ...TaskLinearField_task
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
  const { t } = useTranslation("organizations/tasks");
  const navigate = useNavigate();
  const location = useLocation();
  const organizationId = useOrganizationId();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const taskKey = data.node?.__typename === "Task" ? data.node : null;
  const task = useFragment<TaskDetailsPage_task$key>(
    taskDetailsPageFragment,
    taskKey,
  );
  usePageTitle(task?.name ?? "");

  if (!task || task.organization.id !== organizationId) {
    throw new NotFoundError(t("detailsPage.notFound"));
  }

  return (
    <>
      <TaskDetails
        back={(
          <Link
            to={{ pathname: taskListPath(organizationId), search: location.search }}
            size={2}
            color="neutral"
            underline={false}
            iconStart={<CaretLeftIcon />}
            className="self-start"
          >
            {t("detailsPage.back")}
          </Link>
        )}
        name={<TaskNameField taskKey={task} />}
        actions={task.canDelete
          ? (
              <Button
                color="red"
                iconStart={<TrashIcon />}
                onClick={() => setDeleteOpen(true)}
              >
                {t("detailsPage.actions.delete")}
              </Button>
            )
          : undefined}
        description={<TaskDescriptionSection taskKey={task} />}
        engagement={<TaskEngagementSection taskKey={task} />}
        properties={(
          <div className="flex flex-col gap-4">
            <TaskPropertiesSection taskKey={task} />
            <TaskLinearField taskKey={task} />
            <TaskInternalControlsSection taskKey={task} />
          </div>
        )}
      />
      {task.canDelete && (
        <TaskDeleteDialog
          taskId={task.id}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
          onDeleted={() => {
            void navigate({
              pathname: taskListPath(organizationId),
              search: location.search,
            });
          }}
        />
      )}
    </>
  );
}
