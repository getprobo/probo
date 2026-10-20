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

import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskListLinearDialog_task$key } from "#/__generated__/core/TaskListLinearDialog_task.graphql";

import { TaskLinearPublishField } from "./TaskLinearPublishField";

const taskListLinearDialogFragment = graphql`
  fragment TaskListLinearDialog_task on Task {
    ...TaskLinearPublishField_task
  }
`;

export interface TaskListLinearDialogProps {
  taskKey: TaskListLinearDialog_task$key;
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
  const task = useFragment(taskListLinearDialogFragment, taskKey);

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
