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

import { ArrowsClockwiseIcon } from "@phosphor-icons/react";
import { compactDateFormat, formatDuration } from "@probo/i18n";
import { Linear } from "@probo/ui/src/Atoms/ThirdParties/Linear";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type DragEvent, type ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, type To } from "react-router";

import { TASKS_NAMESPACE } from "../i18n";
import { isOverdueTask } from "../isOverdueTask";
import { type TaskPriority, type TaskState } from "../taskState";
import { taskListItem } from "../variants";

import { TaskListItemAssignee as AssigneeControl } from "./TaskListItemAssignee";
import { TaskListItemDeadline } from "./TaskListItemDeadline";
import { TaskListItemEstimate } from "./TaskListItemEstimate";
import { TaskListItemMenu } from "./TaskListItemMenu";
import { TaskListItemPriority } from "./TaskListItemPriority";
import { TaskListItemState } from "./TaskListItemState";

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
  createdAt: string;
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
  detailsTo?: To;
  canUpdateAssignee?: boolean;
  renderAssigneeMenu?: () => ReactNode;
  extraMenu?: ReactNode;
  onDelete?: () => void;
  sectionState?: TaskState;
  canDrag?: boolean;
  isDragging?: boolean;
  isGhost?: boolean;
  ignoreClick?: boolean;
  onStateChange: (state: TaskState) => void;
  onPriorityChange: (priority: TaskPriority) => void;
  onDeadlineChange: (deadline: string | null) => void;
  onTimeEstimateChange: (timeEstimate: string | null) => void;
  onDragStart?: () => void;
  onDragOver?: (event: DragEvent) => void;
  onDrop?: () => void;
  onDragEnd?: () => void;
}

function isInteractiveDragTarget(target: EventTarget | null) {
  return target instanceof Element
    && target.closest("a, button, input, textarea, select") != null;
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
  const slots = taskListItem({
    linked: props.detailsTo != null,
    interaction: rowInteraction(props, isMouseDown),
  });
  const { canDrag } = props;
  const timeEstimate = formatDuration(props.task.timeEstimate, t, {
    style: "short",
    locale: i18n.language,
  });
  const recurrence = formatDuration(props.task.recurrenceInterval, t, {
    style: "short",
    locale: i18n.language,
  });
  const createdAt = compactDateFormat(i18n.language, props.task.createdAt);
  const showAssignee = props.canUpdateAssignee === true || props.task.assignedTo != null;
  const showEstimate = timeEstimate != null;
  const showDeadline = props.task.deadline != null;
  const showMeta = showEstimate
    || showDeadline
    || recurrence != null;

  return (
    <div
      className={slots.root()}
      draggable={canDrag}
      onDragStart={canDrag ? props.onDragStart : undefined}
      onDragOver={canDrag ? props.onDragOver : undefined}
      onDrop={canDrag ? props.onDrop : undefined}
      onDragEnd={canDrag
        ? (event) => {
            event.currentTarget.draggable = true;
            setIsMouseDown(false);
            props.onDragEnd?.();
          }
        : undefined}
      onMouseDown={canDrag
        ? (event) => {
            if (isInteractiveDragTarget(event.target)) {
              event.currentTarget.draggable = false;
              return;
            }
            setIsMouseDown(true);
          }
        : undefined}
      onMouseUp={canDrag
        ? (event) => {
            event.currentTarget.draggable = true;
            setIsMouseDown(false);
          }
        : undefined}
      onMouseLeave={canDrag
        ? (event) => {
            event.currentTarget.draggable = true;
            setIsMouseDown(false);
          }
        : undefined}
    >
      <div className={slots.main()}>
        <TaskListItemPriority
          priority={props.task.priority}
          canUpdate={props.canUpdate}
          isUpdating={props.isUpdating}
          slots={slots}
          onPriorityChange={props.onPriorityChange}
        />
        <TaskListItemState
          displayState={displayState}
          taskState={props.task.state}
          canUpdate={props.canUpdate}
          isUpdating={props.isUpdating}
          slots={slots}
          onStateChange={props.onStateChange}
        />
        <div className={slots.title()}>
          <Text size={2} weight="medium" className={slots.titleText()}>
            {props.detailsTo == null
              ? props.task.name
              : (
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
                )}
          </Text>
        </div>
        {(showMeta || props.task.externalLink != null) && (
          <div className={slots.metaGroup()}>
            {props.task.externalLink != null && (
              <a
                className={`${slots.meta()} hover:underline`}
                href={props.task.externalLink.url}
                target="_blank"
                rel="noreferrer"
                draggable={false}
                onMouseDown={event => event.stopPropagation()}
              >
                <Linear aria-hidden />
                <Text size={1} color="current">
                  {props.task.externalLink.identifier}
                </Text>
              </a>
            )}
            {showEstimate && (
              <TaskListItemEstimate
                timeEstimate={props.task.timeEstimate}
                label={timeEstimate}
                canUpdate={props.canUpdate}
                isUpdating={props.isUpdating}
                slots={slots}
                onTimeEstimateChange={props.onTimeEstimateChange}
              />
            )}
            {showDeadline && (
              <TaskListItemDeadline
                deadline={props.task.deadline}
                isOverdue={isOverdue}
                canUpdate={props.canUpdate}
                isUpdating={props.isUpdating}
                slots={slots}
                onDeadlineChange={props.onDeadlineChange}
              />
            )}
            {recurrence && (
              <span
                className={slots.meta()}
                title={t("recurringBadge.tooltip", { interval: recurrence })}
              >
                <ArrowsClockwiseIcon className={slots.metaIcon()} weight="bold" />
                <Text size={1} color="current">
                  {recurrence}
                </Text>
              </span>
            )}
          </div>
        )}
        {showAssignee && (
          <div className={slots.trailing()}>
            <AssigneeControl
              assignee={props.task.assignedTo}
              canUpdate={props.canUpdateAssignee === true}
              isUpdating={props.isUpdating}
              className={slots.assignee()}
              renderAssigneeMenu={props.renderAssigneeMenu}
            />
          </div>
        )}
        {createdAt !== "" && (
          <Tooltip>
            <TooltipTrigger
              render={(
                <span className={slots.created()} aria-label={t("fields.createdAt")}>
                  <Text size={1} color="current">
                    <time dateTime={props.task.createdAt}>
                      {createdAt}
                    </time>
                  </Text>
                </span>
              )}
            />
            <TooltipPopup>{t("fields.createdAt")}</TooltipPopup>
          </Tooltip>
        )}
        <TaskListItemMenu
          displayState={displayState}
          taskState={props.task.state}
          priority={props.task.priority}
          deadline={props.task.deadline}
          timeEstimate={props.task.timeEstimate}
          canUpdate={props.canUpdate}
          canUpdateAssignee={props.canUpdateAssignee === true}
          isUpdating={props.isUpdating}
          slots={slots}
          renderAssigneeMenu={props.renderAssigneeMenu}
          extraMenu={props.extraMenu}
          onStateChange={props.onStateChange}
          onPriorityChange={props.onPriorityChange}
          onDeadlineChange={props.onDeadlineChange}
          onTimeEstimateChange={props.onTimeEstimateChange}
          onDelete={props.onDelete}
        />
      </div>
    </div>
  );
}
