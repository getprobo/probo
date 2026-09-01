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
import { dateTimeFormat } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskCommentListItem_comment$key } from "#/__generated__/core/TaskCommentListItem_comment.graphql";

import { taskCommentListItem } from "../variants";

import { TaskCommentDeleteDialog } from "./TaskCommentDeleteDialog";

const taskCommentListItemFragment = graphql`
  fragment TaskCommentListItem_comment on TaskComment {
    description
    createdAt
    owner {
      fullName
    }
    canDelete: permission(action: "core:task-comment:delete")
    ...TaskCommentDeleteDialog_comment
  }
`;

interface TaskCommentListItemProps {
  taskCommentKey: TaskCommentListItem_comment$key;
}

export function TaskCommentListItem({ taskCommentKey }: TaskCommentListItemProps) {
  const { i18n, t } = useTranslation("organizations/tasks");
  const comment = useFragment(taskCommentListItemFragment, taskCommentKey);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const { header, description } = taskCommentListItem();
  const ownerInitial = comment.owner.fullName.charAt(0).toUpperCase();

  return (
    <ListItem>
      <Avatar
        size={2}
        variant="soft"
        color="gold"
        radius="full"
        fallback={ownerInitial || <UserIcon />}
      />
      <ListItemContent>
        <div className={header()}>
          <Text size={2} weight="medium">
            {comment.owner.fullName}
          </Text>
          <Text size={1} color="faint">
            {dateTimeFormat(i18n.language, comment.createdAt)}
          </Text>
        </div>
        <div className={description()}>
          <Text size={2}>{comment.description}</Text>
        </div>
      </ListItemContent>
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
    </ListItem>
  );
}
