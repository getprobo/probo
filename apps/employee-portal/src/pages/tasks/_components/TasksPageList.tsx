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
import { useEffect, useRef, useTransition } from "react";
import { graphql, readInlineData, useRefetchableFragment } from "react-relay";

import type { TasksPageList_task$key } from "#/__generated__/core/TasksPageList_task.graphql";
import type { TasksPageList_viewer$key } from "#/__generated__/core/TasksPageList_viewer.graphql";
import type { TasksPageListRefetchQuery } from "#/__generated__/core/TasksPageListRefetchQuery.graphql";

import { useTasksFilters } from "../_lib/useTasksFilters";
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
    filter: { type: "TaskFilter" }
  )
  @throwOnFieldError {
    assignedTasks(
      organizationId: $organizationId
      first: 500
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: $filter
    ) {
      edges {
        node {
          ...TasksPageList_task
          ...TasksPageTaskListItem_task
        }
      }
    }
  }
`;

function readTask(key: TasksPageList_task$key) {
  return readInlineData(taskInlineFragment, key);
}

export interface TasksPageListProps {
  viewerKey: TasksPageList_viewer$key;
}

export function TasksPageList({ viewerKey }: TasksPageListProps) {
  const { query, state: selectedState, graphqlFilter } = useTasksFilters();
  const [isPending, startTransition] = useTransition();
  const skipFirstFilterRefetch = useRef(true);
  const [viewer, refetch] = useRefetchableFragment<
    TasksPageListRefetchQuery,
    TasksPageList_viewer$key
  >(tasksPageListFragment, viewerKey);
  const { list } = tasksPage({ busy: isPending });

  useEffect(() => {
    if (skipFirstFilterRefetch.current) {
      skipFirstFilterRefetch.current = false;
      return;
    }

    startTransition(() => {
      refetch({ filter: graphqlFilter }, { fetchPolicy: "store-or-network" });
    });
  }, [graphqlFilter, refetch]);

  const items = viewer.assignedTasks.edges.map(({ node }) => ({
    ...readTask(node),
    node,
  }));

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
