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

import type { ReactNode } from "react";

import { type TaskDueFilter } from "../taskDueFilter";
import { type TaskState } from "../taskState";
import { taskListToolbar } from "../variants";

import { TaskListDueDateFilter } from "./TaskListDueDateFilter";
import { TaskListSearchField } from "./TaskListSearchField";
import { TaskListStatusFilter } from "./TaskListStatusFilter";

export interface TaskListToolbarProps {
  query: string;
  onQueryChange: (value: string) => void;
  state: TaskState | null;
  onStateChange: (value: TaskState | null) => void;
  due: TaskDueFilter | null;
  onDueChange: (value: TaskDueFilter | null) => void;
  children?: ReactNode;
}

export function TaskListToolbar({
  query,
  onQueryChange,
  state,
  onStateChange,
  due,
  onDueChange,
  children,
}: TaskListToolbarProps) {
  const { tools, filters } = taskListToolbar();

  return (
    <div className={tools()}>
      <TaskListSearchField value={query} onValueChange={onQueryChange} />
      <div className={filters()}>
        <TaskListStatusFilter value={state} onValueChange={onStateChange} />
        <TaskListDueDateFilter value={due} onValueChange={onDueChange} />
        {children}
      </div>
    </div>
  );
}
