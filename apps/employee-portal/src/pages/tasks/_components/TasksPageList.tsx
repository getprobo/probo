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

import { TaskListView, type TaskState } from "@probo/tasks";
import { useEffect, useRef, useState, useTransition } from "react";
import { graphql, readInlineData, useRefetchableFragment } from "react-relay";

import type { TasksPageList_task$key } from "#/__generated__/core/TasksPageList_task.graphql";
import type { TasksPageList_viewer$key } from "#/__generated__/core/TasksPageList_viewer.graphql";
import type { TasksPageListRefetchQuery } from "#/__generated__/core/TasksPageListRefetchQuery.graphql";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

import {
  mergeTaskListIncludeFlags,
  taskListNeedsFetch,
  taskListReadableIncludeFlags,
} from "../_lib/taskListIncludes";
import { rememberTaskListDeadline } from "../_lib/taskPath";
import { useTasksFilters } from "../_lib/useTasksFilters";
import {
  type TasksPageListGroupNode,
  useTasksPageListGroups,
} from "../_lib/useTasksPageListGroups";

import { TasksPageTaskListItem } from "./TasksPageTaskListItem";

const taskInlineFragment = graphql`
  fragment TasksPageList_task on Task @inline {
    id
    state
  }
`;

const tasksPageListFragment = graphql`
  fragment TasksPageList_viewer on Viewer
  @refetchable(queryName: "TasksPageListRefetchQuery")
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    query: { type: "String" }
    deadlineOnOrBefore: { type: "Datetime" }
    deadlineOnOrAfter: { type: "Datetime" }
    hasDeadline: { type: "Boolean" }
    includeInProgress: { type: "Boolean!" }
    includeTodo: { type: "Boolean!" }
    includeBacklog: { type: "Boolean!" }
    includeDone: { type: "Boolean!" }
    includeCanceled: { type: "Boolean!" }
    includeDuplicate: { type: "Boolean!" }
  )
  @throwOnFieldError {
    ...useTasksPageListGroups_inProgress
      @arguments(
        organizationId: $organizationId
        query: $query
        deadlineOnOrBefore: $deadlineOnOrBefore
        deadlineOnOrAfter: $deadlineOnOrAfter
        hasDeadline: $hasDeadline
      )
      @include(if: $includeInProgress)
    ...useTasksPageListGroups_todo
      @arguments(
        organizationId: $organizationId
        query: $query
        deadlineOnOrBefore: $deadlineOnOrBefore
        deadlineOnOrAfter: $deadlineOnOrAfter
        hasDeadline: $hasDeadline
      )
      @include(if: $includeTodo)
    ...useTasksPageListGroups_backlog
      @arguments(
        organizationId: $organizationId
        query: $query
        deadlineOnOrBefore: $deadlineOnOrBefore
        deadlineOnOrAfter: $deadlineOnOrAfter
        hasDeadline: $hasDeadline
      )
      @include(if: $includeBacklog)
    ...useTasksPageListGroups_done
      @arguments(
        organizationId: $organizationId
        query: $query
        deadlineOnOrBefore: $deadlineOnOrBefore
        deadlineOnOrAfter: $deadlineOnOrAfter
        hasDeadline: $hasDeadline
      )
      @include(if: $includeDone)
    ...useTasksPageListGroups_canceled
      @arguments(
        organizationId: $organizationId
        query: $query
        deadlineOnOrBefore: $deadlineOnOrBefore
        deadlineOnOrAfter: $deadlineOnOrAfter
        hasDeadline: $hasDeadline
      )
      @include(if: $includeCanceled)
    ...useTasksPageListGroups_duplicate
      @arguments(
        organizationId: $organizationId
        query: $query
        deadlineOnOrBefore: $deadlineOnOrBefore
        deadlineOnOrAfter: $deadlineOnOrAfter
        hasDeadline: $hasDeadline
      )
      @include(if: $includeDuplicate)
  }
`;

