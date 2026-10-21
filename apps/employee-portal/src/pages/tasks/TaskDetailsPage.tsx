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

import { CaretLeftIcon } from "@phosphor-icons/react";
import {
  TaskDetails,
  TaskEngagement,
  taskListItem,
  TaskListItemAssignee,
  TaskNameField,
  TaskProperties,
  TASKS_NAMESPACE,
} from "@probo/tasks";
import { RichEditor } from "@probo/ui/src/RichEditor/RichEditor";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  useFragment,
  usePreloadedQuery,
} from "react-relay";
import { useLocation } from "react-router";

import type { TaskDetailsPage_task$data, TaskDetailsPage_task$key } from "#/__generated__/core/TaskDetailsPage_task.graphql";
import type { TaskDetailsPageQuery } from "#/__generated__/core/TaskDetailsPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

import { TaskDetailsActivity, TaskDetailsCommentForm } from "./_components/TaskDetailsEngagement";
import { isRichTextEmpty } from "./_lib/richText";
import { useUpdateTask } from "./_lib/useUpdateTask";

export const taskDetailsPageFragment = graphql`
  fragment TaskDetailsPage_task on Task {
    id
    name
    content
    state
    priority
    timeEstimate
    deadline
    recurrenceInterval
    canUpdate: permission(action: "core:task:update")
    canAssign: permission(action: "core:task:assign")
    canListActivities: permission(action: "core:task-activity:list")
    canCreateComment: permission(action: "core:task-comment:create")
    organization {
      id
    }
    assignedTo {
      id
      fullName
      emailAddress
      avatar {
        downloadUrl
      }
    }
  }
`;

export const taskDetailsPageQuery = graphql`
  query TaskDetailsPageQuery($taskId: ID!) {
    node(id: $taskId) {
      __typename
      ... on Task {
        ...TaskDetailsPage_task
      }
    }
  }
`;

interface TaskDetailsPageProps {
  queryRef: PreloadedQuery<TaskDetailsPageQuery>;
}

export function TaskDetailsPage({ queryRef }: TaskDetailsPageProps) {
  const data = usePreloadedQuery<TaskDetailsPageQuery>(taskDetailsPageQuery, queryRef);
  const { t } = useTranslation("tasks");
  const { organizationId } = useEmployeePortalRoute();
  const taskKey = data.node?.__typename === "Task" ? data.node : null;
  const task = useFragment<TaskDetailsPage_task$key>(
    taskDetailsPageFragment,
    taskKey,
  );

  if (task == null || task.organization.id !== organizationId) {
    throw new NotFoundError(t("details.notFound"));
  }

  return <TaskDetailsBody task={task} />;
}

function TaskDetailsBody({ task }: { task: TaskDetailsPage_task$data }) {
  const { t } = useTranslation("tasks");
  const { t: tTasks } = useTranslation(TASKS_NAMESPACE);
  const { employeePortalId } = useEmployeePortalRoute();
  const location = useLocation();
  const [updateTask, isUpdating] = useUpdateTask(task);
  const saveDescription = useCallback((content: string | null) => {
    return updateTask({ content });
  }, [updateTask]);
  const { stateOption } = taskListItem({ field: true });
  const assignedTo = task.assignedTo;
  const showEngagement = task.canListActivities || task.canCreateComment;

  return (
    <main className="mx-auto flex w-full max-w-5xl flex-col px-8 pt-8 pb-32">
      <TaskDetails
        back={(
          <Link
            to={{
              pathname: `/${employeePortalId}/tasks`,
              search: location.search,
            }}
            size={2}
            color="neutral"
            underline={false}
            iconStart={<CaretLeftIcon />}
            className="self-start"
          >
            {t("breadcrumb")}
          </Link>
        )}
        name={(
          <TaskNameField
            name={task.name}
            canUpdate={task.canAssign}
            onSave={async (name) => {
              await updateTask({ name });
            }}
          />
        )}
        description={(
          <TaskDetailsDescription
            content={task.content}
            canAssign={task.canAssign}
            onSave={saveDescription}
          />
        )}
        engagement={showEngagement
          ? (
              <TaskEngagement
                canCreateComment={task.canCreateComment && !task.canListActivities}
                canListActivities={task.canListActivities}
                activity={task.canListActivities
                  ? <TaskDetailsActivity taskId={task.id} />
                  : null}
                composer={task.canCreateComment && !task.canListActivities
                  ? <TaskDetailsCommentForm taskId={task.id} />
                  : null}
              />
            )
          : null}
        properties={(
          <TaskProperties
            task={{
              state: task.state,
              priority: task.priority,
              timeEstimate: task.timeEstimate,
              deadline: task.deadline,
              recurrenceInterval: task.recurrenceInterval,
            }}
            canUpdate={task.canUpdate}
            isUpdating={isUpdating}
            assignee={(
              <TaskListItemAssignee
                assignee={assignedTo == null
                  ? null
                  : {
                      id: assignedTo.id,
                      fullName: assignedTo.fullName,
                      emailAddress: assignedTo.emailAddress,
                      avatarUrl: assignedTo.avatar?.downloadUrl,
                    }}
                canUpdate={false}
                showLabel
                placeholder={tTasks("placeholders.assignee")}
                className={stateOption()}
              />
            )}
            onChange={change => updateTask(change)}
          />
        )}
      />
    </main>
  );
}

function TaskDetailsDescription({
  content,
  canAssign,
  onSave,
}: {
  content: string;
  canAssign: boolean;
  onSave: (content: string | null) => Promise<unknown>;
}) {
  const { t } = useTranslation("tasks");
  const [draft, setDraft] = useState(content);
  const [seen, setSeen] = useState(content);
  if (content !== seen) {
    setSeen(content);
    setDraft(content);
  }

  useEffect(() => {
    if (!canAssign || draft === content) {
      return;
    }

    const handle = setTimeout(() => {
      const next = isRichTextEmpty(draft) ? null : draft;
      void onSave(next).catch(() => {
        setDraft(content);
      });
    }, 1000);

    return () => {
      clearTimeout(handle);
    };
  }, [canAssign, content, draft, onSave]);

  if (!canAssign && isRichTextEmpty(content)) {
    return <Text size={2} color="faint">{t("details.noDescription")}</Text>;
  }

  return (
    <RichEditor
      className="min-h-40 bg-transparent! py-0! pr-0! shadow-none!"
      content={canAssign ? draft : content}
      controls={false}
      disabled={!canAssign}
      placeholder={t("details.descriptionPlaceholder")}
      aria-label={t("details.description")}
      onChangeContent={canAssign ? setDraft : undefined}
    />
  );
}
