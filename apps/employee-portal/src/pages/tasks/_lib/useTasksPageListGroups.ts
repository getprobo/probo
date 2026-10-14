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

import { taskListGroupStates, taskListPageSize, type TaskState } from "@probo/tasks";
import { graphql, usePaginationFragment } from "react-relay";
import type { FragmentRefs } from "relay-runtime";

import type { TasksPageList_task$key } from "#/__generated__/core/TasksPageList_task.graphql";
import type { TasksPageTaskListItem_task$key } from "#/__generated__/core/TasksPageTaskListItem_task.graphql";
import type { useTasksPageListGroups_backlog$key } from "#/__generated__/core/useTasksPageListGroups_backlog.graphql";
import type { useTasksPageListGroups_canceled$key } from "#/__generated__/core/useTasksPageListGroups_canceled.graphql";
import type { useTasksPageListGroups_done$key } from "#/__generated__/core/useTasksPageListGroups_done.graphql";
import type { useTasksPageListGroups_duplicate$key } from "#/__generated__/core/useTasksPageListGroups_duplicate.graphql";
import type { useTasksPageListGroups_inProgress$key } from "#/__generated__/core/useTasksPageListGroups_inProgress.graphql";
import type { useTasksPageListGroups_todo$key } from "#/__generated__/core/useTasksPageListGroups_todo.graphql";
import type { useTasksPageListGroupsBacklogQuery } from "#/__generated__/core/useTasksPageListGroupsBacklogQuery.graphql";
import type { useTasksPageListGroupsCanceledQuery } from "#/__generated__/core/useTasksPageListGroupsCanceledQuery.graphql";
import type { useTasksPageListGroupsDoneQuery } from "#/__generated__/core/useTasksPageListGroupsDoneQuery.graphql";
import type { useTasksPageListGroupsDuplicateQuery } from "#/__generated__/core/useTasksPageListGroupsDuplicateQuery.graphql";
import type { useTasksPageListGroupsInProgressQuery } from "#/__generated__/core/useTasksPageListGroupsInProgressQuery.graphql";
import type { useTasksPageListGroupsTodoQuery } from "#/__generated__/core/useTasksPageListGroupsTodoQuery.graphql";

import type { TaskListIncludeFlags } from "./taskListIncludes";

const inProgressGroupFragment = graphql`
  fragment useTasksPageListGroups_inProgress on Viewer
  @refetchable(queryName: "useTasksPageListGroupsInProgressQuery")
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    query: { type: "String" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    inProgressTasks: assignedTasks(
      organizationId: $organizationId
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: { query: $query, state: IN_PROGRESS }
    ) @connection(key: "useTasksPageListGroups_inProgress__inProgressTasks", filters: ["organizationId", "filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TasksPageTaskListItem_task
        }
      }
    }
  }
`;

const todoGroupFragment = graphql`
  fragment useTasksPageListGroups_todo on Viewer
  @refetchable(queryName: "useTasksPageListGroupsTodoQuery")
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    query: { type: "String" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    todoTasks: assignedTasks(
      organizationId: $organizationId
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: { query: $query, state: TODO }
    ) @connection(key: "useTasksPageListGroups_todo__todoTasks", filters: ["organizationId", "filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TasksPageTaskListItem_task
        }
      }
    }
  }
`;

const backlogGroupFragment = graphql`
  fragment useTasksPageListGroups_backlog on Viewer
  @refetchable(queryName: "useTasksPageListGroupsBacklogQuery")
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    query: { type: "String" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    backlogTasks: assignedTasks(
      organizationId: $organizationId
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: { query: $query, state: BACKLOG }
    ) @connection(key: "useTasksPageListGroups_backlog__backlogTasks", filters: ["organizationId", "filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TasksPageTaskListItem_task
        }
      }
    }
  }
`;

const doneGroupFragment = graphql`
  fragment useTasksPageListGroups_done on Viewer
  @refetchable(queryName: "useTasksPageListGroupsDoneQuery")
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    query: { type: "String" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    doneTasks: assignedTasks(
      organizationId: $organizationId
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: { query: $query, state: DONE }
    ) @connection(key: "useTasksPageListGroups_done__doneTasks", filters: ["organizationId", "filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TasksPageTaskListItem_task
        }
      }
    }
  }
`;

const canceledGroupFragment = graphql`
  fragment useTasksPageListGroups_canceled on Viewer
  @refetchable(queryName: "useTasksPageListGroupsCanceledQuery")
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    query: { type: "String" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    canceledTasks: assignedTasks(
      organizationId: $organizationId
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: { query: $query, state: CANCELED }
    ) @connection(key: "useTasksPageListGroups_canceled__canceledTasks", filters: ["organizationId", "filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TasksPageTaskListItem_task
        }
      }
    }
  }
`;

const duplicateGroupFragment = graphql`
  fragment useTasksPageListGroups_duplicate on Viewer
  @refetchable(queryName: "useTasksPageListGroupsDuplicateQuery")
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    query: { type: "String" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    duplicateTasks: assignedTasks(
      organizationId: $organizationId
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: { query: $query, state: DUPLICATE }
    ) @connection(key: "useTasksPageListGroups_duplicate__duplicateTasks", filters: ["organizationId", "filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TasksPageTaskListItem_task
        }
      }
    }
  }
`;

