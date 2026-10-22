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

import { ArrowsClockwiseIcon, ClockIcon } from "@phosphor-icons/react";
import { formatDuration } from "@probo/i18n";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "./i18n";
import { isOverdueTask } from "./isOverdueTask";
import { taskEstimateDurationUnits, taskRecurrenceDurationUnits } from "./taskDuration";
import { type TaskPriority, type TaskState } from "./taskState";
import { TaskDurationField } from "./TaskDurationField";
import { TaskListItemDeadline } from "./TaskListItem/TaskListItemDeadline";
import { TaskListItemPriority } from "./TaskListItem/TaskListItemPriority";
import { TaskListItemState } from "./TaskListItem/TaskListItemState";
import { taskListItem, taskProperties } from "./variants";

export interface TaskPropertiesTask {
  state: TaskState;
  priority: TaskPriority;
  timeEstimate?: string | null;
  deadline?: string | null;
  recurrenceInterval?: string | null;
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
  assignee?: ReactNode;
  onChange: (change: TaskPropertiesChange) => void | Promise<unknown>;
}

export function TaskProperties({
  task,
  canUpdate,
  isUpdating = false,
  assignee,
  onChange,
}: TaskPropertiesProps) {
  const { t, i18n } = useTranslation(TASKS_NAMESPACE);
  const slots = taskListItem({ field: true });
  const { root, recurrence, recurrenceIcon, recurrenceField } = taskProperties();
  const timeEstimate = formatDuration(task.timeEstimate, t, {
    style: "short",
    locale: i18n.language,
  });
  const recurrenceLabel = formatDuration(task.recurrenceInterval, t, {
    style: "short",
    locale: i18n.language,
  });

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <Heading level={2} size={2} weight="medium">
          {t("properties.title")}
        </Heading>
        <TaskListItemState
          displayState={task.state}
          taskState={task.state}
          canUpdate={canUpdate}
          isUpdating={isUpdating}
          showLabel
          placeholder={t("placeholders.state")}
          slots={slots}
          onStateChange={(state) => {
            onChange({ state });
          }}
        />
        <TaskListItemPriority
          priority={task.priority}
          canUpdate={canUpdate}
          isUpdating={isUpdating}
          showLabel
          placeholder={t("placeholders.priority")}
          slots={slots}
          onPriorityChange={(priority) => {
            onChange({ priority });
          }}
        />
        {assignee}
        <div className={canUpdate ? `${recurrence()} ${slots.chipTrigger()}` : recurrence()}>
          <ClockIcon className={slots.metaIcon()} weight="bold" />
          {canUpdate
            ? (
                <div className={recurrenceField()}>
                  <TaskDurationField
                    value={task.timeEstimate ?? null}
                    disabled={isUpdating}
                    units={taskEstimateDurationUnits}
                    clearLabel={t("actions.clearEstimate")}
                    amountAriaLabel={t("fields.timeEstimate")}
                    unitAriaLabel={t("fields.timeEstimateUnit")}
                    onValueChange={next => onChange({ timeEstimate: next })}
                  />
                </div>
              )
            : (
                <Text size={2} color={timeEstimate == null ? "faint" : undefined}>
                  {timeEstimate ?? t("placeholders.timeEstimate")}
                </Text>
              )}
        </div>
        <TaskListItemDeadline
          deadline={task.deadline}
          isOverdue={isOverdueTask(task.deadline, task.state)}
          canUpdate={canUpdate}
          isUpdating={isUpdating}
          size={2}
          placeholder={t("placeholders.deadline")}
          slots={slots}
          onDeadlineChange={(deadline) => {
            onChange({
              deadline,
              ...(deadline == null && task.recurrenceInterval != null
                ? { recurrenceInterval: null }
                : {}),
            });
          }}
        />
        {canUpdate && recurrenceLabel == null
          ? (
              <Tooltip>
                <TooltipTrigger
                  render={(
                    <button
                      type="button"
                      className={slots.chipTrigger()}
                      disabled={isUpdating || !task.deadline}
                      onClick={() => {
                        onChange({ recurrenceInterval: "P7D" });
                      }}
                    >
                      <span className={slots.stateOption()}>
                        <ArrowsClockwiseIcon className={recurrenceIcon()} weight="bold" />
                        <Text size={2} color="faint">{t("placeholders.recurrence")}</Text>
                      </span>
                    </button>
                  )}
                />
                <TooltipPopup>
                  {task.deadline
                    ? t("placeholders.recurrence")
                    : t("hints.recurrenceNeedsDeadline")}
                </TooltipPopup>
              </Tooltip>
            )
          : (
              <div className={canUpdate ? `${recurrence()} ${slots.chipTrigger()}` : recurrence()}>
                <ArrowsClockwiseIcon className={recurrenceIcon()} weight="bold" />
                {canUpdate
                  ? (
                      <div className={recurrenceField()}>
                        <TaskDurationField
                          value={task.recurrenceInterval ?? null}
                          disabled={isUpdating}
                          defaultValue="P7D"
                          units={taskRecurrenceDurationUnits}
                          clearLabel={t("actions.clearRecurrence")}
                          amountAriaLabel={t("fields.recurrence")}
                          unitAriaLabel={t("fields.recurrenceUnit")}
                          onValueChange={recurrenceInterval => onChange({ recurrenceInterval })}
                        />
                      </div>
                    )
                  : (
                      <Text size={2} color={recurrenceLabel == null ? "faint" : undefined}>
                        {recurrenceLabel ?? t("placeholders.recurrence")}
                      </Text>
                    )}
              </div>
            )}
      </div>
    </Card>
  );
}
