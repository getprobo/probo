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

import { formatError } from "@probo/helpers";
import {
  TaskListToolbar,
  TaskListView,
  taskPriorities,
  type TaskPriority,
  type TaskState,
} from "@probo/tasks";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { type DragEvent, useEffect, useMemo, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, readInlineData, useRefetchableFragment } from "react-relay";
import type { RecordSourceSelectorProxy } from "relay-runtime";

import type { TasksPageList_organization$key } from "#/__generated__/core/TasksPageList_organization.graphql";
import type { TasksPageList_task$key } from "#/__generated__/core/TasksPageList_task.graphql";
import type { TasksPageListRefetchQuery } from "#/__generated__/core/TasksPageListRefetchQuery.graphql";
import type { TasksPageListUpdateRankMutation } from "#/__generated__/core/TasksPageListUpdateRankMutation.graphql";
import { TaskInternalControlFilter } from "#/components/tasks/TaskInternalControlFilter";
import { TaskListItem } from "#/components/tasks/TaskListItem";
import { TaskUserFilter } from "#/components/tasks/TaskUserFilter";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { applyOrganizationTaskRankDrop } from "../_lib/moveOrganizationTaskListEdge";
import {
  mergeTaskListIncludeFlags,
  taskListIncludeFlags,
  taskListNeedsFetch,
  taskListReadableIncludeFlags,
} from "../_lib/taskListIncludes";
import { organizationTaskListFilterVariants } from "../_lib/taskPath";
import { useOrganizationTaskListGroups } from "../_lib/useOrganizationTaskListGroups";
import { useTasksCardFilters } from "../_lib/useTasksCardFilters";
import { useTasksCardSearch } from "../_lib/useTasksCardSearch";
import { tasksPage } from "../variants";

const taskInlineFragment = graphql`
  fragment TasksPageList_task on Task @inline {
    id
    state
    priority
    rank
  }
`;

const tasksPageListFragment = graphql`
  fragment TasksPageList_organization on Organization
  @refetchable(queryName: "TasksPageListRefetchQuery")
  @argumentDefinitions(
    query: { type: "String" }
    assignedToId: { type: "ID" }
    internalControlId: { type: "ID" }
    includeInProgress: { type: "Boolean!" }
    includeTodo: { type: "Boolean!" }
    includeBacklog: { type: "Boolean!" }
    includeDone: { type: "Boolean!" }
    includeCanceled: { type: "Boolean!" }
    includeDuplicate: { type: "Boolean!" }
  ) {
    canUpdateTask: permission(action: "core:task:update")
    ...useOrganizationTaskListGroups_inProgress
      @arguments(
        query: $query
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      )
      @include(if: $includeInProgress)
    ...useOrganizationTaskListGroups_todo
      @arguments(
        query: $query
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      )
      @include(if: $includeTodo)
    ...useOrganizationTaskListGroups_backlog
      @arguments(
        query: $query
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      )
      @include(if: $includeBacklog)
    ...useOrganizationTaskListGroups_done
      @arguments(
        query: $query
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      )
      @include(if: $includeDone)
    ...useOrganizationTaskListGroups_canceled
      @arguments(
        query: $query
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      )
      @include(if: $includeCanceled)
    ...useOrganizationTaskListGroups_duplicate
      @arguments(
        query: $query
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      )
      @include(if: $includeDuplicate)
  }
`;

const updateRankMutation = graphql`
  mutation TasksPageListUpdateRankMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        id
        priority
        rank
        state
        ...TasksPageList_task
        ...TaskListItem_task
      }
    }
  }
`;

function readTask(key: TasksPageList_task$key) {
  return readInlineData(taskInlineFragment, key);
}

function resolveDropPriority(
  dragged: TaskPriority,
  above?: TaskPriority,
  below?: TaskPriority,
): TaskPriority | undefined {
  if (above === dragged || below === dragged) {
    return undefined;
  }

  const draggedIndex = taskPriorities.indexOf(dragged);
  if (above == null && below != null) {
    return taskPriorities.indexOf(below) <= draggedIndex ? below : undefined;
  }
  if (below == null && above != null) {
    return taskPriorities.indexOf(above) >= draggedIndex ? above : undefined;
  }
  if (above != null && below != null) {
    const distanceAbove = Math.abs(taskPriorities.indexOf(above) - draggedIndex);
    const distanceBelow = Math.abs(taskPriorities.indexOf(below) - draggedIndex);
    return distanceAbove <= distanceBelow ? above : below;
  }

  return undefined;
}

export interface TasksPageListProps {
  organizationKey: TasksPageList_organization$key;
}

