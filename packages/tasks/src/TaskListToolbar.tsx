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

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "./i18n";
import { type TaskDueFilter, taskDueFilters } from "./taskDueFilter";
import { taskListGroupStates, type TaskState, taskStateKeys } from "./taskState";
import { TaskStateIcon } from "./TaskStateIcon";
import { taskListToolbar } from "./variants";

export interface TaskListSearchFieldProps {
  value: string;
  onValueChange: (value: string) => void;
}

export function TaskListSearchField({
  value,
  onValueChange,
}: TaskListSearchFieldProps) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const { search } = taskListToolbar();

  return (
    <div className={search()}>
      <TextField
        icon={<MagnifyingGlassIcon />}
        value={value}
        onValueChange={onValueChange}
        placeholder={t("searchPlaceholder")}
        aria-label={t("searchPlaceholder")}
      />
    </div>
  );
}

export interface TaskListStatusFilterProps {
  value: TaskState | null;
  onValueChange: (value: TaskState | null) => void;
}

export function TaskListStatusFilter({
  value,
  onValueChange,
}: TaskListStatusFilterProps) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const { filter, stateOption } = taskListToolbar();

  return (
    <div className={filter()}>
      <Select<TaskState | null>
        value={value}
        onValueChange={onValueChange}
      >
        <SelectTrigger
          size={2}
          placeholder={t("filters.allStatuses")}
          aria-label={t("filters.status")}
        >
          {(selected: TaskState | null) => selected == null
            ? t("filters.allStatuses")
            : (
                <span className={stateOption()}>
                  <TaskStateIcon state={selected} />
                  {t(`states.${taskStateKeys[selected]}`)}
                </span>
              )}
        </SelectTrigger>
        <SelectPopup align="start">
          <SelectItem value={null}>
            {t("filters.allStatuses")}
          </SelectItem>
          {taskListGroupStates.map(state => (
            <SelectItem key={state} value={state}>
              <span className={stateOption()}>
                <TaskStateIcon state={state} />
                {t(`states.${taskStateKeys[state]}`)}
              </span>
            </SelectItem>
          ))}
        </SelectPopup>
      </Select>
    </div>
  );
}

export interface TaskListToolbarProps {
  query: string;
  onQueryChange: (value: string) => void;
  state: TaskState | null;
  onStateChange: (value: TaskState | null) => void;
  due: TaskDueFilter | null;
  onDueChange: (value: TaskDueFilter | null) => void;
  children?: ReactNode;
}

export interface TaskListDueDateFilterProps {
  value: TaskDueFilter | null;
  onValueChange: (value: TaskDueFilter | null) => void;
}

export function TaskListDueDateFilter({
  value,
  onValueChange,
}: TaskListDueDateFilterProps) {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const { dueFilter } = taskListToolbar();

  return (
    <div className={dueFilter()}>
      <Select<TaskDueFilter | null>
        value={value}
        onValueChange={onValueChange}
      >
        <SelectTrigger
          size={2}
          placeholder={t("filters.anyDueDate")}
          aria-label={t("filters.dueDate")}
        >
          {(selected: TaskDueFilter | null) => selected == null
            ? t("filters.anyDueDate")
            : t(`filters.due.${selected}`)}
        </SelectTrigger>
        <SelectPopup align="start">
          <SelectItem value={null}>
            {t("filters.anyDueDate")}
          </SelectItem>
          {taskDueFilters.map(filter => (
            <SelectItem key={filter} value={filter}>
              {t(`filters.due.${filter}`)}
            </SelectItem>
          ))}
        </SelectPopup>
      </Select>
    </div>
  );
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
