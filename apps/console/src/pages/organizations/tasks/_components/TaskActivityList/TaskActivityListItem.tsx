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

import { PencilSimpleIcon, TrashIcon, UserIcon } from "@phosphor-icons/react";
import { relativeDateFormat } from "@probo/i18n";
import {
  formatTaskActivityValue,
  taskActivityComment,
  taskActivitySentenceKey,
  taskCommentEditor,
  TASKS_NAMESPACE,
} from "@probo/tasks";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Field } from "@probo/ui/src/v2/form/Field";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { TimelineContent } from "@probo/ui/src/v2/Timeline/TimelineContent";
import { TimelineItem } from "@probo/ui/src/v2/Timeline/TimelineItem";
import { TimelineMarker } from "@probo/ui/src/v2/Timeline/TimelineMarker";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode, useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type {
  TaskActivityListItem_activity$data,
  TaskActivityListItem_activity$key,
} from "#/__generated__/core/TaskActivityListItem_activity.graphql";
import { RichEditor } from "#/pages/organizations/_components/RichEditor";
import {
  isRichEditorContentEmpty,
  richEditorContentTextLength,
} from "#/pages/organizations/_lib/richEditorContent";
import { useMentionMembers } from "#/pages/organizations/_lib/useMentionMembers";

import { taskCommentMaxLength } from "../../_lib/taskCommentMaxLength";
import { useUpdateTaskComment } from "../../_lib/useUpdateTaskComment";
import { TaskCommentDeleteDialog } from "../TaskCommentsList/TaskCommentDeleteDialog";

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
      id
      content
      canUpdate: permission(action: "core:task-comment:update")
      canDelete: permission(action: "core:task-comment:delete")
      ...TaskCommentDeleteDialog_taskComment
    }
  }
