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

export const taskCommentsConnectionKey = "TaskCommentsSection_comments";

export const taskListConnectionKeys = {
  IN_PROGRESS: "useTasksPageListGroups_inProgress__inProgressTasks",
  TODO: "useTasksPageListGroups_todo__todoTasks",
  BACKLOG: "useTasksPageListGroups_backlog__backlogTasks",
  DONE: "useTasksPageListGroups_done__doneTasks",
  CANCELED: "useTasksPageListGroups_canceled__canceledTasks",
  DUPLICATE: "useTasksPageListGroups_duplicate__duplicateTasks",
} as const satisfies Record<TaskState, string>;

export function taskListConnectionFilters(
  organizationId: string,
  query: string | null,
  state: TaskState,
) {
  return {
    organizationId,
    filter: {
      query,
      state,
    },
  };
}

export function taskListConnectionId(
  viewerId: string,
  organizationId: string,
  query: string | null,
  state: TaskState,
) {
  return ConnectionHandler.getConnectionID(
    viewerId,
    taskListConnectionKeys[state],
    taskListConnectionFilters(organizationId, query, state),
  );
}

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
