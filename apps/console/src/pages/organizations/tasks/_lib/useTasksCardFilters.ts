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
  type TaskDueFilter,
  taskDueFilterBounds,
  taskDueFilterFromParam,
  taskMatchesDeadline,
  type TaskState,
  taskStates,
} from "@probo/tasks";
import { use, useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";

import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CurrentUser } from "#/providers/CurrentUser";

import {
  taskListFilterPreferenceFromParams,
  writeTaskListFilterPreference,
} from "./taskListFilterPreference";
import { unassignedAssignee } from "./taskPath";

export interface TasksCardFilterInput {
  query: string | null;
  state: TaskState | null;
  assignedToId: string | null;
  internalControlId: string | null;
  deadlineOnOrBefore: string | null;
  deadlineOnOrAfter: string | null;
  hasDeadline: boolean | null;
  hasAssignee: boolean | null;
}

function taskStateFromParam(value: string | null): TaskState | null {
  return taskStates.find(
    state => state.toLowerCase().replace("_", "-") === value,
  ) ?? null;
}

export function taskMatchesFilter(
  task: {
    name: string;
    state: TaskState;
    assignedToId?: string | null;
    internalControlIds?: readonly string[];
    deadline?: string | null;
  },
  filter: TasksCardFilterInput,
): boolean {
  if (filter.state != null && task.state !== filter.state) {
    return false;
  }
  if (filter.assignedToId != null && task.assignedToId !== filter.assignedToId) {
    return false;
  }
  if (filter.hasAssignee === false && task.assignedToId != null) {
    return false;
  }
  if (
    filter.internalControlId != null
    && !task.internalControlIds?.includes(filter.internalControlId)
  ) {
    return false;
  }
  if (filter.query != null && !task.name.toLowerCase().includes(filter.query.toLowerCase())) {
    return false;
  }

  return taskMatchesDeadline(task.deadline ?? null, filter) !== false;
}

export function useTasksCardFilters() {
  const organizationId = useOrganizationId();
  const { profile } = use(CurrentUser);
  const profileId = profile?.id ?? "";
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const state = taskStateFromParam(searchParams.get("status"));
  const assignedToParam = searchParams.get("assignee");
  const assignee = assignedToParam === "" ? null : assignedToParam;
  const assignedToId = assignee == null || assignee === unassignedAssignee
    ? null
    : assignee;
  const hasAssignee = assignee === unassignedAssignee ? false : null;
  const internalControlParam = searchParams.get("internalControl");
  const internalControlId = internalControlParam === "" ? null : internalControlParam;
  const due = taskDueFilterFromParam(searchParams.get("due"));
  const deadlineBounds = useMemo(() => taskDueFilterBounds(due), [due]);
  const graphqlFilter = useMemo<TasksCardFilterInput>(() => ({
    query: query === "" ? null : query,
    state,
    assignedToId,
    internalControlId,
    deadlineOnOrBefore: deadlineBounds.deadlineOnOrBefore,
    deadlineOnOrAfter: deadlineBounds.deadlineOnOrAfter,
    hasDeadline: deadlineBounds.hasDeadline,
    hasAssignee,
  }), [assignedToId, deadlineBounds, hasAssignee, internalControlId, query, state]);

  const setQuery = useCallback((value: string) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value === "") {
        next.delete("q");
      } else {
        next.set("q", value);
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const commitFilters = useCallback((update: (params: URLSearchParams) => void) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      update(next);
      writeTaskListFilterPreference(
        organizationId,
        profileId,
        taskListFilterPreferenceFromParams(next),
      );
      return next;
    }, { replace: true });
  }, [organizationId, profileId, setSearchParams]);

  const setState = useCallback((value: TaskState | null) => {
    commitFilters((params) => {
      if (value == null) {
        params.delete("status");
      } else {
        params.set("status", value.toLowerCase().replace("_", "-"));
      }
    });
  }, [commitFilters]);

  const setAssignedToId = useCallback((value: string | null) => {
    commitFilters((params) => {
      if (value == null || value === "") {
        params.delete("assignee");
      } else {
        params.set("assignee", value);
      }
    });
  }, [commitFilters]);

  const setInternalControlId = useCallback((value: string | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value == null || value === "") {
        next.delete("internalControl");
      } else {
        next.set("internalControl", value);
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setDue = useCallback((value: TaskDueFilter | null) => {
    commitFilters((params) => {
      if (value == null) {
        params.delete("due");
      } else {
        params.set("due", value);
      }
    });
  }, [commitFilters]);

  return {
    query,
    state,
    assignee,
    assignedToId,
    internalControlId,
    due,
    graphqlFilter,
    setQuery,
    setState,
    setAssignedToId,
    setInternalControlId,
    setDue,
  };
}
