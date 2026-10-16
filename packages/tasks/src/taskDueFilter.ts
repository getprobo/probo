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

export const taskDueFilters = [
  "overdue",
  "1d",
  "3d",
  "1w",
  "1m",
  "3m",
  "none",
] as const;

export type TaskDueFilter = (typeof taskDueFilters)[number];

export interface TaskDeadlineBounds {
  deadlineOnOrBefore: string | null;
  hasDeadline: boolean | null;
}

export const emptyDeadlineBounds: TaskDeadlineBounds = {
  deadlineOnOrBefore: null,
  hasDeadline: null,
};

export function taskDueFilterFromParam(value: string | null): TaskDueFilter | null {
  return taskDueFilters.find(filter => filter === value) ?? null;
}

export function taskDueFilterBounds(
  filter: TaskDueFilter | null,
  now = new Date(),
): TaskDeadlineBounds {
  if (filter == null) {
    return emptyDeadlineBounds;
  }
  if (filter === "none") {
    return { deadlineOnOrBefore: null, hasDeadline: false };
  }

  return {
    deadlineOnOrBefore: `${dueDay(filter, now)}T00:00:00Z`,
    hasDeadline: null,
  };
}

export function taskMatchesDeadline(
  deadline: string | null | undefined,
  filter: TaskDeadlineBounds,
): boolean | null {
  if (filter.hasDeadline === false) {
    if (deadline === undefined) {
      return null;
    }
    return deadline == null;
  }
  if (filter.deadlineOnOrBefore == null) {
    return true;
  }
  if (deadline == null) {
    return deadline === undefined ? null : false;
  }

  return deadline.slice(0, 10) <= filter.deadlineOnOrBefore.slice(0, 10);
}

function dueDay(filter: Exclude<TaskDueFilter, "none">, now: Date): string {
  const date = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  switch (filter) {
    case "overdue":
      date.setDate(date.getDate() - 1);
      break;
    case "1d":
      date.setDate(date.getDate() + 1);
      break;
    case "3d":
      date.setDate(date.getDate() + 3);
      break;
    case "1w":
      date.setDate(date.getDate() + 7);
      break;
    case "1m":
      date.setMonth(date.getMonth() + 1);
      break;
    case "3m":
      date.setMonth(date.getMonth() + 3);
      break;
  }

  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}
