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
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskListAsanaDialog_task$key } from "#/__generated__/core/TaskListAsanaDialog_task.graphql";

import { taskAsanaField } from "../variants";

import { TaskAsanaPublishField } from "./TaskAsanaPublishField";

const taskListAsanaDialogFragment = graphql`
  fragment TaskListAsanaDialog_task on Task {
    ...TaskAsanaPublishField_task
  }
`;

export interface TaskListAsanaDialogProps {
  taskKey: TaskListAsanaDialog_task$key;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function TaskListAsanaDialog({
  taskKey,
  open,
  onOpenChange,
}: TaskListAsanaDialogProps) {
  const { t } = useTranslation("organizations/tasks");
  const task = useFragment(taskListAsanaDialogFragment, taskKey);
  const { skeleton } = taskAsanaField();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("detailsPage.asana.linkHint")}</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <Suspense fallback={<SelectSkeleton size={1} className={skeleton()} />}>
            <TaskAsanaPublishField
              taskKey={task}
              spacing="relaxed"
              onCompleted={() => {
                onOpenChange(false);
              }}
            />
          </Suspense>
        </DialogBody>
      </DialogPopup>
    </Dialog>
  );
}
