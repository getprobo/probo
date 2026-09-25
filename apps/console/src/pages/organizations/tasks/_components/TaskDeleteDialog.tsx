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

import { TrashIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { useNavigate } from "react-router";

import type { TaskDeleteDialog_task$key } from "#/__generated__/core/TaskDeleteDialog_task.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { taskListPath } from "../_lib/taskPath";
import { useDeleteTask } from "../_lib/useDeleteTask";

const taskDeleteDialogFragment = graphql`
  fragment TaskDeleteDialog_task on Task {
    id
    measure {
      id
    }
  }
`;

interface TaskDeleteDialogProps {
  taskKey: TaskDeleteDialog_task$key;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function TaskDeleteDialog({
  taskKey,
  open,
  onOpenChange,
}: TaskDeleteDialogProps) {
  const { t } = useTranslation("organizations/tasks");
  const navigate = useNavigate();
  const organizationId = useOrganizationId();
  const task = useFragment(taskDeleteDialogFragment, taskKey);
  const [deleteTask, isDeleting] = useDeleteTask();

  function handleDelete() {
    void deleteTask(task.id, task.measure?.id ?? undefined).then(
      () => {
        onOpenChange(false);
        void navigate(taskListPath(organizationId));
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("detailsPage.delete.title")}</DialogTitle>
          <DialogDescription>
            {t("detailsPage.delete.description")}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("detailsPage.actions.cancel")}
              </Button>
            )}
          />
          <Button
            type="button"
            variant="solid"
            color="red"
            iconStart={<TrashIcon />}
            loading={isDeleting}
            onClick={handleDelete}
          >
            {t("detailsPage.delete.confirm")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
