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

import { taskDueFilterFromParam, taskStates } from "@probo/tasks";

const taskListFilterPreferenceKeys = ["status", "assignee", "due"] as const;

const maxAssigneeParamLength = 200;

export interface TaskListFilterPreference {
  status: string | null;
  assignee: string | null;
  due: string | null;
}

export interface RestoredTaskListFilters {
  params: URLSearchParams;
  remember: TaskListFilterPreference | null;
}

export function taskListFilterPreferenceInParams(params: URLSearchParams): boolean {
  return taskListFilterPreferenceKeys.some(key => params.has(key));
}

export function defaultTaskListFilterPreference(profileId: string): TaskListFilterPreference | null {
  const assignee = assigneeParam(profileId);
  if (assignee == null) {
    return null;
  }

  return {
    status: null,
    assignee,
    due: null,
  };
}

export function taskListFilterPreferenceFromParams(params: URLSearchParams): TaskListFilterPreference {
  return {
    status: statusParam(params.get("status")),
    assignee: assigneeParam(params.get("assignee")),
    due: dueParam(params.get("due")),
  };
}

export function applyTaskListFilterPreference(
  params: URLSearchParams,
  preference: TaskListFilterPreference,
): URLSearchParams {
  const next = new URLSearchParams(params);
  for (const key of taskListFilterPreferenceKeys) {
    const value = preference[key];
    if (value == null || value === "") {
      next.delete(key);
    } else {
      next.set(key, value);
    }
  }
  return next;
}

// A shared link that already sets a filter is left alone. The first bare
// visit is the current user's tasks, and that choice is what gets saved.
export function restoredTaskListFilters(
  params: URLSearchParams,
  stored: TaskListFilterPreference | null,
  profileId: string,
): RestoredTaskListFilters | null {
  if (taskListFilterPreferenceInParams(params)) {
    return null;
  }

  const preference = stored ?? defaultTaskListFilterPreference(profileId);
  if (preference == null) {
    return null;
  }

  return {
    params: applyTaskListFilterPreference(params, preference),
    remember: stored == null ? preference : null,
  };
}

export function readTaskListFilterPreference(
  organizationId: string,
  profileId: string,
): TaskListFilterPreference | null {
  const key = storageKey(organizationId, profileId);
  if (key == null) {
    return null;
  }

  try {
    const raw = localStorage.getItem(key);
    if (raw == null) {
      return null;
    }
    return taskListFilterPreferenceFromUnknown(JSON.parse(raw));
  } catch {
    return null;
  }
}

export function writeTaskListFilterPreference(
  organizationId: string,
  profileId: string,
  preference: TaskListFilterPreference,
): void {
  const key = storageKey(organizationId, profileId);
  if (key == null) {
    return;
  }

  try {
    localStorage.setItem(key, JSON.stringify({
      status: preference.status,
      assignee: preference.assignee,
      due: preference.due,
    }));
  } catch {
    // Storage can be unavailable in private mode or when the quota is full.
  }
}

function storageKey(organizationId: string, profileId: string): string | null {
  if (organizationId === "" || profileId === "") {
    return null;
  }

  return `probo.console.tasks.filters.v1.${organizationId}.${profileId}`;
}

function taskListFilterPreferenceFromUnknown(value: unknown): TaskListFilterPreference | null {
  if (typeof value !== "object" || value == null || Array.isArray(value)) {
    return null;
  }

  const record = value as Record<string, unknown>;
  return {
    status: statusParam(stringOrNull(record.status)),
    assignee: assigneeParam(stringOrNull(record.assignee)),
    due: dueParam(stringOrNull(record.due)),
  };
}

function stringOrNull(value: unknown): string | null {
  return typeof value === "string" ? value : null;
}

function statusParam(value: string | null): string | null {
  if (value == null || value === "") {
    return null;
  }

  const match = taskStates.find(
    state => state.toLowerCase().replace("_", "-") === value,
  );
  return match == null ? null : value;
}

function dueParam(value: string | null): string | null {
  return taskDueFilterFromParam(value);
}

function assigneeParam(value: string | null): string | null {
  if (value == null || value === "" || value.length > maxAssigneeParamLength) {
    return null;
  }

  return value;
}
