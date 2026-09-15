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
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";
import { useParams } from "react-router";

import type { TaskActivityList_activities$key } from "#/__generated__/core/TaskActivityList_activities.graphql";
import type { TaskActivityListQuery } from "#/__generated__/core/TaskActivityListQuery.graphql";
import type { TaskActivityListRefetchQuery } from "#/__generated__/core/TaskActivityListRefetchQuery.graphql";

import { taskCommentsSection } from "../variants";

import { TaskActivityEmpty } from "./TaskActivityEmpty";
import { TaskActivityListItem } from "./TaskActivityListItem";

const taskActivityListQuery = graphql`
  query TaskActivityListQuery($taskId: ID!) {
    node(id: $taskId) {
      __typename
      ... on Task {
        ...TaskActivityList_activities
      }
    }
  }
`;

const taskActivityListFragment = graphql`
  fragment TaskActivityList_activities on Task
  @refetchable(queryName: "TaskActivityListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    activities(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: DESC }
    ) @connection(key: "TaskActivitySection_activities") {
      edges {
        node {
          id
          ...TaskActivityListItem_activity
        }
      }
    }
  }
`;

interface TaskActivityListContentProps {
  taskKey: TaskActivityList_activities$key;
}

function TaskActivityListContent({ taskKey }: TaskActivityListContentProps) {
  const { t } = useTranslation("organizations/tasks");
  const {
    data: task,
    hasNext,
    loadNext,
    isLoadingNext,
  } = usePaginationFragment<
    TaskActivityListRefetchQuery,
    TaskActivityList_activities$key
  >(taskActivityListFragment, taskKey);
  const { root, actions } = taskCommentsSection();
  const activities = task.activities.edges.map(edge => edge.node);

  return (
    <div className={root()}>
      {activities.length === 0
        ? <TaskActivityEmpty />
        : (
            <List>
              {activities.map(activity => (
                <TaskActivityListItem
                  key={activity.id}
                  taskActivityKey={activity}
                />
              ))}
            </List>
          )}
      {hasNext && (
        <div className={actions()}>
          <Button
            variant="ghost"
            color="neutral"
            disabled={isLoadingNext}
            onClick={() => {
              loadNext(20);
            }}
          >
            {t("detailsPage.activity.actions.showMore")}
          </Button>
        </div>
      )}
    </div>
  );
}

function TaskActivityListLoaded({
  taskId,
  fetchKey,
}: {
  taskId: string;
  fetchKey: string;
}) {
  const data = useLazyLoadQuery<TaskActivityListQuery>(
    taskActivityListQuery,
    { taskId },
    { fetchKey, fetchPolicy: "store-and-network" },
  );

  if (data.node?.__typename !== "Task") {
    return null;
  }

  return <TaskActivityListContent taskKey={data.node} />;
}

interface TaskActivityListProps {
  fetchKey: string;
}

export function TaskActivityList({ fetchKey }: TaskActivityListProps) {
  const { taskId } = useParams<{ taskId: string }>();
  if (taskId == null) {
    throw new Error(":taskId missing in route params");
  }

  return <TaskActivityListLoaded taskId={taskId} fetchKey={fetchKey} />;
}
