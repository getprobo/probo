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
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
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
  onLink?: () => void;
}

export function TaskListLinearMenu({
  taskKey,
  linked,
  onLink,
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
    <DropdownItem
      color="linear"
      iconStart={<Linear color="currentColor" aria-hidden />}
      onClick={onLink}
    >
      {t("detailsPage.linear.linkTo")}
    </DropdownItem>
  );
}

export interface TaskListLinearDialogProps {
  taskKey: TaskListLinearMenu_task$key;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

// Rendered beside the row menu. The menu popup unmounts when it closes, so the
// form cannot live inside the "Link to" item.
export function TaskListLinearDialog({
  taskKey,
  open,
  onOpenChange,
}: TaskListLinearDialogProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskListLinearMenuFragment, taskKey);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup className="max-w-96!">
        <DialogHeader>
          <DialogTitle>{t("detailsPage.linear.linkHint")}</DialogTitle>
        </DialogHeader>
        <div className="px-8">
          <Suspense fallback={<SelectSkeleton size={1} className="w-full" />}>
            <TaskLinearPublishField
              taskKey={task}
              spacing="relaxed"
              onCompleted={() => {
                onOpenChange(false);
              }}
            />
          </Suspense>
        </div>
      </DialogPopup>
    </Dialog>
  );
}
