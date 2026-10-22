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
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Suspense, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { TaskDeleteDialogQuery } from "#/__generated__/core/TaskDeleteDialogQuery.graphql";

import { useDeleteTask } from "../_lib/useDeleteTask";
import { useTaskInternalControlIds } from "../_lib/useTaskInternalControlIds";

const taskDeleteDialogQuery = graphql`
  query TaskDeleteDialogQuery($taskId: ID!) {
    node(id: $taskId) {
      __typename
      ... on Task {
        ...useTaskInternalControlIds_task
      }
    }
  }
`;

interface TaskDeleteDialogProps {
  taskId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onDeleted?: () => void;
}

export function TaskDeleteDialog({
  taskId,
  open,
  onOpenChange,
  onDeleted,
}: TaskDeleteDialogProps) {
  const [attempt, setAttempt] = useState(0);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {open && (
        <ErrorBoundary
          key={attempt}
          fallback={(
            <TaskDeleteDialogBody
              failed
              pending={false}
              deleting={false}
              onRetry={() => {
                setAttempt(value => value + 1);
              }}
              onDelete={() => {}}
            />
          )}
        >
          <Suspense fallback={<TaskDeleteDialogFallback />}>
            <TaskDeleteDialogLoaded
              key={attempt}
              taskId={taskId}
              onOpenChange={onOpenChange}
              onDeleted={onDeleted}
            />
          </Suspense>
        </ErrorBoundary>
      )}
    </Dialog>
  );
}

function TaskDeleteDialogLoaded({
  taskId,
  onOpenChange,
  onDeleted,
}: Omit<TaskDeleteDialogProps, "open">) {
  const data = useLazyLoadQuery<TaskDeleteDialogQuery>(
    taskDeleteDialogQuery,
    { taskId },
    { fetchPolicy: "network-only" },
  );
  const task = data.node?.__typename === "Task" ? data.node : null;
  if (task == null) {
    return (
      <TaskDeleteDialogBody
        failed
        pending={false}
        deleting={false}
        onDelete={() => {}}
      />
    );
  }

  return (
    <TaskDeleteDialogReady
      taskKey={task}
      onOpenChange={onOpenChange}
      onDeleted={onDeleted}
    />
  );
}

function TaskDeleteDialogReady({
  taskKey,
  onOpenChange,
  onDeleted,
}: {
  taskKey: Parameters<typeof useTaskInternalControlIds>[0];
  onOpenChange: (open: boolean) => void;
  onDeleted?: () => void;
}) {
  const { taskId, ids, pending, failed, retry } = useTaskInternalControlIds(taskKey);
  const [deleteTask, isDeleting] = useDeleteTask();

  function handleDelete() {
    if (pending || failed) {
      return;
    }

    void deleteTask(taskId, ids).then(
      () => {
        onOpenChange(false);
        onDeleted?.();
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <TaskDeleteDialogBody
      failed={failed}
      pending={pending}
      deleting={isDeleting}
      onRetry={retry}
      onDelete={handleDelete}
    />
  );
}

function TaskDeleteDialogFallback() {
  return (
    <TaskDeleteDialogBody
      failed={false}
      pending
      deleting={false}
      onDelete={() => {}}
    />
  );
}

function TaskDeleteDialogBody({
  failed,
  pending,
  deleting,
  onRetry,
  onDelete,
}: {
  failed: boolean;
  pending: boolean;
  deleting: boolean;
  onRetry?: () => void;
  onDelete: () => void;
}) {
  const { t } = useTranslation("organizations/tasks");

  return (
    <DialogPopup>
      <DialogHeader>
        <DialogTitle>{t("detailsPage.delete.title")}</DialogTitle>
        <DialogDescription>
          {failed
            ? t("detailsPage.delete.loadError")
            : t("detailsPage.delete.description")}
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
        {failed && onRetry != null && (
          <Button type="button" variant="soft" color="neutral" onClick={onRetry}>
            {t("detailsPage.actions.retry")}
          </Button>
        )}
        <Button
          type="button"
          variant="solid"
          color="red"
          iconStart={<TrashIcon />}
          disabled={failed}
          loading={deleting || pending}
          onClick={onDelete}
        >
          {t("detailsPage.delete.confirm")}
        </Button>
      </DialogFooter>
    </DialogPopup>
  );
}
