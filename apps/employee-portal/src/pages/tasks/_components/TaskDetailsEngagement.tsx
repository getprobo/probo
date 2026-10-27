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

import { Form } from "@base-ui/react/form";
import { PencilSimpleIcon, TrashIcon, UserIcon } from "@phosphor-icons/react";
import { relativeDateFormat } from "@probo/i18n";
import {
  formatTaskActivityValue,
  taskActivityComment,
  taskActivitySentenceKey,
  taskCommentEditor,
  taskCommentForm,
  taskCommentsSection,
  TASKS_NAMESPACE,
} from "@probo/tasks";
import { RichEditor } from "@probo/ui/src/RichEditor/RichEditor";
import { isRichEditorContentEmpty, richEditorContentTextLength } from "@probo/ui/src/RichEditor/richEditorContent";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Field } from "@probo/ui/src/v2/form/Field";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Timeline } from "@probo/ui/src/v2/Timeline/Timeline";
import { TimelineContent } from "@probo/ui/src/v2/Timeline/TimelineContent";
import { TimelineItem } from "@probo/ui/src/v2/Timeline/TimelineItem";
import { TimelineMarker } from "@probo/ui/src/v2/Timeline/TimelineMarker";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode, Suspense, useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { graphql, useFragment, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { TaskDetailsEngagement_task$key } from "#/__generated__/core/TaskDetailsEngagement_task.graphql";
import type {
  TaskDetailsEngagementActivity_activity$data,
  TaskDetailsEngagementActivity_activity$key,
} from "#/__generated__/core/TaskDetailsEngagementActivity_activity.graphql";
import type { TaskDetailsEngagementCreateCommentMutation } from "#/__generated__/core/TaskDetailsEngagementCreateCommentMutation.graphql";
import type { TaskDetailsEngagementDeleteCommentMutation } from "#/__generated__/core/TaskDetailsEngagementDeleteCommentMutation.graphql";
import type { TaskDetailsEngagementQuery } from "#/__generated__/core/TaskDetailsEngagementQuery.graphql";
import type { TaskDetailsEngagementRefetchQuery } from "#/__generated__/core/TaskDetailsEngagementRefetchQuery.graphql";
import type { TaskDetailsEngagementUpdateCommentMutation } from "#/__generated__/core/TaskDetailsEngagementUpdateCommentMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { taskCommentMaxLength } from "../_lib/richText";
import { taskActivitiesConnectionId } from "../_lib/taskPath";
import { useMentionMembers } from "../_lib/useMentionMembers";

const taskDetailsActivityListQuery = graphql`
  query TaskDetailsEngagementQuery($taskId: ID!) {
    node(id: $taskId) {
      __typename
      ... on Task {
        ...TaskDetailsEngagement_task
      }
    }
  }
`;

const taskDetailsActivityListFragment = graphql`
  fragment TaskDetailsEngagement_task on Task
  @refetchable(queryName: "TaskDetailsEngagementRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey" }
  ) {
    id
    canCreateComment: permission(action: "core:task-comment:create")
    activities(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: ASC }
    ) @connection(key: "EmployeeTaskActivityList_activities") {
      edges {
        node {
          id
          ...TaskDetailsEngagementActivity_activity
        }
      }
    }
  }
`;

const createCommentMutation = graphql`
  mutation TaskDetailsEngagementCreateCommentMutation(
    $input: CreateTaskCommentInput!
    $connections: [ID!]!
  ) {
    createTaskComment(input: $input) {
      taskActivityEdge @appendEdge(connections: $connections) {
        node {
          id
          ...TaskDetailsEngagementActivity_activity
        }
      }
    }
  }
`;

const updateCommentMutation = graphql`
  mutation TaskDetailsEngagementUpdateCommentMutation($input: UpdateTaskCommentInput!) {
    updateTaskComment(input: $input) {
      taskComment {
        id
        content
      }
    }
  }
`;

const deleteCommentMutation = graphql`
  mutation TaskDetailsEngagementDeleteCommentMutation(
    $input: DeleteTaskCommentInput!
    $connections: [ID!]!
  ) {
    deleteTaskComment(input: $input) {
      deletedTaskActivityId @deleteEdge(connections: $connections)
    }
  }
`;

const taskDetailsActivityItemFragment = graphql`
  fragment TaskDetailsEngagementActivity_activity on TaskActivity {
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
    }
  }
`;

export function TaskDetailsActivity({ taskId }: { taskId: string }) {
  return (
    <Suspense fallback={null}>
      <TaskDetailsActivityLoaded taskId={taskId} />
    </Suspense>
  );
}

function TaskDetailsActivityLoaded({ taskId }: { taskId: string }) {
  const data = useLazyLoadQuery<TaskDetailsEngagementQuery>(
    taskDetailsActivityListQuery,
    { taskId },
    { fetchPolicy: "store-and-network" },
  );
  const taskKey = data.node?.__typename === "Task" ? data.node : null;
  if (taskKey == null) {
    return null;
  }

  return <TaskDetailsActivityList taskKey={taskKey} />;
}

function TaskDetailsActivityList({
  taskKey,
}: {
  taskKey: TaskDetailsEngagement_task$key;
}) {
  const { t } = useTranslation("tasks");
  const {
    data: task,
    hasNext,
    loadNext,
    isLoadingNext,
  } = usePaginationFragment<
    TaskDetailsEngagementRefetchQuery,
    TaskDetailsEngagement_task$key
  >(taskDetailsActivityListFragment, taskKey);
  const { inset, actions } = taskCommentsSection();
  const activities = task.activities.edges.map(edge => edge.node);

  return (
    <>
      <div className={inset()}>
        {activities.length === 0
          ? (
              <Text size={2} color="faint">{t("details.activityEmpty")}</Text>
            )
          : (
              <Timeline className="[&>li:has(+li[data-activity=comment])]:before:hidden!">
                {activities.map(activity => (
                  <TaskDetailsActivityItem
                    key={activity.id}
                    taskId={task.id}
                    taskActivityKey={activity}
                  />
                ))}
              </Timeline>
            )}
        {hasNext && (
          <div className={actions()}>
            <Button
              variant="ghost"
              color="neutral"
              size={1}
              disabled={isLoadingNext}
              loading={isLoadingNext}
              onClick={() => {
                loadNext(20);
              }}
            >
              {t("details.showMore")}
            </Button>
          </div>
        )}
      </div>
      {task.canCreateComment && <TaskDetailsCommentForm taskId={task.id} />}
    </>
  );
}

function TaskDetailsActivityItem({
  taskId,
  taskActivityKey,
}: {
  taskId: string;
  taskActivityKey: TaskDetailsEngagementActivity_activity$key;
}) {
  const { i18n, t } = useTranslation("tasks");
  const { t: tTasks } = useTranslation(TASKS_NAMESPACE);
  const activity = useFragment(taskDetailsActivityItemFragment, taskActivityKey);
  const occurredAt = relativeDateFormat(i18n.language, activity.createdAt)
    || t("details.justNow");
  const isComment = activity.activityType === "COMMENTED";
  const { item } = taskActivityComment();
  const actor = (
    <Avatar
      className={isComment ? undefined : "!size-4 [&_svg]:!size-3 [&>span]:!text-[0.625rem] [&>span]:!leading-none"}
      name={activity.actor?.fullName}
      email={activity.actor?.emailAddress}
      src={activity.actor?.avatar?.downloadUrl}
      size={1}
      radius="full"
      fallback={activity.actor == null ? <UserIcon /> : undefined}
    />
  );

  if (isComment && activity.comment != null) {
    return (
      <TimelineItem data-activity="comment" className={item()}>
        <CommentStep
          taskId={taskId}
          actor={actor}
          authorName={activity.actor?.fullName}
          createdAt={activity.createdAt}
          comment={activity.comment}
        />
      </TimelineItem>
    );
  }

  const actorName = activity.actor?.fullName ?? t("details.actorFallback");
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
    { actor: actorName, oldValue, newValue },
  );

  return (
    <TimelineItem className="items-center!">
      <TimelineMarker>{actor}</TimelineMarker>
      <TimelineContent className="pt-0!">
        <Text size={1} className="min-w-0 flex-1">{sentence}</Text>
        <Text size={1} color="faint" className="shrink-0 whitespace-nowrap">
          {occurredAt}
        </Text>
      </TimelineContent>
    </TimelineItem>
  );
}

