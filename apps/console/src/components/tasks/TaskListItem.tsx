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
  ArrowsClockwiseIcon,
  CalendarBlankIcon,
  ClockIcon,
} from "@phosphor-icons/react";
import { dateFormat, formatDuration } from "@probo/i18n";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Anchor } from "@probo/ui/src/v2/Link/Anchor";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useRelayEnvironment } from "react-relay";
import { Link } from "react-router";

import type { TaskListItem_task$key } from "#/__generated__/core/TaskListItem_task.graphql";
import type { TaskListItemUpdateStateMutation } from "#/__generated__/core/TaskListItemUpdateStateMutation.graphql";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";
import { TaskPriorityIcon } from "#/pages/organizations/tasks/_components/TaskPriorityIcon";
import { TaskStateIcon } from "#/pages/organizations/tasks/_components/TaskStateIcon";
import { insertNextTaskEdge } from "#/pages/organizations/tasks/_lib/taskConnectionOrder";
import { taskDetailsPath } from "#/pages/organizations/tasks/_lib/taskPath";
import {
  type TaskState,
  taskStateKeys,
  taskStates,
} from "#/pages/organizations/tasks/_lib/taskState";

import { taskListItem } from "./variants";

const taskListItemFragment = graphql`
  fragment TaskListItem_task on Task {
    id
    name
    state
    priority
    timeEstimate
    deadline
    recurrenceInterval
    canUpdate: permission(action: "core:task:update")
    externalLink {
      identifier
      url
    }
    assignedTo {
      id
      fullName
      emailAddress
      avatar {
        downloadUrl
      }
    }
  }
`;

const updateStateMutation = graphql`
  mutation TaskListItemUpdateStateMutation($input: UpdateTaskInput!) {
    updateTask(input: $input) {
      task {
        ...TasksCard_task
        ...TaskListItem_task
        ...TaskDetailsPage_task
      }
      nextTaskEdge {
        node {
          ...TasksCard_task
          ...TaskListItem_task
          measure {
            id
          }
        }
      }
    }
  }
`;

function rowInteraction(
  props: TaskListItemProps,
  isMouseDown: boolean,
): "idle" | "grab" | "grabbing" | "dragging" | "ghost" {
  if (props.isGhost) {
    return "ghost";
  }
  if (!props.canDrag) {
    return "idle";
  }
  if (props.isDragging) {
    return "dragging";
  }

  return isMouseDown ? "grabbing" : "grab";
}

interface TaskListItemProps {
  taskKey: TaskListItem_task$key;
  connectionId: string;
  sectionState?: TaskState;
  canDrag?: boolean;
  isDragging?: boolean;
  isGhost?: boolean;
  ignoreClick: boolean;
  onDragStart?: () => void;
  onDragOver?: (event: React.DragEvent) => void;
  onDrop?: () => void;
  onDragEnd?: () => void;
  onStateChange?: () => void;
}