export function TasksPageList({ organizationKey }: TasksPageListProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const {
    query,
    state: selectedState,
    assignedToId,
    internalControlId,
    graphqlFilter,
    setState,
    setAssignedToId,
    setInternalControlId,
  } = useTasksCardFilters();
  const [queryInput, setQueryInput] = useTasksCardSearch();
  const includeFlags = useMemo(
    () => taskListIncludeFlags(selectedState),
    [selectedState],
  );
  const [isPending, startTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [fetched, setFetched] = useState(() => ({
    query: graphqlFilter.query,
    assignedToId: graphqlFilter.assignedToId,
    internalControlId: graphqlFilter.internalControlId,
    includes: includeFlags,
  }));
  const [organization, refetch] = useRefetchableFragment<
    TasksPageListRefetchQuery,
    TasksPageList_organization$key
  >(tasksPageListFragment, organizationKey);
  const groups = useOrganizationTaskListGroups(
    organization,
    taskListReadableIncludeFlags(includeFlags, fetched.includes),
  );
  const toast = useToast();
  const [draggedId, setDraggedId] = useState<string | null>(null);
  const [draggedState, setDraggedState] = useState<TaskState | null>(null);
  const [preview, setPreview] = useState<{ state: TaskState; ids: string[] } | null>(null);
  const [ignoreClick, setIgnoreClick] = useState(false);
  const droppedRef = useRef(false);
  const [updateRank] = useMutation<TasksPageListUpdateRankMutation>(
    updateRankMutation,
    { errorToast: false },
  );
  const { list, userFilter, internalControlFilter } = tasksPage();
  const listFilter = {
    query: graphqlFilter.query,
    assignedToId: graphqlFilter.assignedToId,
    internalControlId: graphqlFilter.internalControlId,
  };
  const canDrag = organization.canUpdateTask
    && query === ""
    && assignedToId == null
    && internalControlId == null;

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    const filtersChanged = graphqlFilter.query !== fetched.query
      || graphqlFilter.assignedToId !== fetched.assignedToId
      || graphqlFilter.internalControlId !== fetched.internalControlId;
    const includesStale = taskListNeedsFetch(graphqlFilter.query, includeFlags, {
      query: fetched.query,
      includes: fetched.includes,
    });
    if (!filtersChanged && !includesStale) {
      return;
    }

    const requested = {
      query: graphqlFilter.query,
      assignedToId: graphqlFilter.assignedToId,
      internalControlId: graphqlFilter.internalControlId,
      includes: includeFlags,
    };
    startTransition(() => {
      refetch(
        {
          query: requested.query,
          assignedToId: requested.assignedToId,
          internalControlId: requested.internalControlId,
          ...requested.includes,
        },
        {
          fetchPolicy: "store-or-network",
          onComplete(error) {
            if (error) {
              return;
            }

            setFetched(current => ({
              query: requested.query,
              assignedToId: requested.assignedToId,
              internalControlId: requested.internalControlId,
              includes: current.query === requested.query
                && current.assignedToId === requested.assignedToId
                && current.internalControlId === requested.internalControlId
                ? mergeTaskListIncludeFlags(current.includes, requested.includes)
                : requested.includes,
            }));
          },
        },
      );
    });
  }, [fetched, graphqlFilter, includeFlags, refetch]);

  function resetDragState() {
    setDraggedId(null);
    setDraggedState(null);
    setPreview(null);
  }

  function handleDragOver(
    event: DragEvent,
    hoveredId: string,
    hoveredState: TaskState,
  ) {
    event.preventDefault();
    if (
      draggedId == null
      || draggedState == null
      || hoveredState !== draggedState
      || hoveredId === draggedId
    ) {
      return;
    }

    const group = groups.find(candidate => candidate.state === hoveredState);
    const ids = (group?.edges ?? []).map(({ node }) => readTask(node).id);
    const fromIndex = ids.indexOf(draggedId);
    const hoverIndex = ids.indexOf(hoveredId);
    if (fromIndex === -1 || hoverIndex === -1) {
      return;
    }

    const rect = event.currentTarget.getBoundingClientRect();
    const insertBefore = event.clientY < rect.top + rect.height / 2;
    let targetIndex = insertBefore ? hoverIndex : hoverIndex + 1;
    if (targetIndex > fromIndex) {
      targetIndex -= 1;
    }
    if (targetIndex === fromIndex) {
      setPreview(null);
      return;
    }

    const reordered = [...ids];
    reordered.splice(fromIndex, 1);
    reordered.splice(targetIndex, 0, draggedId);
    setPreview({ state: hoveredState, ids: reordered });
  }

  function handleDrop() {
    if (draggedId == null || draggedState == null || preview == null || preview.state !== draggedState) {
      resetDragState();
      return;
    }

    const group = groups.find(candidate => candidate.state === draggedState);
    const items = (group?.edges ?? []).map(({ node }) => readTask(node));
    const byId = new Map(items.map(item => [item.id, item]));
    const order = preview.ids;
    const newIndex = order.indexOf(draggedId);
    const dragged = byId.get(draggedId);
    if (newIndex === -1 || dragged == null) {
      resetDragState();
      return;
    }

    const targetRank = items[newIndex]?.rank ?? dragged.rank;
    const above = newIndex > 0 ? byId.get(order[newIndex - 1]) : undefined;
    const below = newIndex < order.length - 1 ? byId.get(order[newIndex + 1]) : undefined;
    const targetPriority = resolveDropPriority(
      dragged.priority,
      above?.priority,
      below?.priority,
    );
    const taskId = draggedId;
    const state = draggedState;
    const previousRank = dragged.rank;
    const filters = organizationTaskListFilterVariants(listFilter);
    droppedRef.current = true;

    function applyRankDrop(store: RecordSourceSelectorProxy) {
      const node = store.get(taskId);
      if (node == null) {
        return;
      }
      applyOrganizationTaskRankDrop(
        store,
        node,
        organizationId,
        state,
        filters,
        {
          previousRank,
          nextRank: targetRank,
          nextPriority: targetPriority,
        },
      );
    }

    void updateRank({
      variables: {
        input: {
          taskId,
          rank: targetRank,
          ...(targetPriority != null ? { priority: targetPriority } : {}),
        },
      },
      optimisticUpdater: applyRankDrop,
      updater: applyRankDrop,
      onCompleted(_, errors) {
        if (errors?.length) {
          toast.add({
            title: t("tasksCard.error.title"),
            description: formatError(t("tasksCard.error.reorder"), errors),
            type: "error",
          });
          startTransition(() => {
            refetch(
              {
                query: listFilter.query,
                assignedToId: listFilter.assignedToId,
                internalControlId: listFilter.internalControlId,
                ...includeFlags,
              },
              { fetchPolicy: "network-only" },
            );
          });
        }
        droppedRef.current = false;
        resetDragState();
      },
      onError() {
        droppedRef.current = false;
        resetDragState();
        toast.add({
          title: t("tasksCard.error.title"),
          description: t("tasksCard.error.reorder"),
          type: "error",
        });
      },
    }).catch(() => {
      // Error feedback is handled by the callbacks above.
    });
  }

  return (
    <div className={list()}>
      <TaskListToolbar
        query={queryInput}
        onQueryChange={setQueryInput}
        state={selectedState}
        onStateChange={setState}
      >
        <div className={userFilter()}>
          <TaskUserFilter value={assignedToId} onValueChange={setAssignedToId} />
        </div>
        <div className={internalControlFilter()}>
          <TaskInternalControlFilter
            value={internalControlId}
            onValueChange={setInternalControlId}
          />
        </div>
      </TaskListToolbar>
      <TaskListView
        groups={groups.map(group => ({
          state: group.state,
          totalCount: group.totalCount,
          loadedItems: (() => {
            const items = group.edges.map(({ node }) => ({
              ...readTask(node),
              node,
            }));
            if (preview == null || preview.state !== group.state) {
              return items;
            }
            const byId = new Map(items.map(item => [item.id, item]));
            return preview.ids.flatMap((id) => {
              const item = byId.get(id);
              return item == null ? [] : [item];
            });
          })(),
          hasNext: group.hasNext,
          isLoadingNext: group.isLoadingNext,
          loadNext: (count, options) => {
            group.loadNext(count, options);
          },
        }))}
        selectedState={selectedState}
        query={query}
        filtered={assignedToId != null || internalControlId != null}
        busy={isPending}
        renderRow={(item, sectionState) => (
          <TaskListItem
            taskKey={item.node}
            sectionState={sectionState}
            canDrag={canDrag}
            isDragging={draggedId === item.id}
            isGhost={preview != null && draggedId === item.id}
            ignoreClick={ignoreClick}
            onDragStart={() => {
              setIgnoreClick(true);
              setDraggedId(item.id);
              setDraggedState(sectionState);
            }}
            onDragOver={event => handleDragOver(event, item.id, sectionState)}
            onDrop={handleDrop}
            onDragEnd={() => {
              if (!droppedRef.current) {
                resetDragState();
              }
              window.setTimeout(() => {
                setIgnoreClick(false);
              }, 0);
            }}
          />
        )}
      />
    </div>
  );
}
