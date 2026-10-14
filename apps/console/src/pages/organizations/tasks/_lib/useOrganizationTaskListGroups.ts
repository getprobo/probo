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

import type { TaskListItem_task$key } from "#/__generated__/core/TaskListItem_task.graphql";
import type { TasksPageList_task$key } from "#/__generated__/core/TasksPageList_task.graphql";
import type { useOrganizationTaskListGroups_backlog$key } from "#/__generated__/core/useOrganizationTaskListGroups_backlog.graphql";
import type { useOrganizationTaskListGroups_canceled$key } from "#/__generated__/core/useOrganizationTaskListGroups_canceled.graphql";
import type { useOrganizationTaskListGroups_done$key } from "#/__generated__/core/useOrganizationTaskListGroups_done.graphql";
import type { useOrganizationTaskListGroups_duplicate$key } from "#/__generated__/core/useOrganizationTaskListGroups_duplicate.graphql";
import type { useOrganizationTaskListGroups_inProgress$key } from "#/__generated__/core/useOrganizationTaskListGroups_inProgress.graphql";
import type { useOrganizationTaskListGroups_todo$key } from "#/__generated__/core/useOrganizationTaskListGroups_todo.graphql";
import type { useOrganizationTaskListGroupsBacklogQuery } from "#/__generated__/core/useOrganizationTaskListGroupsBacklogQuery.graphql";
import type { useOrganizationTaskListGroupsCanceledQuery } from "#/__generated__/core/useOrganizationTaskListGroupsCanceledQuery.graphql";
import type { useOrganizationTaskListGroupsDoneQuery } from "#/__generated__/core/useOrganizationTaskListGroupsDoneQuery.graphql";
import type { useOrganizationTaskListGroupsDuplicateQuery } from "#/__generated__/core/useOrganizationTaskListGroupsDuplicateQuery.graphql";
import type { useOrganizationTaskListGroupsInProgressQuery } from "#/__generated__/core/useOrganizationTaskListGroupsInProgressQuery.graphql";
import type { useOrganizationTaskListGroupsTodoQuery } from "#/__generated__/core/useOrganizationTaskListGroupsTodoQuery.graphql";

import type { TaskListIncludeFlags } from "./taskListIncludes";

const inProgressGroupFragment = graphql`
  fragment useOrganizationTaskListGroups_inProgress on Organization
  @refetchable(queryName: "useOrganizationTaskListGroupsInProgressQuery")
  @argumentDefinitions(
    query: { type: "String" }
    assignedToId: { type: "ID" }
    internalControlId: { type: "ID" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    inProgressTasks: tasks(
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: {
        query: $query
        state: IN_PROGRESS
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      }
    ) @connection(key: "useOrganizationTaskListGroups_inProgress__inProgressTasks", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TaskListItem_task
        }
      }
    }
  }
`;

const todoGroupFragment = graphql`
  fragment useOrganizationTaskListGroups_todo on Organization
  @refetchable(queryName: "useOrganizationTaskListGroupsTodoQuery")
  @argumentDefinitions(
    query: { type: "String" }
    assignedToId: { type: "ID" }
    internalControlId: { type: "ID" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    todoTasks: tasks(
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: {
        query: $query
        state: TODO
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      }
    ) @connection(key: "useOrganizationTaskListGroups_todo__todoTasks", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TaskListItem_task
        }
      }
    }
  }
`;

const backlogGroupFragment = graphql`
  fragment useOrganizationTaskListGroups_backlog on Organization
  @refetchable(queryName: "useOrganizationTaskListGroupsBacklogQuery")
  @argumentDefinitions(
    query: { type: "String" }
    assignedToId: { type: "ID" }
    internalControlId: { type: "ID" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    backlogTasks: tasks(
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: {
        query: $query
        state: BACKLOG
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      }
    ) @connection(key: "useOrganizationTaskListGroups_backlog__backlogTasks", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TaskListItem_task
        }
      }
    }
  }
`;

const doneGroupFragment = graphql`
  fragment useOrganizationTaskListGroups_done on Organization
  @refetchable(queryName: "useOrganizationTaskListGroupsDoneQuery")
  @argumentDefinitions(
    query: { type: "String" }
    assignedToId: { type: "ID" }
    internalControlId: { type: "ID" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    doneTasks: tasks(
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: {
        query: $query
        state: DONE
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      }
    ) @connection(key: "useOrganizationTaskListGroups_done__doneTasks", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TaskListItem_task
        }
      }
    }
  }
`;

const canceledGroupFragment = graphql`
  fragment useOrganizationTaskListGroups_canceled on Organization
  @refetchable(queryName: "useOrganizationTaskListGroupsCanceledQuery")
  @argumentDefinitions(
    query: { type: "String" }
    assignedToId: { type: "ID" }
    internalControlId: { type: "ID" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    canceledTasks: tasks(
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: {
        query: $query
        state: CANCELED
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      }
    ) @connection(key: "useOrganizationTaskListGroups_canceled__canceledTasks", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TaskListItem_task
        }
      }
    }
  }
`;

