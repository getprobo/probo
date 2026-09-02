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

import { TrashIcon, UserIcon } from "@phosphor-icons/react";
import { RichEditor } from "@probo/ui";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskCommentListItem_taskComment$key } from "#/__generated__/core/TaskCommentListItem_taskComment.graphql";

import { taskCommentEditor, taskCommentListItem } from "../variants";

import { TaskCommentDeleteDialog } from "./TaskCommentDeleteDialog";
import { TaskCommentEditor } from "./TaskCommentEditor";

const taskCommentListItemFragment = graphql`
  fragment TaskCommentListItem_taskComment on TaskComment {
    content
    createdAt
    owner {
      fullName
    }
    canUpdate: permission(action: "core:task-comment:update")
    canDelete: permission(action: "core:task-comment:delete")
    ...TaskCommentEditor_taskComment
    ...TaskCommentDeleteDialog_taskComment
  }
`;

interface TaskCommentListItemProps {
  taskCommentKey: TaskCommentListItem_taskComment$key;
}

export function TaskCommentListItem({ taskCommentKey }: TaskCommentListItemProps) {
  const { i18n, t } = useTranslation("organizations/tasks");
  const comment = useFragment(taskCommentListItemFragment, taskCommentKey);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const { root, header, meta, content } = taskCommentListItem();
  const ownerInitial = comment.owner.fullName.charAt(0).toUpperCase();
  const createdAt = new Intl.DateTimeFormat(i18n.language, {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(new Date(comment.createdAt));

  return (
    <li className={root()}>
      <div className={header()}>
        <Avatar
          size={2}
          variant="soft"
          color="gold"
          radius="full"
          fallback={ownerInitial || <UserIcon />}
        />
        <div className={meta()}>
          <Text size={2} weight="medium">
            {comment.owner.fullName}
          </Text>
          <Text size={1} color="faint">
            {createdAt}
          </Text>
        </div>
        {comment.canDelete && (
          <>
            <IconButton
              variant="ghost"
              color="neutral"
              aria-label={t("detailsPage.comments.actions.delete")}
              onClick={() => {
                setDeleteOpen(true);
              }}
            >
              <TrashIcon />
            </IconButton>
            <TaskCommentDeleteDialog
              taskCommentKey={comment}
              open={deleteOpen}
              onOpenChange={setDeleteOpen}
            />
          </>
        )}
      </div>
      <div className={content()}>
        <ErrorBoundary
          fallback={(
            <Text size={2} color="faint">
              {t("detailsPage.comments.errors.content")}
            </Text>
          )}
        >
          {comment.canUpdate
            ? <TaskCommentEditor taskCommentKey={comment} />
            : (
                <RichEditor
                  className={taskCommentEditor()}
                  content={comment.content}
                  disabled
                  aria-label={t("detailsPage.comments.fields.comment")}
                />
              )}
        </ErrorBoundary>
      </div>
    </li>
  );
}
