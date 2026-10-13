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

import { TaskList } from "@probo/tasks";
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
import { useTasksFilters } from "../_lib/useTasksFilters";
import { useTasksPageListGroups } from "../_lib/useTasksPageListGroups";
import { tasksPage } from "../variants";

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
    includeInProgress: { type: "Boolean!" }
    includeTodo: { type: "Boolean!" }
    includeBacklog: { type: "Boolean!" }
    includeDone: { type: "Boolean!" }
    includeCanceled: { type: "Boolean!" }
    includeDuplicate: { type: "Boolean!" }
  )
  @throwOnFieldError {
    ...useTasksPageListGroups_inProgress
      @arguments(organizationId: $organizationId, query: $query)
      @include(if: $includeInProgress)
    ...useTasksPageListGroups_todo
      @arguments(organizationId: $organizationId, query: $query)
      @include(if: $includeTodo)
    ...useTasksPageListGroups_backlog
      @arguments(organizationId: $organizationId, query: $query)
      @include(if: $includeBacklog)
    ...useTasksPageListGroups_done
      @arguments(organizationId: $organizationId, query: $query)
      @include(if: $includeDone)
    ...useTasksPageListGroups_canceled
      @arguments(organizationId: $organizationId, query: $query)
      @include(if: $includeCanceled)
    ...useTasksPageListGroups_duplicate
      @arguments(organizationId: $organizationId, query: $query)
      @include(if: $includeDuplicate)
  }
`;

function readTask(key: TasksPageList_task$key) {
  return readInlineData(taskInlineFragment, key);
}

export interface TasksPageListProps {
  viewerKey: TasksPageList_viewer$key;
}

export function TasksPageList({ viewerKey }: TasksPageListProps) {
  const { organizationId } = useEmployeePortalRoute();
  const { query, graphqlQuery, state: selectedState, includeFlags } = useTasksFilters();
  const [isPending, startTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [fetched, setFetched] = useState(() => ({
    query: graphqlQuery,
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
  const { list } = tasksPage({ busy: isPending });
  const visibleGroups = groups.filter(
    group => selectedState == null || group.state === selectedState,
  );
  const items = visibleGroups.flatMap(group =>
    group.edges.map(({ node }) => ({
      ...readTask(node),
      node,
    })),
  );

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    if (!taskListNeedsFetch(graphqlQuery, includeFlags, fetched)) {
      return;
    }

    const requestedQuery = graphqlQuery;
    const requestedIncludes = includeFlags;
    startTransition(() => {
      refetch(
        {
          organizationId,
          query: requestedQuery,
          ...requestedIncludes,
        },
        {
          fetchPolicy: "store-or-network",
          onComplete(error) {
            if (error) {
              return;
            }

            setFetched(current => ({
              query: requestedQuery,
              includes: current.query === requestedQuery
                ? mergeTaskListIncludeFlags(current.includes, requestedIncludes)
                : requestedIncludes,
            }));
          },
        },
      );
    });
  }, [fetched, graphqlQuery, includeFlags, organizationId, refetch]);

  return (
    <div className={list()} aria-busy={isPending || undefined}>
      <TaskList
        items={items}
        filtered={query !== "" || selectedState != null}
        renderItem={(item, sectionState) => (
          <TasksPageTaskListItem
            key={item.id}
            taskKey={item.node}
            sectionState={sectionState}
          />
        )}
      />
    </div>
  );
}