function CommentStep({
  taskId,
  actor,
  authorName,
  createdAt,
  comment,
}: {
  taskId: string;
  actor: ReactNode;
  authorName: string | null | undefined;
  createdAt: string;
  comment: NonNullable<TaskDetailsEngagementActivity_activity$data["comment"]>;
}) {
  const { i18n, t } = useTranslation("tasks");
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(comment.content);
  const [error, setError] = useState<string>();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const mentionSource = useMentionMembers();
  const [updateComment, isSaving] = useMutation<TaskDetailsEngagementUpdateCommentMutation>(
    updateCommentMutation,
    {
      successMessage: t("details.commentUpdated"),
      errorToast: t("details.commentUpdateError"),
    },
  );
  const { card, row, body, header, identity, actions, save } = taskActivityComment();
  const name = authorName ?? t("details.actorFallback");
  const relativeTime = relativeDateFormat(i18n.language, createdAt)
    || t("details.justNow");
  const showActions = comment.canUpdate || comment.canDelete;

  async function saveComment() {
    if (isRichEditorContentEmpty(draft)) {
      setError(t("details.commentRequired"));
      return;
    }
    if (richEditorContentTextLength(draft) > taskCommentMaxLength) {
      setError(t("details.commentTooLong", { max: taskCommentMaxLength }));
      return;
    }

    try {
      await updateComment({
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
              ns="tasks"
              i18nKey="details.commented"
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
                    aria-label={t("details.editComment")}
                    onClick={() => {
                      setDraft(comment.content);
                      setError(undefined);
                      setEditing(true);
                    }}
                  >
                    <PencilSimpleIcon />
                  </IconButton>
                )}
                {comment.canDelete && (
                  <IconButton
                    size={1}
                    variant="ghost"
                    color="red"
                    aria-label={t("details.deleteComment")}
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
            <DeleteCommentDialog
              taskId={taskId}
              commentId={comment.id}
              open={deleteOpen}
              onOpenChange={setDeleteOpen}
            />
          )}
          <ErrorBoundary
            fallback={<Text size={2} color="faint">{t("details.commentContentError")}</Text>}
          >
            <Field error={editing ? error : undefined}>
              <RichEditor
                key={editing ? "edit" : comment.content}
                className={taskCommentEditor({ editing })}
                content={editing ? draft : comment.content}
                controls={false}
                disabled={!editing}
                aria-label={t("details.commentPlaceholder")}
                mentionSource={mentionSource}
                onChangeContent={editing
                  ? (next) => {
                      setDraft(next);
                      setError(undefined);
                    }
                  : undefined}
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
                {t("details.cancel")}
              </Button>
              <Button
                size={1}
                variant="solid"
                color="neutral"
                highContrast
                loading={isSaving}
                onClick={() => {
                  void saveComment();
                }}
              >
                {t("details.save")}
              </Button>
            </div>
          )}
        </div>
      </div>
    </Card>
  );
}

