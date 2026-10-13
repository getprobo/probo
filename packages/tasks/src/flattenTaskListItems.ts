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

import type { TaskState } from "./taskState";

export type TaskListRowPosition = "first" | "middle" | "last" | "only";

export interface TaskListGroupInput<T> {
  state: TaskState;
  totalCount: number;
  loadedItems: readonly T[];
  open: boolean;
}

export type TaskListFlatItem<T>
  = | { type: "header"; state: TaskState; count: number; offset: boolean }
    | { type: "row"; state: TaskState; item: T; position: TaskListRowPosition }
    | { type: "placeholder"; state: TaskState; position: TaskListRowPosition };

export function taskListRowPosition(
  index: number,
  totalCount: number,
): TaskListRowPosition {
  if (totalCount <= 1) {
    return "only";
  }
  if (index === 0) {
    return "first";
  }
  if (index === totalCount - 1) {
    return "last";
  }

  return "middle";
}

export function flattenTaskListItems<T>(
  groups: readonly TaskListGroupInput<T>[],
): TaskListFlatItem<T>[] {
  const items: TaskListFlatItem<T>[] = [];

  for (const group of groups) {
    if (group.totalCount <= 0) {
      continue;
    }

    items.push({
      type: "header",
      state: group.state,
      count: group.totalCount,
      offset: items.length > 0,
    });

    if (!group.open) {
      continue;
    }

    const loadedCount = Math.min(group.loadedItems.length, group.totalCount);
    for (let index = 0; index < loadedCount; index += 1) {
      items.push({
        type: "row",
        state: group.state,
        item: group.loadedItems[index],
        position: taskListRowPosition(index, group.totalCount),
      });
    }
    for (let index = loadedCount; index < group.totalCount; index += 1) {
      items.push({
        type: "placeholder",
        state: group.state,
        position: taskListRowPosition(index, group.totalCount),
      });
    }
  }

  return items;
}
