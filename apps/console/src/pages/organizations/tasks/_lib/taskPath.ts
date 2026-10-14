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

import { type TaskState } from "@probo/tasks";
import { ConnectionHandler } from "react-relay";

export const taskListConnectionKeys = {
  IN_PROGRESS: "useOrganizationTaskListGroups_inProgress__inProgressTasks",
  TODO: "useOrganizationTaskListGroups_todo__todoTasks",
  BACKLOG: "useOrganizationTaskListGroups_backlog__backlogTasks",
  DONE: "useOrganizationTaskListGroups_done__doneTasks",
  CANCELED: "useOrganizationTaskListGroups_canceled__canceledTasks",
  DUPLICATE: "useOrganizationTaskListGroups_duplicate__duplicateTasks",
} as const satisfies Record<TaskState, string>;

export interface OrganizationTaskListFilter {
  query: string | null;
  assignedToId: string | null;
  internalControlId: string | null;
}

export function organizationTaskListFilterVariants(
  filter: OrganizationTaskListFilter,
): OrganizationTaskListFilter[] {
  const empty = {
    query: null,
    assignedToId: null,
    internalControlId: null,
  };
  if (
    filter.query == null
    && filter.assignedToId == null
    && filter.internalControlId == null
  ) {
    return [empty];
  }

  return [
    {
      query: filter.query,
      assignedToId: filter.assignedToId,
      internalControlId: filter.internalControlId,
    },
    empty,
  ];
}

export function organizationTaskListConnectionFilters(
  filter: OrganizationTaskListFilter,
  state: TaskState,
) {
  return {
    filter: {
      query: filter.query,
      state,
      assignedToId: filter.assignedToId,
      internalControlId: filter.internalControlId,
    },
  };
}

export function organizationTaskListConnectionId(
  organizationId: string,
  filter: OrganizationTaskListFilter,
  state: TaskState,
) {
  return ConnectionHandler.getConnectionID(
    organizationId,
    taskListConnectionKeys[state],
    organizationTaskListConnectionFilters(filter, state),
  );
}

export const internalControlTasksConnectionKey = "InternalControl__tasks";
export const taskInternalControlIdsConnectionKey = "useTaskInternalControlIds_internalControls";

export const emptyTaskListFilter = {
  query: null,
  state: null,
  assignedToId: null,
  internalControlId: null,
};

const taskListConnectionFilters = {
  orderBy: { field: "PRIORITY_RANK" as const, direction: "ASC" as const },
  filter: emptyTaskListFilter,
};

export function taskListPath(organizationId: string) {
  return `/organizations/${organizationId}/governance/tasks`;
}

export function taskDetailsPath(organizationId: string, taskId: string) {
  return `${taskListPath(organizationId)}/${taskId}`;
}

export function taskConnectionId(parentId: string, connectionKey: string) {
  return ConnectionHandler.getConnectionID(
    parentId,
    connectionKey,
    taskListConnectionFilters,
  );
}

export const taskCommentsConnectionKey = "TaskCommentsSection_comments";

const taskCommentsConnectionFilters = {
  orderBy: { field: "CREATED_AT" as const, direction: "ASC" as const },
};

export function taskCommentsConnectionId(taskId: string) {
  return ConnectionHandler.getConnectionID(
    taskId,
    taskCommentsConnectionKey,
    taskCommentsConnectionFilters,
  );
}