export type TasksPageListGroupNode
  = TasksPageList_task$key & TasksPageTaskListItem_task$key;

export interface TaskListLoadNextOptions {
  onComplete?: (error: Error | null) => void;
}

export type TaskListLoadNext = (
  count?: number,
  options?: TaskListLoadNextOptions,
) => void;

export interface TasksPageListGroup {
  state: TaskState;
  connectionId: string;
  totalCount: number;
  edges: readonly { readonly node: TasksPageListGroupNode }[];
  hasNext: boolean;
  isLoadingNext: boolean;
  loadNext: TaskListLoadNext;
}

function paginationGroup(
  state: TaskState,
  assignedTasks: {
    readonly __id: string;
    readonly totalCount: number;
    readonly edges: readonly { readonly node: TasksPageListGroupNode }[];
  } | null | undefined,
  hasNext: boolean,
  isLoadingNext: boolean,
  loadNext: TaskListLoadNext,
): TasksPageListGroup {
  const edges = assignedTasks?.edges.map(edge => ({ node: edge.node })) ?? [];
  return {
    state,
    connectionId: assignedTasks?.__id ?? "",
    totalCount: assignedTasks?.totalCount ?? 0,
    edges,
    hasNext,
    isLoadingNext,
    loadNext,
  };
}

export type TasksPageListGroupsKey = {
  readonly " $fragmentSpreads": FragmentRefs<
    | "useTasksPageListGroups_backlog"
    | "useTasksPageListGroups_canceled"
    | "useTasksPageListGroups_done"
    | "useTasksPageListGroups_duplicate"
    | "useTasksPageListGroups_inProgress"
    | "useTasksPageListGroups_todo"
  >;
};

export function useTasksPageListGroups(
  viewerKey: TasksPageListGroupsKey,
  readable: TaskListIncludeFlags,
): TasksPageListGroup[] {
  const inProgress = usePaginationFragment<
    useTasksPageListGroupsInProgressQuery,
    useTasksPageListGroups_inProgress$key
  >(
    inProgressGroupFragment,
    readable.includeInProgress ? viewerKey : null,
  );
  const todo = usePaginationFragment<
    useTasksPageListGroupsTodoQuery,
    useTasksPageListGroups_todo$key
  >(
    todoGroupFragment,
    readable.includeTodo ? viewerKey : null,
  );
  const backlog = usePaginationFragment<
    useTasksPageListGroupsBacklogQuery,
    useTasksPageListGroups_backlog$key
  >(
    backlogGroupFragment,
    readable.includeBacklog ? viewerKey : null,
  );
  const done = usePaginationFragment<
    useTasksPageListGroupsDoneQuery,
    useTasksPageListGroups_done$key
  >(
    doneGroupFragment,
    readable.includeDone ? viewerKey : null,
  );
  const canceled = usePaginationFragment<
    useTasksPageListGroupsCanceledQuery,
    useTasksPageListGroups_canceled$key
  >(
    canceledGroupFragment,
    readable.includeCanceled ? viewerKey : null,
  );
  const duplicate = usePaginationFragment<
    useTasksPageListGroupsDuplicateQuery,
    useTasksPageListGroups_duplicate$key
  >(
    duplicateGroupFragment,
    readable.includeDuplicate ? viewerKey : null,
  );

  const groupsByState: Record<TaskState, TasksPageListGroup> = {
    IN_PROGRESS: paginationGroup(
      "IN_PROGRESS",
      inProgress.data?.inProgressTasks,
      inProgress.hasNext,
      inProgress.isLoadingNext,
      (count, options) => inProgress.loadNext(count ?? taskListPageSize, options),
    ),
    TODO: paginationGroup(
      "TODO",
      todo.data?.todoTasks,
      todo.hasNext,
      todo.isLoadingNext,
      (count, options) => todo.loadNext(count ?? taskListPageSize, options),
    ),
    BACKLOG: paginationGroup(
      "BACKLOG",
      backlog.data?.backlogTasks,
      backlog.hasNext,
      backlog.isLoadingNext,
      (count, options) => backlog.loadNext(count ?? taskListPageSize, options),
    ),
    DONE: paginationGroup(
      "DONE",
      done.data?.doneTasks,
      done.hasNext,
      done.isLoadingNext,
      (count, options) => done.loadNext(count ?? taskListPageSize, options),
    ),
    CANCELED: paginationGroup(
      "CANCELED",
      canceled.data?.canceledTasks,
      canceled.hasNext,
      canceled.isLoadingNext,
      (count, options) => canceled.loadNext(count ?? taskListPageSize, options),
    ),
    DUPLICATE: paginationGroup(
      "DUPLICATE",
      duplicate.data?.duplicateTasks,
      duplicate.hasNext,
      duplicate.isLoadingNext,
      (count, options) => duplicate.loadNext(count ?? taskListPageSize, options),
    ),
  };

  return taskListGroupStates.map(state => groupsByState[state]);
}
