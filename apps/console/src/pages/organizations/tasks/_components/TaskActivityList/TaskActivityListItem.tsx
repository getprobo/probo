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
import { Card } from "@probo/ui/src/v2/Card/Card";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { TimelineContent } from "@probo/ui/src/v2/Timeline/TimelineContent";
import { TimelineItem } from "@probo/ui/src/v2/Timeline/TimelineItem";
import { TimelineMarker } from "@probo/ui/src/v2/Timeline/TimelineMarker";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type {
  TaskActivityListItem_activity$data,
  TaskActivityListItem_activity$key,
} from "#/__generated__/core/TaskActivityListItem_activity.graphql";
import { RichEditor } from "#/pages/organizations/_components/RichEditor";

import {
  formatTaskActivityValue,
  taskActivitySentenceKey,
} from "../../_lib/taskActivity";
import { taskActivityComment, taskCommentEditor } from "../../variants";
import { TaskCommentDeleteDialog } from "../TaskCommentsList/TaskCommentDeleteDialog";
import { TaskCommentEditor } from "../TaskCommentsList/TaskCommentEditor";

const taskActivityListItemFragment = graphql`
  fragment TaskActivityListItem_activity on TaskActivity {
    activityType
    field
    oldValue
    newValue
    createdAt
    actor {
      fullName
      emailAddress
      avatar {
        downloadUrl
      }
    }
    comment {
      content
      canUpdate: permission(action: "core:task-comment:update")
      canDelete: permission(action: "core:task-comment:delete")
      ...TaskCommentEditor_taskComment
      ...TaskCommentDeleteDialog_taskComment
    }
  }
`;

interface TaskActivityListItemProps {
  taskActivityKey: TaskActivityListItem_activity$key;
}

export function TaskActivityListItem({ taskActivityKey }: TaskActivityListItemProps) {
  const { i18n } = useTranslation("organizations/tasks");
  const activity = useFragment(taskActivityListItemFragment, taskActivityKey);
  const occurredAt = dateTimeFormat(i18n.language, activity.createdAt);

  return (
    <TimelineItem>
      <TimelineMarker className="!bg-transparent">
        <Avatar
          name={activity.actor?.fullName}
          email={activity.actor?.emailAddress}
          src={activity.actor?.avatar?.downloadUrl}
          size={1}
          radius="full"
          fallback={activity.actor == null ? <UserIcon /> : undefined}
        />
      </TimelineMarker>
      {activity.activityType === "COMMENTED"
        ? activity.comment == null
          ? (
              <TimelineContent>
                <Text size={1} color="faint" className="ml-auto">
                  {occurredAt}
                </Text>
              </TimelineContent>
            )
          : (
              <CommentStep
                comment={activity.comment}
                occurredAt={occurredAt}
              />
            )
        : (
            <ActivityStep
              activity={activity}
              occurredAt={occurredAt}
            />
          )}
    </TimelineItem>
  );
}

function ActivityStep({
  activity,
  occurredAt,
}: {
  activity: TaskActivityListItem_activity$data;
  occurredAt: string;
}) {
  const { i18n, t } = useTranslation("organizations/tasks");
  const actorName = activity.actor?.fullName
    ?? t("detailsPage.activity.actorFallback");
  const oldValue = formatTaskActivityValue(
    activity.field,
    activity.oldValue,
    i18n.language,
    t,
  );
  const newValue = formatTaskActivityValue(
    activity.field,
    activity.newValue,
    i18n.language,
    t,
  );
  const sentence = t(
    taskActivitySentenceKey(
      activity.activityType,
      activity.field,
      oldValue,
      newValue,
    ),
    {
      actor: actorName,
      oldValue,
      newValue,
    },
  );

  return (
    <TimelineContent>
      <Text size={2}>{sentence}</Text>
      <Text size={1} color="faint">
        {occurredAt}
      </Text>
    </TimelineContent>
  );
}

function CommentStep({
  comment,
  occurredAt,
}: {
  comment: NonNullable<TaskActivityListItem_activity$data["comment"]>;
  occurredAt: string;
}) {
  const { t } = useTranslation("organizations/tasks");
  const [deleteOpen, setDeleteOpen] = useState(false);
  const { content, card, header } = taskActivityComment();

  return (
    <TimelineContent className={content()}>
      <Card variant="soft" size={2} className={card()}>
        <div className={header()}>
          <Text size={1} color="faint">
            {occurredAt}
          </Text>
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
      </Card>
    </TimelineContent>
  );
}
