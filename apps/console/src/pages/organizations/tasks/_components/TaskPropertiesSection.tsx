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
import { PriorityLevel, TaskStateIcon } from "@probo/ui";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { TaskPropertiesSection_task$key } from "#/__generated__/core/TaskPropertiesSection_task.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import type { TaskPriority, TaskState } from "../_lib/taskState";
import {
  taskPriorities,
  taskStateKeys,
  taskStates,
} from "../_lib/taskState";
import { useUpdateTask } from "../_lib/useUpdateTask";
import { taskPropertiesSection } from "../variants";

import { TaskAssigneeField } from "./TaskAssigneeField";
import { TaskDurationField } from "./TaskDurationField";
import { TaskMeasureField } from "./TaskMeasureField";

const taskPropertiesSectionFragment = graphql`
  fragment TaskPropertiesSection_task on Task {
    id
    state
    priority
    timeEstimate
    deadline
    createdAt
    updatedAt
    canUpdate: permission(action: "core:task:update")
    assignedTo {
      id
      fullName
    }
    measure {
      id
      name
    }
    ...TaskAssigneeField_task
    ...TaskMeasureField_task
  }
`;

interface TaskPropertiesSectionProps {
  taskKey: TaskPropertiesSection_task$key;
}

export function TaskPropertiesSection({ taskKey }: TaskPropertiesSectionProps) {
  const { t, i18n } = useTranslation("organizations/tasks");
  const { t: tApp } = useTranslation();
  const organizationId = useOrganizationId();
  const task = useFragment(taskPropertiesSectionFragment, taskKey);
  const [updateTask, isUpdating] = useUpdateTask();
  const { root, value } = taskPropertiesSection();
  const empty = t("detailsPage.empty");

  function save(
    input: {
      state?: TaskState;
      priority?: TaskPriority;
      assignedToId?: string | null;
      measureId?: string | null;
      timeEstimate?: string | null;
      deadline?: string | null;
    },
  ) {
    return updateTask({
      variables: {
        input: {
          taskId: task.id,
          ...input,
        },
      },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <Card size={2}>
      <div className={root()}>
        <PropertyRow label={t("detailsPage.fields.state")}>
          {task.canUpdate
            ? (
                <Select
                  value={task.state}
                  disabled={isUpdating}
                  onValueChange={(state: TaskState | null) => {
                    if (state == null || state === task.state) {
                      return;
                    }
                    void save({ state });
                  }}
                >
                  <SelectTrigger size={1} aria-label={t("detailsPage.fields.state")}>
                    {(state: TaskState | null) =>
                      state
                        ? (
                            <span className={value()}>
                              <TaskStateIcon state={state} />
                              {t(`detailsPage.states.${taskStateKeys[state]}`)}
                            </span>
                          )
                        : null}
                  </SelectTrigger>
                  <SelectPopup>
                    {taskStates.map(state => (
                      <SelectItem key={state} value={state}>
                        <span className={value()}>
                          <TaskStateIcon state={state} />
                          {t(`detailsPage.states.${taskStateKeys[state]}`)}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectPopup>
                </Select>
              )
            : (
                <span className={value()}>
                  <TaskStateIcon state={task.state} />
                  <Text size={2}>{t(`detailsPage.states.${taskStateKeys[task.state]}`)}</Text>
                </span>
              )}
        </PropertyRow>
        <PropertyRow label={t("detailsPage.fields.priority")}>
          {task.canUpdate
            ? (
                <Select
                  value={task.priority}
                  disabled={isUpdating}
                  onValueChange={(priority: TaskPriority | null) => {
                    if (priority == null || priority === task.priority) {
                      return;
                    }
                    void save({ priority });
                  }}
                >
                  <SelectTrigger size={1} aria-label={t("detailsPage.fields.priority")}>
                    {(priority: TaskPriority | null) =>
                      priority
                        ? (
                            <span className={value()}>
                              <PriorityLevel level={priority} />
                              {t(`detailsPage.priorities.${priority.toLowerCase()}`)}
                            </span>
                          )
                        : null}
                  </SelectTrigger>
                  <SelectPopup>
                    {taskPriorities.map(priority => (
                      <SelectItem key={priority} value={priority}>
                        <span className={value()}>
                          <PriorityLevel level={priority} />
                          {t(`detailsPage.priorities.${priority.toLowerCase()}`)}
                        </span>
                      </SelectItem>
                    ))}
                  </SelectPopup>
                </Select>
              )
            : (
                <span className={value()}>
                  <PriorityLevel level={task.priority} />
                  <Text size={2}>
                    {t(`detailsPage.priorities.${task.priority.toLowerCase()}`)}
                  </Text>
                </span>
              )}
        </PropertyRow>
        <PropertyRow label={t("detailsPage.fields.assignedTo")}>
          {task.canUpdate
            ? (
                <Suspense fallback={<SelectSkeleton size={1} className="w-full" />}>
                  <TaskAssigneeField
                    taskKey={task}
                    disabled={isUpdating}
                    onValueChange={(assignedToId) => {
                      void save({ assignedToId });
                    }}
                  />
                </Suspense>
              )
            : task.assignedTo
              ? (
                  <Link
                    size={2}
                    to={`/organizations/${organizationId}/settings/people/${task.assignedTo.id}`}
                  >
                    {task.assignedTo.fullName}
                  </Link>
                )
              : (
                  <Text size={2} color="faint">{t("detailsPage.unassigned")}</Text>
                )}
        </PropertyRow>
        <PropertyRow label={t("detailsPage.fields.measure")}>
          {task.canUpdate
            ? (
                <Suspense fallback={<SelectSkeleton size={1} className="w-full" />}>
                  <TaskMeasureField
                    taskKey={task}
                    disabled={isUpdating}
                    onValueChange={(measureId) => {
                      void save({ measureId });
                    }}
                  />
                </Suspense>
              )
            : task.measure
              ? (
                  <Link
                    size={2}
                    to={`/organizations/${organizationId}/governance/measures/${task.measure.id}`}
                  >
                    {task.measure.name}
                  </Link>
                )
              : (
                  <Text size={2} color="faint">{empty}</Text>
                )}
        </PropertyRow>
        <PropertyRow label={t("detailsPage.fields.timeEstimate")}>
          {task.canUpdate
            ? (
                <TaskDurationField
                  value={task.timeEstimate ?? null}
                  disabled={isUpdating}
                  onValueChange={timeEstimate => save({ timeEstimate })}
                />
              )
            : (
                <Text size={2}>{formatDuration(task.timeEstimate, tApp) ?? empty}</Text>
              )}
        </PropertyRow>
        <PropertyRow label={t("detailsPage.fields.deadline")}>
          {task.canUpdate
            ? (
                <TextField
                  size={1}
                  type="date"
                  value={toDateInput(task.deadline)}
                  disabled={isUpdating}
                  aria-label={t("detailsPage.fields.deadline")}
                  onChange={(event) => {
                    const next = event.currentTarget.value;
                    const deadline = next ? formatDatetime(next) ?? null : null;
                    const current = task.deadline ? toDateInput(task.deadline) : "";
                    if (next === current) {
                      return;
                    }
                    void save({ deadline });
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
        <PropertyRow label={t("detailsPage.fields.createdAt")}>
          <Text size={2}>
            <time dateTime={task.createdAt}>
              {dateTimeFormat(i18n.language, task.createdAt)}
            </time>
          </Text>
        </PropertyRow>
        <PropertyRow label={t("detailsPage.fields.updatedAt")}>
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

function PropertyRow({ label, children }: { label: string; children: ReactNode }) {
  const { row } = taskPropertiesSection();

  return (
    <div className={row()}>
      <Text size={2} color="faint">{label}</Text>
      {children}
    </div>
  );
}
