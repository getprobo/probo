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

import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "../i18n";
import { type TaskDueFilter, taskDueFilters } from "../taskDueFilter";
import { taskListToolbar } from "../variants";

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
