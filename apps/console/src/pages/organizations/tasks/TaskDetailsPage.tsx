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

import {
  DotsThreeVerticalIcon,
  TrashIcon,
} from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";

import type { TaskDetailsPage_task$key } from "#/__generated__/core/TaskDetailsPage_task.graphql";
import type { TaskDetailsPageQuery } from "#/__generated__/core/TaskDetailsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

import { TaskDeleteDialog } from "./_components/TaskDeleteDialog";
import { TaskDescriptionSection } from "./_components/TaskDescriptionSection";
import { TaskNameField } from "./_components/TaskNameField";
import { TaskPropertiesSection } from "./_components/TaskPropertiesSection";
import { taskDetailsPage } from "./variants";

export const taskDetailsPageFragment = graphql`
  fragment TaskDetailsPage_task on Task {
    name
    canDelete: permission(action: "core:task:delete")
    organization {
      id
    }
    ...TaskNameField_task
    ...TaskDescriptionSection_task
    ...TaskPropertiesSection_task
    ...TaskDeleteDialog_task
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

  const { root, header, titleRow, title, actions, body } = taskDetailsPage();

  return (
    <div className={root()}>
      <div className={header()}>
        <div className={titleRow()}>
          <div className={title()}>
            <TaskNameField taskKey={task} />
          </div>
          {task.canDelete && (
            <div className={actions()}>
              <Dropdown>
                <DropdownTrigger
                  render={(
                    <IconButton
                      variant="soft"
                      color="neutral"
                      aria-label={t("detailsPage.actions.more")}
                    >
                      <DotsThreeVerticalIcon />
                    </IconButton>
                  )}
                />
                <DropdownPopup align="end">
                  <DropdownItem
                    color="error"
                    iconStart={<TrashIcon />}
                    onClick={() => setDeleteOpen(true)}
                  >
                    {t("detailsPage.actions.delete")}
                  </DropdownItem>
                </DropdownPopup>
              </Dropdown>
            </div>
          )}
        </div>
      </div>
      <div className={body()}>
        <TaskDescriptionSection taskKey={task} />
        <TaskPropertiesSection taskKey={task} />
      </div>
      {task.canDelete && (
        <TaskDeleteDialog
          taskKey={task}
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
        />
      )}
    </div>
  );
}