`;

interface TaskActivityListItemProps {
  taskActivityKey: TaskActivityListItem_activity$key;
}

export function TaskActivityListItem({ taskActivityKey }: TaskActivityListItemProps) {
  const { i18n, t } = useTranslation("organizations/tasks");
  const activity = useFragment(taskActivityListItemFragment, taskActivityKey);
  const occurredAt = relativeDateFormat(i18n.language, activity.createdAt)
    || t("detailsPage.activity.justNow");

  const actor = (className?: string) => (
    <Avatar
      className={className}
      name={activity.actor?.fullName}
      email={activity.actor?.emailAddress}
      src={activity.actor?.avatar?.downloadUrl}
      size={1}
      radius="full"
      fallback={activity.actor == null ? <UserIcon /> : undefined}
    />
  );
  const isComment = activity.activityType === "COMMENTED";
  const { item } = taskActivityComment();

  return (
    <TimelineItem
      data-activity={isComment ? "comment" : undefined}
      className={isComment ? item() : "items-center!"}
    >
      {isComment && activity.comment != null
        ? (
            <CommentStep
              actor={actor()}
              authorName={activity.actor?.fullName}
              createdAt={activity.createdAt}
              comment={activity.comment}
            />
          )
        : (
            <>
              <TimelineMarker className="bg-transparent!">
                {actor("!size-4 [&_svg]:!size-3 [&>span]:!text-[0.625rem] [&>span]:!leading-none")}
              </TimelineMarker>
              {isComment
                ? (
                    <TimelineContent>
                      <Text size={1} color="faint" className="ml-auto">
                        {occurredAt}
                      </Text>
                    </TimelineContent>
                  )
                : (
                    <ActivityStep
                      activity={activity}
                      occurredAt={occurredAt}
                    />
                  )}
            </>
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
  const { t: tTasks } = useTranslation(TASKS_NAMESPACE);
  const actorName = activity.actor?.fullName
    ?? t("detailsPage.activity.actorFallback");
  const oldValue = formatTaskActivityValue(
    activity.field,
    activity.oldValue,
    i18n.language,
    tTasks,
  );
  const newValue = formatTaskActivityValue(
    activity.field,
    activity.newValue,
    i18n.language,
    tTasks,
  );
  const sentence = tTasks(
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
    <TimelineContent className="pt-0!">
      <Text size={1} className="min-w-0 flex-1">
        {sentence}
      </Text>
      <Text size={1} color="faint" className="shrink-0 whitespace-nowrap">
        {occurredAt}
      </Text>
    </TimelineContent>
  );
}

function CommentStep({
  actor,
  authorName,
  createdAt,
  comment,
}: {
  actor: ReactNode;
  authorName: string | null | undefined;
  createdAt: string;
  comment: NonNullable<TaskActivityListItem_activity$data["comment"]>;
}) {
  const { i18n, t } = useTranslation("organizations/tasks");
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(comment.content);
  const [error, setError] = useState<string>();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [updateTaskComment, isSaving] = useUpdateTaskComment();
  const mentionSource = useMentionMembers();
  const [uploading, setUploading] = useState(false);
  const { card, row, body, header, identity, actions, save } = taskActivityComment();
  const name = authorName ?? t("detailsPage.activity.actorFallback");
  const relativeTime = relativeDateFormat(i18n.language, createdAt)
    || t("detailsPage.activity.justNow");
  const showActions = comment.canUpdate || comment.canDelete;

  function startEditing() {
    setDraft(comment.content);
    setError(undefined);
    setEditing(true);
  }

  async function saveComment() {
    if (uploading) {
      return;
    }

    if (isRichEditorContentEmpty(draft)) {
      setError(t("detailsPage.comments.errors.contentRequired"));
      return;
    }

    if (richEditorContentTextLength(draft) > taskCommentMaxLength) {
      setError(t("detailsPage.comments.errors.contentTooLong", {
        max: taskCommentMaxLength,
      }));
      return;
    }

    try {
      await updateTaskComment({
        variables: {
          input: {
            taskCommentId: comment.id,
            content: draft,
          },
        },
      });
      setEditing(false);
      setError(undefined);
    } catch {
      // The mutation already shows an error toast.
    }
  }

  return (
    <Card variant="soft" padding={1} className={card()}>
      <div className={row()}>
        {actor}
        <div className={body()}>
          <div className={header()}>
            <Trans
              ns="organizations/tasks"
              i18nKey="detailsPage.activity.commented"
              parent="div"
              className={identity()}
              values={{ name, time: relativeTime }}
              components={{
                name: <Text size={1} weight="medium" className="truncate" />,
                time: <Text size={1} color="faint" className="shrink-0 whitespace-nowrap" />,
              }}
            />
            {showActions && (
              <div className={actions()}>
                {comment.canUpdate && !editing && (
                  <IconButton
                    size={1}
                    variant="ghost"
                    color="neutral"
                    aria-label={t("detailsPage.comments.actions.edit")}
                    onClick={startEditing}
                  >
                    <PencilSimpleIcon />
                  </IconButton>
                )}
                {comment.canDelete && (
                  <IconButton
                    size={1}
                    variant="ghost"
                    color="red"
                    aria-label={t("detailsPage.comments.actions.delete")}
                    onClick={() => {
                      setDeleteOpen(true);
                    }}
                  >
                    <TrashIcon />
                  </IconButton>
                )}
              </div>
            )}
          </div>
          {comment.canDelete && (
            <TaskCommentDeleteDialog
              taskCommentKey={comment}
              open={deleteOpen}
              onOpenChange={setDeleteOpen}
            />
          )}
          <ErrorBoundary
            fallback={(
              <Text size={2} color="faint">
                {t("detailsPage.comments.errors.content")}
              </Text>
            )}
          >
            <Field error={editing ? error : undefined}>
              <RichEditor
                key={editing ? "edit" : comment.content}
                className={taskCommentEditor({ editing })}
                content={editing ? draft : comment.content}
                controls={false}
                disabled={!editing}
                aria-label={t("detailsPage.comments.fields.comment")}
                mentionSource={mentionSource}
                onChangeContent={editing
                  ? (next) => {
                      setDraft(next);
                      setError(undefined);
                    }
                  : undefined}
                onUploadingChange={editing ? setUploading : undefined}
              />
            </Field>
          </ErrorBoundary>
          {editing && (
            <div className={save()}>
              <Button
                size={1}
                variant="ghost"
                color="neutral"
                disabled={isSaving}
                onClick={() => {
                  setEditing(false);
                  setError(undefined);
                }}
              >
                {t("detailsPage.comments.actions.cancel")}
              </Button>
              <Button
                size={1}
                variant="solid"
                color="neutral"
                highContrast
                loading={isSaving}
                disabled={uploading}
                onClick={() => {
                  void saveComment();
                }}
              >
                {t("detailsPage.comments.actions.save")}
              </Button>
            </div>
          )}
        </div>
      </div>
    </Card>
  );
}
