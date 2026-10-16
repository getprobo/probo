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

import { formatDatetime, toDateInput } from "@probo/helpers";
import { dateFormat, dateTimeFormat, formatDuration } from "@probo/i18n";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { cloneElement, isValidElement, type ReactElement, type ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "./i18n";
import { TaskPriorityIcon } from "./TaskPriorityIcon";
import {
  taskPriorities,
  type TaskPriority,
  taskPriorityKeys,
  type TaskState,
  taskStateKeys,
  taskStates,
} from "./taskState";
import { TaskStateIcon } from "./TaskStateIcon";
import { taskProperties } from "./variants";

export interface TaskPropertiesTask {
  state: TaskState;
  priority: TaskPriority;
  timeEstimate?: string | null;
  deadline?: string | null;
  recurrenceInterval?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface TaskPropertiesChange {
  state?: TaskState;
  priority?: TaskPriority;
  timeEstimate?: string | null;
  deadline?: string | null;
  recurrenceInterval?: string | null;
}

export interface TaskPropertiesProps {
  task: TaskPropertiesTask;
  canUpdate: boolean;
  isUpdating?: boolean;
  extraProperties?: ReactNode;
  timeEstimateField?: ReactNode;
  recurrenceField?: ReactNode;
  onChange: (change: TaskPropertiesChange) => void;
}

export function TaskProperties({
  task,
  canUpdate,
  isUpdating = false,
  extraProperties,
  timeEstimateField,
  recurrenceField,
  onChange,
}: TaskPropertiesProps) {
  const { t, i18n } = useTranslation(TASKS_NAMESPACE);
  const { root, value } = taskProperties();
  const empty = t("emptyValue");

  return (
    <Card size={2}>
      <div className={root()}>
        <PropertyRow label={t("fields.state")}>
          {canUpdate
            ? (
                <Select
                  value={task.state}
                  disabled={isUpdating}
                  onValueChange={(state: TaskState | null) => {
                    if (state == null || state === task.state) {
                      return;
                    }
                    onChange({ state });
                  }}
                >
                  <SelectTrigger size={1} aria-label={t("fields.state")}>
                    {(state: TaskState | null) =>
                      state
                        ? (
                            <span className={value()}>
                              <TaskStateIcon state={state} />
                              {t(`states.${taskStateKeys[state]}`)}
                            </span>
                          )
                        : null}
                  </SelectTrigger>
                  <SelectPopup>
                    {taskStates.map(state => (
                      <SelectItem key={state} value={state}>
                        <span className={value()}>
                          <TaskStateIcon state={state} />
                          {t(`states.${taskStateKeys[state]}`)}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectPopup>
                </Select>
              )
            : (
                <span className={value()}>
                  <TaskStateIcon state={task.state} />
                  <Text size={2}>{t(`states.${taskStateKeys[task.state]}`)}</Text>
                </span>
              )}
        </PropertyRow>
        <PropertyRow label={t("fields.priority")}>
          {canUpdate
            ? (
                <Select
                  value={task.priority}
                  disabled={isUpdating}
                  onValueChange={(priority: TaskPriority | null) => {
                    if (priority == null || priority === task.priority) {
                      return;
                    }
                    onChange({ priority });
                  }}
                >
                  <SelectTrigger size={1} aria-label={t("fields.priority")}>
                    {(priority: TaskPriority | null) =>
                      priority
                        ? (
                            <span className={value()}>
                              <TaskPriorityIcon priority={priority} />
                              {t(`priorities.${taskPriorityKeys[priority]}`)}
                            </span>
                          )
                        : null}
                  </SelectTrigger>
                  <SelectPopup>
                    {taskPriorities.map(priority => (
                      <SelectItem key={priority} value={priority}>
                        <span className={value()}>
                          <TaskPriorityIcon priority={priority} />
                          {t(`priorities.${taskPriorityKeys[priority]}`)}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectPopup>
                </Select>
              )
            : (
                <span className={value()}>
                  <TaskPriorityIcon priority={task.priority} />
                  <Text size={2}>
                    {t(`priorities.${taskPriorityKeys[task.priority]}`)}
                  </Text>
                </span>
              )}
        </PropertyRow>
        {extraProperties}
        <PropertyRow label={t("fields.timeEstimate")}>
          {timeEstimateField != null
            ? withDisabled(timeEstimateField, isUpdating)
            : (
                <Text size={2}>{formatDuration(task.timeEstimate, t) ?? empty}</Text>
              )}
        </PropertyRow>
        <PropertyRow label={t("fields.deadline")}>
          {canUpdate
            ? (
                <TextField
                  size={1}
                  type="date"
                  value={toDateInput(task.deadline)}
                  disabled={isUpdating}
                  aria-label={t("fields.deadline")}
                  onChange={(event) => {
                    const next = event.currentTarget.value;
                    const deadline = next ? formatDatetime(next) ?? null : null;
                    const current = task.deadline ? toDateInput(task.deadline) : "";
                    if (next === current) {
                      return;
                    }
                    if (!deadline && task.recurrenceInterval) {
                      onChange({ deadline, recurrenceInterval: null });
                      return;
                    }
                    onChange({ deadline });
                  }}
                />
              )
            : task.deadline
              ? (
                  <Text size={2}>
                    <time dateTime={toDateInput(task.deadline)}>
                      {dateFormat(i18n.language, toDateInput(task.deadline))}
                    </time>
                  </Text>
                )
              : (
                  <Text size={2} color="faint">{empty}</Text>
                )}
        </PropertyRow>
        <PropertyRow label={t("fields.recurrence")}>
          {recurrenceField != null
            ? withDisabled(recurrenceField, isUpdating)
            : (
                task.recurrenceInterval
                  ? (
                      <Text size={2}>{formatDuration(task.recurrenceInterval, t)}</Text>
                    )
                  : (
                      <Text size={2} color="faint">{t("doesNotRepeat")}</Text>
                    )
              )}
        </PropertyRow>
        <PropertyRow label={t("fields.createdAt")}>
          <Text size={2}>
            <time dateTime={task.createdAt}>
              {dateTimeFormat(i18n.language, task.createdAt)}
            </time>
          </Text>
        </PropertyRow>
        <PropertyRow label={t("fields.updatedAt")}>
          <Text size={2}>
            <time dateTime={task.updatedAt}>
              {dateTimeFormat(i18n.language, task.updatedAt)}
            </time>
          </Text>
        </PropertyRow>
      </div>
    </Card>
  );
}

function withDisabled(field: ReactNode, disabled: boolean): ReactNode {
  if (!disabled || !isValidElement<{ disabled?: boolean }>(field)) {
    return field;
  }

  const element = field as ReactElement<{ disabled?: boolean }>;
  return cloneElement(element, {
    disabled: element.props.disabled || disabled,
  });
}

export function PropertyRow({
  label,
  children,
  align = "center",
}: {
  label: string;
  children: ReactNode;
  align?: "center" | "start";
}) {
  const { row, label: labelClassName } = taskProperties({ align });

  return (
    <div className={row()}>
      <Text size={2} color="faint" className={labelClassName()}>{label}</Text>
      {children}
    </div>
  );
}