const duplicateGroupFragment = graphql`
  fragment useOrganizationTaskListGroups_duplicate on Organization
  @refetchable(queryName: "useOrganizationTaskListGroupsDuplicateQuery")
  @argumentDefinitions(
    query: { type: "String" }
    assignedToId: { type: "ID" }
    internalControlId: { type: "ID" }
    first: { type: "Int", defaultValue: 40 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    duplicateTasks: tasks(
      first: $first
      after: $after
      orderBy: { field: PRIORITY_RANK, direction: ASC }
      filter: {
        query: $query
        state: DUPLICATE
        assignedToId: $assignedToId
        internalControlId: $internalControlId
      }
    ) @connection(key: "useOrganizationTaskListGroups_duplicate__duplicateTasks", filters: ["filter"]) {
      __id
      totalCount
      edges {
        node {
          ...TasksPageList_task
          ...TaskListItem_task
        }
      }
    }
  }
`;

export type OrganizationTaskListGroupNode
  = TasksPageList_task$key & TaskListItem_task$key;

export interface OrganizationTaskListLoadNextOptions {
  onComplete?: (error: Error | null) => void;
}

export type OrganizationTaskListLoadNext = (
  count?: number,
  options?: OrganizationTaskListLoadNextOptions,
) => void;

export interface OrganizationTaskListGroup {
  state: TaskState;
  connectionId: string;
  totalCount: number;
  edges: readonly { readonly node: OrganizationTaskListGroupNode }[];
  hasNext: boolean;
  isLoadingNext: boolean;
  loadNext: OrganizationTaskListLoadNext;
}

function paginationGroup(
  state: TaskState,
  tasks: {
    readonly __id: string;
    readonly totalCount: number;
    readonly edges: readonly { readonly node: OrganizationTaskListGroupNode }[];
  } | null | undefined,
  hasNext: boolean,
  isLoadingNext: boolean,
  loadNext: OrganizationTaskListLoadNext,
): OrganizationTaskListGroup {
  return {
    state,
    connectionId: tasks?.__id ?? "",
    totalCount: tasks?.totalCount ?? 0,
    edges: tasks?.edges.map(edge => ({ node: edge.node })) ?? [],
    hasNext,
    isLoadingNext,
    loadNext,
  };
}

export type OrganizationTaskListGroupsKey = {
  readonly " $fragmentSpreads": FragmentRefs<
    | "useOrganizationTaskListGroups_backlog"
    | "useOrganizationTaskListGroups_canceled"
    | "useOrganizationTaskListGroups_done"
    | "useOrganizationTaskListGroups_duplicate"
    | "useOrganizationTaskListGroups_inProgress"
    | "useOrganizationTaskListGroups_todo"
  >;
};

export function useOrganizationTaskListGroups(
  organizationKey: OrganizationTaskListGroupsKey,
  readable: TaskListIncludeFlags,
): OrganizationTaskListGroup[] {
  const inProgress = usePaginationFragment<
    useOrganizationTaskListGroupsInProgressQuery,
    useOrganizationTaskListGroups_inProgress$key
  >(
    inProgressGroupFragment,
    readable.includeInProgress ? organizationKey : null,
  );
  const todo = usePaginationFragment<
    useOrganizationTaskListGroupsTodoQuery,
    useOrganizationTaskListGroups_todo$key
  >(
    todoGroupFragment,
    readable.includeTodo ? organizationKey : null,
  );
  const backlog = usePaginationFragment<
    useOrganizationTaskListGroupsBacklogQuery,
    useOrganizationTaskListGroups_backlog$key
  >(
    backlogGroupFragment,
    readable.includeBacklog ? organizationKey : null,
  );
  const done = usePaginationFragment<
    useOrganizationTaskListGroupsDoneQuery,
    useOrganizationTaskListGroups_done$key
  >(
    doneGroupFragment,
    readable.includeDone ? organizationKey : null,
  );
  const canceled = usePaginationFragment<
    useOrganizationTaskListGroupsCanceledQuery,
    useOrganizationTaskListGroups_canceled$key
  >(
    canceledGroupFragment,
    readable.includeCanceled ? organizationKey : null,
  );
  const duplicate = usePaginationFragment<
    useOrganizationTaskListGroupsDuplicateQuery,
    useOrganizationTaskListGroups_duplicate$key
  >(
    duplicateGroupFragment,
    readable.includeDuplicate ? organizationKey : null,
  );

  const groupsByState: Record<TaskState, OrganizationTaskListGroup> = {
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