type TasksPageListRowItem = {
  id: string;
  state: TaskState;
  node: TasksPageListGroupNode;
};

function readTask(key: TasksPageList_task$key) {
  return readInlineData(taskInlineFragment, key);
}

export interface TasksPageListProps {
  viewerKey: TasksPageList_viewer$key;
}

export function TasksPageList({ viewerKey }: TasksPageListProps) {
  const { organizationId } = useEmployeePortalRoute();
  const { query, graphqlQuery, deadlineBounds, due, state: selectedState, includeFlags } = useTasksFilters();
  rememberTaskListDeadline(deadlineBounds);
  const [isPending, startTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [fetched, setFetched] = useState(() => ({
    query: graphqlQuery,
    deadlineOnOrBefore: deadlineBounds.deadlineOnOrBefore,
    deadlineOnOrAfter: deadlineBounds.deadlineOnOrAfter,
    hasDeadline: deadlineBounds.hasDeadline,
    includes: includeFlags,
  }));
  const [viewer, refetch] = useRefetchableFragment<
    TasksPageListRefetchQuery,
    TasksPageList_viewer$key
  >(tasksPageListFragment, viewerKey);
  const groups = useTasksPageListGroups(
    viewer,
    taskListReadableIncludeFlags(includeFlags, fetched.includes),
  );

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    const deadlineChanged = deadlineBounds.deadlineOnOrBefore !== fetched.deadlineOnOrBefore
      || deadlineBounds.deadlineOnOrAfter !== fetched.deadlineOnOrAfter
      || deadlineBounds.hasDeadline !== fetched.hasDeadline;
    if (!deadlineChanged && !taskListNeedsFetch(graphqlQuery, includeFlags, fetched)) {
      return;
    }

    const requestedQuery = graphqlQuery;
    const requestedDeadline = deadlineBounds;
    const requestedIncludes = includeFlags;
    startTransition(() => {
      refetch(
        {
          organizationId,
          query: requestedQuery,
          deadlineOnOrBefore: requestedDeadline.deadlineOnOrBefore,
          deadlineOnOrAfter: requestedDeadline.deadlineOnOrAfter,
          hasDeadline: requestedDeadline.hasDeadline,
          ...requestedIncludes,
        },
        {
          fetchPolicy: "network-only",
          onComplete(error) {
            if (error) {
              return;
            }

            setFetched(current => ({
              query: requestedQuery,
              deadlineOnOrBefore: requestedDeadline.deadlineOnOrBefore,
              deadlineOnOrAfter: requestedDeadline.deadlineOnOrAfter,
              hasDeadline: requestedDeadline.hasDeadline,
              includes: current.query === requestedQuery
                && current.deadlineOnOrBefore === requestedDeadline.deadlineOnOrBefore
                && current.deadlineOnOrAfter === requestedDeadline.deadlineOnOrAfter
                && current.hasDeadline === requestedDeadline.hasDeadline
                ? mergeTaskListIncludeFlags(current.includes, requestedIncludes)
                : requestedIncludes,
            }));
          },
        },
      );
    });
  }, [deadlineBounds, fetched, graphqlQuery, includeFlags, organizationId, refetch]);

  return (
    <TaskListView<TasksPageListRowItem>
      groups={(due === "overdue"
        ? groups.filter(group => group.state !== "DONE" && group.state !== "CANCELED" && group.state !== "DUPLICATE")
        : groups).map(group => ({
        state: group.state,
        totalCount: group.totalCount,
        loadedItems: group.edges.map(({ node }) => ({
          ...readTask(node),
          node,
        })),
        hasNext: group.hasNext,
        isLoadingNext: group.isLoadingNext,
        loadNext: group.loadNext,
      }))}
      selectedState={selectedState}
      query={query}
      filtered={due != null}
      busy={isPending}
      renderRow={(item, sectionState) => (
        <TasksPageTaskListItem
          taskKey={item.node}
          sectionState={sectionState}
        />
      )}
    />
  );
}
