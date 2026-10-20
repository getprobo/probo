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

import {
  PropertyRow,
  TaskDurationField,
  type TaskPriority,
  TaskProperties,
  taskRecurrenceDurationUnits,
  TASKS_NAMESPACE,
  type TaskState,
} from "@probo/tasks";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Link/Link";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { TaskPropertiesSection_task$key } from "#/__generated__/core/TaskPropertiesSection_task.graphql";
import type { TaskPropertiesSectionInternalControlsQuery } from "#/__generated__/core/TaskPropertiesSectionInternalControlsQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { useUpdateTask } from "../_lib/useUpdateTask";
import { taskPropertiesSection } from "../variants";

import { TaskAssigneeField } from "./TaskAssigneeField";
import { TaskInternalControlField } from "./TaskInternalControlField";
import { TaskLinearField } from "./TaskLinearField";

const taskInternalControlsPageSize = 50;

const taskPropertiesSectionFragment = graphql`
  fragment TaskPropertiesSection_task on Task
    @refetchable(queryName: "TaskPropertiesSectionInternalControlsQuery")
    @argumentDefinitions(
      first: { type: "Int", defaultValue: 50 }
      after: { type: "CursorKey" }
    ) {
    id
    state
    priority
    timeEstimate
    deadline
    recurrenceInterval
    createdAt
    updatedAt
    canUpdate: permission(action: "core:task:update")
    assignedTo {
      id
      fullName
    }
    internalControls(first: $first, after: $after)
      @connection(key: "TaskPropertiesSection_internalControls") {
      edges {
        node {
          id
          name
        }
      }
    }
    ...TaskAssigneeField_task
    ...TaskLinearField_task
  }
`;

interface TaskPropertiesSectionProps {
  taskKey: TaskPropertiesSection_task$key;
}

export function TaskPropertiesSection({ taskKey }: TaskPropertiesSectionProps) {
  const { t } = useTranslation("organizations/tasks");
  const { t: tTasks } = useTranslation(TASKS_NAMESPACE);
  const organizationId = useOrganizationId();
  const {
    data: task,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    TaskPropertiesSectionInternalControlsQuery,
    TaskPropertiesSection_task$key
  >(taskPropertiesSectionFragment, taskKey);
  const [updateTask, isUpdating] = useUpdateTask();
  const { names } = taskPropertiesSection();
  const internalControls = task.internalControls.edges.map(edge => edge.node);

  function save(
    input: {
      state?: TaskState;
      priority?: TaskPriority;
      assignedToId?: string | null;
      internalControlIds?: string[];
      timeEstimate?: string | null;
      deadline?: string | null;
      recurrenceInterval?: string | null;
    },
  ) {
    const result = updateTask({
      previousInternalControlIds: internalControls.map(internalControl => internalControl.id),
      variables: {
        input: {
          taskId: task.id,
          ...input,
        },
      },
    });
    void result.catch(() => {
      // Error toast is already shown by useMutation.
    });
    return result;
  }

  return (
    <TaskProperties
      task={task}
      canUpdate={task.canUpdate}
      isUpdating={isUpdating}
      extraProperties={(
        <>
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
                      to={`/organizations/${organizationId}/settings/users/${task.assignedTo.id}`}
                    >
                      {task.assignedTo.fullName}
                    </Link>
                  )
                : (
                    <Text size={2} color="faint">{t("detailsPage.unassigned")}</Text>
                  )}
          </PropertyRow>
          <Suspense fallback={null}>
            <TaskLinearField taskKey={task} />
          </Suspense>
          <PropertyRow
            label={t("detailsPage.fields.internalControls")}
            align={internalControls.length > 0 || hasNext ? "start" : "center"}
          >
            <span className={names()}>
              {task.canUpdate
                ? (
                    <Suspense fallback={<SelectSkeleton size={1} className="w-full" />}>
                      <TaskInternalControlField
                        internalControls={internalControls}
                        disabled={isUpdating || hasNext || isLoadingNext}
                        onValueChange={(internalControlIds) => {
                          void save({ internalControlIds });
                        }}
                      />
                    </Suspense>
                  )
                : internalControls.length > 0
                  ? (
                      internalControls.map(internalControl => (
                        <Link
                          key={internalControl.id}
                          size={2}
                          to={`/organizations/${organizationId}/governance/internal-controls/${internalControl.id}`}
                        >
                          {internalControl.name}
                        </Link>
                      ))
                    )
                  : (
                      <Text size={2} color="faint">{t("detailsPage.none")}</Text>
                    )}
              {hasNext && (
                <Button
                  variant="ghost"
                  color="neutral"
                  size={1}
                  className="self-start"
                  loading={isLoadingNext}
                  onClick={() => loadNext(taskInternalControlsPageSize)}
                >
                  {t("detailsPage.actions.loadMoreInternalControls")}
                </Button>
              )}
            </span>
          </PropertyRow>
        </>
      )}
      timeEstimateField={task.canUpdate
        ? (
            <TaskDurationField
              value={task.timeEstimate ?? null}
              onValueChange={timeEstimate => save({ timeEstimate })}
            />
          )
        : undefined}
      recurrenceField={task.canUpdate
        ? (
            <TaskDurationField
              value={task.recurrenceInterval ?? null}
              disabled={!task.deadline && !task.recurrenceInterval}
              defaultValue="P7D"
              units={taskRecurrenceDurationUnits}
              addLabel={tTasks("actions.addRecurrence")}
              addTitle={!task.deadline
                ? tTasks("hints.recurrenceNeedsDeadline")
                : undefined}
              clearLabel={tTasks("actions.clearRecurrence")}
              amountAriaLabel={tTasks("fields.recurrence")}
              unitAriaLabel={tTasks("fields.recurrenceUnit")}
              onValueChange={recurrenceInterval => save({ recurrenceInterval })}
            />
          )
        : undefined}
      onChange={(change) => {
        void save(change);
      }}
    />
  );
}
