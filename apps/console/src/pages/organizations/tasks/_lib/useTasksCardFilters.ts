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

import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";

import { type TaskState, taskStates } from "./taskState";

export interface TasksCardFilterInput {
  query: string | null;
  state: TaskState | null;
}

function taskStateFromParam(value: string | null): TaskState | null {
  return taskStates.find(
    state => state.toLowerCase().replace("_", "-") === value,
  ) ?? null;
}

export function taskMatchesFilter(
  task: { name: string; state: TaskState },
  filter: TasksCardFilterInput,
): boolean {
  if (filter.state != null && task.state !== filter.state) {
    return false;
  }
  if (filter.query == null) {
    return true;
  }

  return task.name.toLowerCase().includes(filter.query.toLowerCase());
}

export function useTasksCardFilters() {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const state = taskStateFromParam(searchParams.get("status"));
  const graphqlFilter = useMemo<TasksCardFilterInput>(() => ({
    query: query === "" ? null : query,
    state,
  }), [query, state]);

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

  const setState = useCallback((value: TaskState | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value == null) {
        next.delete("status");
      } else {
        next.set("status", value.toLowerCase().replace("_", "-"));
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  return {
    query,
    state,
    graphqlFilter,
    setQuery,
    setState,
  };
}
