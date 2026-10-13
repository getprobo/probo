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
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Anchor } from "@probo/ui/src/v2/Link/Anchor";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type DragEvent, type ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import { TASKS_NAMESPACE } from "./i18n";
import { isOverdueTask } from "./isOverdueTask";
import { TaskPriorityIcon } from "./TaskPriorityIcon";
import { TaskStateIcon } from "./TaskStateIcon";
import { type TaskPriority, type TaskState, taskStateKeys, taskStates } from "./taskState";
import { taskListItem } from "./variants";

export interface TaskListItemAssignee {
  id: string;
  fullName: string;
  emailAddress: string;
  avatarUrl?: string | null;
}

export interface TaskListItemTask {
  id: string;
  name: string;
  state: TaskState;
  priority: TaskPriority;
  timeEstimate?: string | null;
  deadline?: string | null;
  recurrenceInterval?: string | null;
  externalLink?: {
    identifier: string;
    url: string;
  } | null;
  assignedTo?: TaskListItemAssignee | null;
}

export interface TaskListItemProps {
  task: TaskListItemTask;
  canUpdate: boolean;
  isUpdating?: boolean;
  detailsTo: string;
  assigneeTo?: string;
  sectionState?: TaskState;
  canDrag?: boolean;
  isDragging?: boolean;
  isGhost?: boolean;
  ignoreClick?: boolean;
  onStateChange: (state: TaskState) => void;
  onDragStart?: () => void;
  onDragOver?: (event: DragEvent) => void;
  onDrop?: () => void;
  onDragEnd?: () => void;
}

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

export function TaskListItem(props: TaskListItemProps) {
  const { t, i18n } = useTranslation(TASKS_NAMESPACE);
  const [isMouseDown, setIsMouseDown] = useState(false);
  const displayState = props.sectionState ?? props.task.state;
  const isOverdue = isOverdueTask(props.task.deadline, displayState);
  const slots = taskListItem({ interaction: rowInteraction(props, isMouseDown) });
  const { canDrag } = props;
  const timeEstimate = formatDuration(props.task.timeEstimate, t);
  const recurrence = formatDuration(props.task.recurrenceInterval, t);

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
        <TaskPriorityIcon priority={props.task.priority} />
        <Text size={2} weight="medium" className={slots.title()}>
          <Link
            to={props.detailsTo}
            className={slots.titleLink()}
            draggable={false}
            onClick={(event) => {
              if (props.ignoreClick) {
                event.preventDefault();
              }
            }}
          >
            {props.task.name}
          </Link>
        </Text>
        {props.task.externalLink && (
          <Anchor
            className={slots.externalLink()}
            size={1}
            color="neutral"
            href={props.task.externalLink.url}
            target="_blank"
            rel="noreferrer"
          >
            {props.task.externalLink.identifier}
          </Anchor>
        )}
        {timeEstimate && (
          <span
            className={slots.meta()}
            title={t("metadata.timeEstimate")}
          >
            <ClockIcon />
            <Text size={1} color="current">
              {timeEstimate}
            </Text>
          </span>
        )}
        {props.task.deadline && (
          <span
            className={slots.meta()}
            title={t("metadata.deadline")}
          >
            <CalendarBlankIcon />
            <Text size={1} color="current">
              <time dateTime={props.task.deadline}>
                {dateFormat(i18n.language, props.task.deadline)}
              </time>
            </Text>
          </span>
        )}
        {recurrence && (
          <span
            className={slots.meta()}
            title={t("recurringBadge.tooltip", { interval: recurrence })}
          >
            <ArrowsClockwiseIcon />
            <Text size={1} color="current">
              {recurrence}
            </Text>
          </span>
        )}
        {(isOverdue || props.task.assignedTo?.fullName) && (
          <div className={slots.trailing()}>
            {isOverdue && (
              <Badge variant="soft" color="red" size={1}>
                {t("metadata.overdue")}
              </Badge>
            )}
            {props.task.assignedTo?.fullName && (
              <TaskListItemAssignee
                assignee={props.task.assignedTo}
                to={props.assigneeTo}
                className={slots.assignee()}
              />
            )}
          </div>
        )}
      </div>
      <div className={slots.state()}>
        <Select
          value={displayState}
          disabled={!props.canUpdate || props.isUpdating}
          onValueChange={(state: TaskState | null) => {
            if (state != null && state !== props.task.state) {
              props.onStateChange(state);
            }
          }}
        >
          <SelectTrigger size={1} variant="ghost">
            {(state: TaskState | null) => state
              ? (
                  <span className={slots.stateOption()}>
                    <TaskStateIcon state={state} />
                    {t(`states.${taskStateKeys[state]}`)}
                  </span>
                )
              : null}
          </SelectTrigger>
          <SelectPopup align="end">
            {taskStates.map(state => (
              <SelectItem key={state} value={state}>
                <span className={slots.stateOption()}>
                  <TaskStateIcon state={state} />
                  {t(`states.${taskStateKeys[state]}`)}
                </span>
              </SelectItem>
            ))}
          </SelectPopup>
        </Select>
      </div>
    </div>
  );
}

function TaskListItemAssignee({
  assignee,
  to,
  className,
}: {
  assignee: TaskListItemAssignee;
  to?: string;
  className: string;
}): ReactNode {
  const avatar = (
    <Avatar
      size={1}
      radius="full"
      name={assignee.fullName}
      email={assignee.emailAddress}
      src={assignee.avatarUrl}
    />
  );

  return (
    <Tooltip>
      <TooltipTrigger
        render={
          to === undefined
            ? (
                <span className={className} aria-label={assignee.fullName}>
                  {avatar}
                </span>
              )
            : (
                <Link
                  className={className}
                  to={to}
                  aria-label={assignee.fullName}
                >
                  {avatar}
                </Link>
              )
        }
      />
      <TooltipPopup>{assignee.fullName}</TooltipPopup>
    </Tooltip>
  );
}
