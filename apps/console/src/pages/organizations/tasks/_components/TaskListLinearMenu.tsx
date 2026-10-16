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
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownSubmenu } from "@probo/ui/src/v2/Dropdown/DropdownSubmenu";
import { DropdownSubmenuTrigger } from "@probo/ui/src/v2/Dropdown/DropdownSubmenuTrigger";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskListLinearMenu_task$key } from "#/__generated__/core/TaskListLinearMenu_task.graphql";
import type { TaskListLinearMenuUnlinkMutation } from "#/__generated__/core/TaskListLinearMenuUnlinkMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { TaskLinearPublishField } from "./TaskLinearPublishField";

const taskListLinearMenuFragment = graphql`
  fragment TaskListLinearMenu_task on Task {
    id
    ...TaskLinearPublishField_task
  }
`;

const unlinkMutation = graphql`
  mutation TaskListLinearMenuUnlinkMutation($input: UnlinkTaskExternalInput!) {
    unlinkTaskExternal(input: $input) {
      task {
        ...TaskLinearField_task
        ...TaskListItem_task
      }
    }
  }
`;

export interface TaskListLinearMenuProps {
  taskKey: TaskListLinearMenu_task$key;
  linked: boolean;
}

export function TaskListLinearMenu({
  taskKey,
  linked,
}: TaskListLinearMenuProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskListLinearMenuFragment, taskKey);
  const [unlink, isUnlinking] = useMutation<TaskListLinearMenuUnlinkMutation>(
    unlinkMutation,
    {
      successMessage: t("detailsPage.linear.unlinked"),
      errorToast: t("detailsPage.linear.errors.unlink"),
    },
  );

  if (linked) {
    return (
      <DropdownItem
        color="linear"
        disabled={isUnlinking}
        iconStart={<Linear color="currentColor" aria-hidden />}
        onClick={() => {
          void unlink({
            variables: { input: { taskId: task.id } },
          });
        }}
      >
        {t("detailsPage.linear.unlink")}
      </DropdownItem>
    );
  }

  return (
    <DropdownSubmenu>
      <DropdownSubmenuTrigger
        color="linear"
        iconStart={<Linear color="currentColor" aria-hidden />}
      >
        {t("detailsPage.linear.linkTo")}
      </DropdownSubmenuTrigger>
      <DropdownPopup size={1} align="start" side="right">
        <div className="w-64 p-2">
          <Suspense fallback={<SelectSkeleton size={1} className="w-full" />}>
            <TaskLinearPublishField taskKey={task} />
          </Suspense>
        </div>
      </DropdownPopup>
    </DropdownSubmenu>
  );
}