export function TaskListItem(props: TaskListItemProps) {
  const organizationId = useOrganizationId();
  const { t, i18n } = useTranslation();
  const relayEnv = useRelayEnvironment();
  const { canUpdate, ...task } = useFragment(taskListItemFragment, props.taskKey);
  const [updateState, isUpdating] = useMutation<TaskListItemUpdateStateMutation>(
    updateStateMutation,
  );
  const [isMouseDown, setIsMouseDown] = useState(false);
  const displayState = props.sectionState ?? task.state;

  function onStateChange(state: TaskState) {
    void updateState({
      variables: {
        input: {
          taskId: task.id,
          state,
        },
      },
      updater: (store) => {
        const spawnedMeasureId = insertNextTaskEdge(
          store,
          organizationId,
          [props.connectionId],
        );
        if (spawnedMeasureId) {
          updateStoreCounter(relayEnv, spawnedMeasureId, "tasks(first:0)", 1);
        }
      },
      onCompleted: () => {
        props.onStateChange?.();
      },
    }).catch(() => {
      // Error feedback is handled by useMutation.
    });
  }

  const { canDrag } = props;
  const slots = taskListItem({ interaction: rowInteraction(props, isMouseDown) });
  const detailsUrl = taskDetailsPath(organizationId, task.id);

  return (
    <div
      className={slots.root()}
      draggable={canDrag}
      onDragStart={canDrag ? props.onDragStart : undefined}
      onDragOver={canDrag ? props.onDragOver : undefined}
      onDrop={canDrag ? props.onDrop : undefined}
      onDragEnd={canDrag ? props.onDragEnd : undefined}
      onMouseDown={canDrag ? () => setIsMouseDown(true) : undefined}
      onMouseUp={canDrag ? () => setIsMouseDown(false) : undefined}
      onMouseLeave={canDrag ? () => setIsMouseDown(false) : undefined}
    >
      <div className={slots.main()}>
        <TaskPriorityIcon priority={task.priority} />
        <Text size={2} weight="medium" className={slots.title()}>
          <Link
            to={detailsUrl}
            className={slots.titleLink()}
            draggable={false}
            onClick={(event) => {
              if (props.ignoreClick) {
                event.preventDefault();
              }
            }}
          >
            {task.name}
          </Link>
        </Text>
        {task.externalLink && (
          <Anchor
            className={slots.externalLink()}
            size={1}
            color="neutral"
            href={task.externalLink.url}
            target="_blank"
            rel="noreferrer"
          >
            {task.externalLink.identifier}
          </Anchor>
        )}
        {task.timeEstimate && (
          <span
            className={slots.meta()}
            title={t("tasksCard.metadata.timeEstimate")}
          >
            <ClockIcon />
            <Text size={1} color="current">
              {formatDuration(task.timeEstimate, t)}
            </Text>
          </span>
        )}
        {task.deadline && (
          <span
            className={slots.meta()}
            title={t("tasksCard.metadata.deadline")}
          >
            <CalendarBlankIcon />
            <Text size={1} color="current">
              <time dateTime={task.deadline}>
                {dateFormat(i18n.language, task.deadline)}
              </time>
            </Text>
          </span>
        )}
        {task.recurrenceInterval && (
          <span
            className={slots.meta()}
            title={t("tasksCard.recurringBadge.tooltip", {
              interval: formatDuration(task.recurrenceInterval, t),
            })}
          >
            <ArrowsClockwiseIcon />
            <Text size={1} color="current">
              {formatDuration(task.recurrenceInterval, t)}
            </Text>
          </span>
        )}
        {task.assignedTo?.fullName && (
          <Tooltip>
            <TooltipTrigger
              render={(
                <Link
                  className={slots.assignee()}
                  to={`/organizations/${organizationId}/settings/users/${task.assignedTo.id}`}
                  aria-label={task.assignedTo.fullName}
                >
                  <Avatar
                    size={1}
                    radius="full"
                    name={task.assignedTo.fullName}
                    email={task.assignedTo.emailAddress}
                    src={task.assignedTo.avatar?.downloadUrl}
                  />
                </Link>
              )}
            />
            <TooltipPopup>{task.assignedTo.fullName}</TooltipPopup>
          </Tooltip>
        )}
      </div>
      <div className={slots.state()}>
        <Select
          value={displayState}
          disabled={!canUpdate || isUpdating}
          onValueChange={(state: TaskState | null) => {
            if (state != null && state !== task.state) {
              onStateChange(state);
            }
          }}
        >
          <SelectTrigger size={1} variant="ghost">
            {(state: TaskState | null) => state
              ? (
                  <span className={slots.stateOption()}>
                    <TaskStateIcon state={state} />
                    {t(`tasksCard.states.${taskStateKeys[state]}`)}
                  </span>
                )
              : null}
          </SelectTrigger>
          <SelectPopup align="end">
            {taskStates.map(state => (
              <SelectItem key={state} value={state}>
                <span className={slots.stateOption()}>
                  <TaskStateIcon state={state} />
                  {t(`tasksCard.states.${taskStateKeys[state]}`)}
                </span>
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
    </div>
  );
}