export function TaskDetailsCommentForm({ taskId }: { taskId: string }) {
  const { t } = useTranslation("tasks");
  const [createComment, isCreating] = useMutation<TaskDetailsEngagementCreateCommentMutation>(
    createCommentMutation,
    {
      successMessage: t("details.commentCreated"),
      errorToast: t("details.commentCreateError"),
    },
  );
  const mentionSource = useMentionMembers();
  const [content, setContent] = useState("");
  const [editorKey, setEditorKey] = useState(0);
  const [error, setError] = useState<string>();
  const { root, actions, editor } = taskCommentForm();

  function handleSubmit() {
    if (isRichEditorContentEmpty(content)) {
      setError(t("details.commentRequired"));
      return;
    }
    if (richEditorContentTextLength(content) > taskCommentMaxLength) {
      setError(t("details.commentTooLong", { max: taskCommentMaxLength }));
      return;
    }

    void createComment({
      variables: {
        input: { taskId, content },
        connections: [taskActivitiesConnectionId(taskId)],
      },
    }).then(
      () => {
        setContent("");
        setEditorKey(key => key + 1);
        setError(undefined);
      },
      () => {
        // The mutation already shows an error toast.
      },
    );
  }

  return (
    <Card variant="soft" size={2}>
      <Form className={root()} onFormSubmit={handleSubmit}>
        <Field error={error}>
          <RichEditor
            key={editorKey}
            controls={false}
            className={editor()}
            content={content}
            disabled={isCreating}
            placeholder={t("details.commentPlaceholder")}
            aria-label={t("details.commentPlaceholder")}
            mentionSource={mentionSource}
            onChangeContent={(next) => {
              setContent(next);
              setError(undefined);
            }}
          />
        </Field>
        <div className={actions()}>
          <Button
            type="submit"
            size={1}
            variant="solid"
            color="neutral"
            highContrast
            loading={isCreating}
          >
            {t("details.send")}
          </Button>
        </div>
      </Form>
    </Card>
  );
}

function DeleteCommentDialog({
  taskId,
  commentId,
  open,
  onOpenChange,
}: {
  taskId: string;
  commentId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const { t } = useTranslation("tasks");
  const [deleteComment, isDeleting] = useMutation<TaskDetailsEngagementDeleteCommentMutation>(
    deleteCommentMutation,
    {
      successMessage: t("details.commentDeleted"),
      errorToast: t("details.commentDeleteError"),
    },
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("details.deleteCommentTitle")}</DialogTitle>
          <DialogDescription>{t("details.deleteCommentDescription")}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("delete.cancel")}
              </Button>
            )}
          />
          <Button
            type="button"
            variant="solid"
            color="red"
            iconStart={<TrashIcon />}
            loading={isDeleting}
            onClick={() => {
              void deleteComment({
                variables: {
                  input: { taskCommentId: commentId },
                  connections: [taskActivitiesConnectionId(taskId)],
                },
              }).then(
                () => {
                  onOpenChange(false);
                },
                () => {
                  // The mutation already shows an error toast.
                },
              );
            }}
          >
            {t("details.deleteComment")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
