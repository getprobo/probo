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

import {
  flattenTaskListItems,
  taskList,
  TaskListGroupHeader,
  taskListItemSize,
  TaskListRowFrame,
  TASKS_NAMESPACE,
  type TaskState,
} from "@probo/tasks";
import { TooltipProvider } from "@probo/ui/src/v2/Tooltip/TooltipProvider";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useVirtualizer } from "@tanstack/react-virtual";
import { useEffect, useMemo, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
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
import {
  type TasksPageListGroupNode,
  useTasksPageListGroups,
} from "../_lib/useTasksPageListGroups";
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

type TasksPageListRowItem = {
  id: string;
  state: TaskState;
  node: TasksPageListGroupNode;
};

function readTask(key: TasksPageList_task$key) {
  return readInlineData(taskInlineFragment, key);
}

function nearestOverflowParent(element: HTMLElement | null): HTMLElement | null {
  let current = element?.parentElement ?? null;
  while (current != null) {
    const overflowY = getComputedStyle(current).overflowY;
    if (overflowY === "auto" || overflowY === "scroll") {
      return current;
    }
    current = current.parentElement;
  }
  return null;
}

function flatItemKey(
  item: ReturnType<typeof flattenTaskListItems<TasksPageListRowItem>>[number],
  index: number,
): string {
  if (item.type === "header") {
    return `header:${item.state}`;
  }
  if (item.type === "row") {
    return item.item.id;
  }
  return `placeholder:${item.state}:${index}`;
}

export interface TasksPageListProps {
  viewerKey: TasksPageList_viewer$key;
}

export function TasksPageList({ viewerKey }: TasksPageListProps) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const { organizationId } = useEmployeePortalRoute();
  const { query, graphqlQuery, state: selectedState, includeFlags } = useTasksFilters();
  const [isPending, startTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [fetched, setFetched] = useState(() => ({
    query: graphqlQuery,
    includes: includeFlags,
  }));
  const [closedStates, setClosedStates] = useState(() => new Set<TaskState>());
  const listRef = useRef<HTMLDivElement>(null);
  const [scrollElement, setScrollElement] = useState<HTMLElement | null>(null);
  const [viewer, refetch] = useRefetchableFragment<
    TasksPageListRefetchQuery,
    TasksPageList_viewer$key
  >(tasksPageListFragment, viewerKey);
  const groups = useTasksPageListGroups(
    viewer,
    taskListReadableIncludeFlags(includeFlags, fetched.includes),
  );
  const { list } = tasksPage({ busy: isPending });
  const slots = taskList();
  const visibleGroups = groups.filter(
    group => selectedState == null || group.state === selectedState,
  );
  const flatItems = useMemo(
    () => flattenTaskListItems(
      visibleGroups.map(group => ({
        state: group.state,
        totalCount: group.totalCount,
        loadedItems: group.edges.map(({ node }) => ({
          ...readTask(node),
          node,
        })),
        open: !closedStates.has(group.state),
      })),
    ),
    [closedStates, visibleGroups],
  );
  const virtualizer = useVirtualizer({
    count: flatItems.length,
    getScrollElement: () => scrollElement,
    estimateSize: index => taskListItemSize(flatItems[index]),
    overscan: 8,
    getItemKey: index => flatItemKey(flatItems[index], index),
  });

  useEffect(() => {
    setScrollElement(nearestOverflowParent(listRef.current));
  }, []);

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

  function onGroupOpenChange(state: TaskState, open: boolean) {
    setClosedStates((current) => {
      const next = new Set(current);
      if (open) {
        next.delete(state);
      } else {
        next.add(state);
      }
      return next;
    });
  }

  if (flatItems.length === 0) {
    return (
      <div className={list()} aria-busy={isPending || undefined}>
        <Text size={2} color="faint" align="center" className={slots.empty()}>
          {query !== "" || selectedState != null ? t("noResults") : t("empty")}
        </Text>
      </div>
    );
  }

  return (
    <div ref={listRef} className={list()} aria-busy={isPending || undefined}>
      <TooltipProvider>
        <div
          className="relative w-full"
          style={{ height: virtualizer.getTotalSize() }}
        >
          {virtualizer.getVirtualItems().map((virtualItem) => {
            const item = flatItems[virtualItem.index];
            return (
              <div
                key={virtualItem.key}
                data-index={virtualItem.index}
                ref={virtualizer.measureElement}
                className="absolute top-0 left-0 w-full"
                style={{ transform: `translateY(${virtualItem.start}px)` }}
              >
                {item.type === "header"
                  ? (
                      <TaskListGroupHeader
                        state={item.state}
                        count={item.count}
                        open={!closedStates.has(item.state)}
                        offset={item.offset}
                        onOpenChange={open => onGroupOpenChange(item.state, open)}
                      />
                    )
                  : (
                      <TaskListRowFrame
                        position={item.position}
                        placeholder={item.type === "placeholder"}
                      >
                        {item.type === "row" && (
                          <TasksPageTaskListItem
                            taskKey={item.item.node}
                            sectionState={item.state}
                          />
                        )}
                      </TaskListRowFrame>
                    )}
              </div>
            );
          })}
        </div>
      </TooltipProvider>
    </div>
  );
}
