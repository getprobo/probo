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

import type { TaskListFlatItem, TaskState } from "@probo/tasks";
import {
  type TransitionStartFunction,
  useCallback,
  useEffect,
  useRef,
} from "react";

import { taskListPageSize } from "./taskListPageSize";
import type { TasksPageListGroup } from "./useTasksPageListGroups";

export interface TasksListFillRange {
  startIndex: number;
  endIndex: number;
}

export interface UseTasksListFillOptions {
  items: readonly TaskListFlatItem<unknown>[];
  groups: readonly TasksPageListGroup[];
  range: TasksListFillRange | null;
  closedStates: ReadonlySet<TaskState>;
  selectedState: TaskState | null;
  startTransition: TransitionStartFunction;
}

function rangeIsUncovered(
  items: readonly TaskListFlatItem<unknown>[],
  range: TasksListFillRange,
): boolean {
  if (range.endIndex >= items.length) {
    return true;
  }

  const start = Math.max(0, range.startIndex);
  const end = Math.min(range.endIndex, items.length - 1);
  for (let index = start; index <= end; index += 1) {
    if (items[index]?.type === "placeholder") {
      return true;
    }
  }

  return false;
}

function pickFillGroup(
  items: readonly TaskListFlatItem<unknown>[],
  groups: readonly TasksPageListGroup[],
  range: TasksListFillRange,
  closedStates: ReadonlySet<TaskState>,
  selectedState: TaskState | null,
  skipped: ReadonlySet<TaskState>,
): TasksPageListGroup | null {
  const start = Math.max(0, range.startIndex);
  const end = range.endIndex;

  for (let index = start; index <= end && index < items.length; index += 1) {
    const item = items[index];
    if (item?.type !== "placeholder" || skipped.has(item.state)) {
      continue;
    }
    if (closedStates.has(item.state)) {
      continue;
    }
    const group = groups.find(candidate => candidate.state === item.state);
    if (group != null && group.hasNext) {
      return group;
    }
  }

  if (selectedState != null) {
    return null;
  }

  const lastIndex = Math.min(Math.max(end, 0), Math.max(items.length - 1, 0));
  const lastState = items[lastIndex]?.state;
  const lastGroupIndex = lastState == null
    ? -1
    : groups.findIndex(group => group.state === lastState);

  for (const group of groups.slice(lastGroupIndex + 1)) {
    if (skipped.has(group.state) || closedStates.has(group.state)) {
      continue;
    }
    if (group.totalCount <= 0 || !group.hasNext) {
      continue;
    }
    return group;
  }

  return null;
}

function loadGroupPage(
  group: TasksPageListGroup,
  startTransition: TransitionStartFunction,
): Promise<void> {
  return new Promise((resolve) => {
    startTransition(() => {
      group.loadNext(taskListPageSize, {
        onComplete() {
          resolve();
        },
      });
    });
  });
}

export function useTasksListFill({
  items,
  groups,
  range,
  closedStates,
  selectedState,
  startTransition,
}: UseTasksListFillOptions): (nextRange?: TasksListFillRange | null) => Promise<void> {
  const fillingRef = useRef(false);
  const skippedRef = useRef(new Set<TaskState>());
  const lastAttemptRef = useRef<{ state: TaskState; edges: number } | null>(null);
  const snapshotRef = useRef({ items, groups, range, closedStates, selectedState });
  const rangeKey = range == null ? "" : `${range.startIndex}:${range.endIndex}`;
  const groupsKey = groups
    .map(group => `${group.state}:${group.edges.length}:${group.hasNext}`)
    .join("|");

  useEffect(() => {
    snapshotRef.current = { items, groups, range, closedStates, selectedState };
  }, [closedStates, groups, items, range, selectedState]);

  useEffect(() => {
    skippedRef.current = new Set();
    lastAttemptRef.current = null;
  }, [rangeKey, selectedState]);

  const fill = useCallback(async (nextRange?: TasksListFillRange | null) => {
    if (fillingRef.current) {
      return;
    }

    const snapshot = snapshotRef.current;
    const currentRange = nextRange ?? snapshot.range;
    if (currentRange == null || !rangeIsUncovered(snapshot.items, currentRange)) {
      return;
    }

    const skipped = skippedRef.current;
    const lastAttempt = lastAttemptRef.current;
    if (
      lastAttempt != null
      && snapshot.groups.some(group =>
        group.state === lastAttempt.state
        && group.edges.length === lastAttempt.edges,
      )
    ) {
      skipped.add(lastAttempt.state);
    }

    const group = pickFillGroup(
      snapshot.items,
      snapshot.groups,
      currentRange,
      snapshot.closedStates,
      snapshot.selectedState,
      skipped,
    );
    if (group == null) {
      return;
    }

    fillingRef.current = true;
    lastAttemptRef.current = {
      state: group.state,
      edges: group.edges.length,
    };
    try {
      await loadGroupPage(group, startTransition);
    } finally {
      fillingRef.current = false;
    }
  }, [startTransition]);

  useEffect(() => {
    void fill();
  }, [closedStates, fill, groupsKey, items.length, rangeKey, selectedState]);

  return fill;
}
