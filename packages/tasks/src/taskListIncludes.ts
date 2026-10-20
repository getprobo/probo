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

import { taskListGroupStates, type TaskState } from "./taskState";

export const taskListIncludeFlagKeys = {
  IN_PROGRESS: "includeInProgress",
  TODO: "includeTodo",
  BACKLOG: "includeBacklog",
  DONE: "includeDone",
  CANCELED: "includeCanceled",
  DUPLICATE: "includeDuplicate",
} as const satisfies Record<TaskState, `include${string}`>;

export type TaskListIncludeFlagKey = (typeof taskListIncludeFlagKeys)[TaskState];

export type TaskListIncludeFlags = Record<TaskListIncludeFlagKey, boolean>;

export function taskListIncludeFlags(state: TaskState | null): TaskListIncludeFlags {
  const flags = {} as TaskListIncludeFlags;
  for (const groupState of taskListGroupStates) {
    flags[taskListIncludeFlagKeys[groupState]] = state == null || state === groupState;
  }
  return flags;
}

export function pickTaskListIncludeFlags(
  variables: TaskListIncludeFlags,
): TaskListIncludeFlags {
  const flags = {} as TaskListIncludeFlags;
  for (const groupState of taskListGroupStates) {
    const key = taskListIncludeFlagKeys[groupState];
    flags[key] = variables[key];
  }
  return flags;
}

export function mergeTaskListIncludeFlags(
  current: TaskListIncludeFlags,
  next: TaskListIncludeFlags,
): TaskListIncludeFlags {
  const flags = {} as TaskListIncludeFlags;
  for (const groupState of taskListGroupStates) {
    const key = taskListIncludeFlagKeys[groupState];
    flags[key] = current[key] || next[key];
  }
  return flags;
}

export function taskListIncludeFlagsEqual(
  left: TaskListIncludeFlags,
  right: TaskListIncludeFlags,
): boolean {
  return taskListGroupStates.every((groupState) => {
    const key = taskListIncludeFlagKeys[groupState];
    return left[key] === right[key];
  });
}

export function taskListNeedsFetch(
  query: string | null,
  includes: TaskListIncludeFlags,
  fetched: { query: string | null; includes: TaskListIncludeFlags },
): boolean {
  if (query !== fetched.query) {
    return true;
  }

  return taskListGroupStates.some((groupState) => {
    const key = taskListIncludeFlagKeys[groupState];
    return includes[key] && !fetched.includes[key];
  });
}

export function taskListReadableIncludeFlags(
  needed: TaskListIncludeFlags,
  fetched: TaskListIncludeFlags,
): TaskListIncludeFlags {
  const flags = {} as TaskListIncludeFlags;
  for (const groupState of taskListGroupStates) {
    const key = taskListIncludeFlagKeys[groupState];
    flags[key] = needed[key] && fetched[key];
  }
  return flags;
}
