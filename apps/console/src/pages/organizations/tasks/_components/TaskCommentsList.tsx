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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { List } from "@probo/ui/src/v2/List/List";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";
import { useParams } from "react-router";

import type { TaskCommentsList_comments$key } from "#/__generated__/core/TaskCommentsList_comments.graphql";
import type { TaskCommentsListQuery } from "#/__generated__/core/TaskCommentsListQuery.graphql";
import type { TaskCommentsListRefetchQuery } from "#/__generated__/core/TaskCommentsListRefetchQuery.graphql";

import { taskCommentsSection } from "../variants";

import { TaskCommentForm } from "./TaskCommentForm";
import { TaskCommentListItem } from "./TaskCommentListItem";
import { TaskCommentsEmpty } from "./TaskCommentsEmpty";

const taskCommentsListQuery = graphql`
  query TaskCommentsListQuery($taskId: ID!) {
    node(id: $taskId) {
      __typename
      ... on Task {
        ...TaskCommentsList_comments
      }
    }
  }
`;

const taskCommentsListFragment = graphql`
  fragment TaskCommentsList_comments on Task
  @refetchable(queryName: "TaskCommentsListRefetchQuery")
  @argumentDefinitions(
    last: { type: "Int", defaultValue: 20 }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    canCreateComment: permission(action: "core:task-comment:create")
    comments(
      last: $last
      before: $before
      orderBy: { field: CREATED_AT, direction: ASC }
    ) @connection(key: "TaskCommentsSection_comments") {
      edges {
        node {
          id
          ...TaskCommentListItem_comment
        }
      }
    }
  }
`;

interface TaskCommentsListContentProps {
  taskKey: TaskCommentsList_comments$key;
}

function TaskCommentsListContent({ taskKey }: TaskCommentsListContentProps) {
  const { t } = useTranslation("organizations/tasks");
  const {
    data: task,
    hasPrevious,
    loadPrevious,
    isLoadingPrevious,
  } = usePaginationFragment<
    TaskCommentsListRefetchQuery,
    TaskCommentsList_comments$key
  >(taskCommentsListFragment, taskKey);
  const { root, header, actions } = taskCommentsSection();
  const comments = task.comments.edges.map(edge => edge.node);

  return (
    <div className={root()}>
      <div className={header()}>
        <Heading level={2} size={4}>
          {t("detailsPage.comments.title")}
        </Heading>
      </div>
      {hasPrevious && (
        <div className={actions()}>
          <Button
            variant="ghost"
            color="neutral"
            disabled={isLoadingPrevious}
            onClick={() => {
              loadPrevious(20);
            }}
          >
            {t("detailsPage.comments.actions.showMore")}
          </Button>
        </div>
      )}
      {comments.length === 0
        ? <TaskCommentsEmpty />
        : (
            <List>
              {comments.map(comment => (
                <TaskCommentListItem
                  key={comment.id}
                  taskCommentKey={comment}
                />
              ))}
            </List>
          )}
      {task.canCreateComment && <TaskCommentForm />}
    </div>
  );
}

function TaskCommentsListLoaded({ taskId }: { taskId: string }) {
  const data = useLazyLoadQuery<TaskCommentsListQuery>(
    taskCommentsListQuery,
    { taskId },
  );

  if (data.node?.__typename !== "Task") {
    return null;
  }

  return <TaskCommentsListContent taskKey={data.node} />;
}

export function TaskCommentsList() {
  const { taskId } = useParams<{ taskId: string }>();
  if (taskId == null) {
    throw new Error(":taskId missing in route params");
  }

  return <TaskCommentsListLoaded taskId={taskId} />;
}
