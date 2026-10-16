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
  CalendarXIcon,
  ClockIcon,
  DotsThreeVerticalIcon,
  ProhibitIcon,
  TrashIcon,
  UserCircleDashedIcon,
} from "@phosphor-icons/react";
import { formatDatetime, parseDate, toDateInput, todayAsDateInput } from "@probo/helpers";
import { compactDateFormat, deadlineDateFormat, formatDuration } from "@probo/i18n";
import { Linear } from "@probo/ui/src/Atoms/ThirdParties/Linear";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownRadioGroup } from "@probo/ui/src/v2/Dropdown/DropdownRadioGroup";
import { DropdownRadioItem } from "@probo/ui/src/v2/Dropdown/DropdownRadioItem";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { DropdownSubmenu } from "@probo/ui/src/v2/Dropdown/DropdownSubmenu";
import { DropdownSubmenuTrigger } from "@probo/ui/src/v2/Dropdown/DropdownSubmenuTrigger";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { DateCalendar } from "@probo/ui/src/v2/form/DateCalendar";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type DragEvent, type ReactNode, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

import { TASKS_NAMESPACE } from "./i18n";
import { isOverdueTask } from "./isOverdueTask";
import { stringifyTaskDuration, taskDurationsEqual } from "./taskDuration";
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
import { taskListItem } from "./variants";

const noneEstimateValue = "__NONE__";

const estimatePresets = [
  stringifyTaskDuration(15, "M"),
  stringifyTaskDuration(30, "M"),
  stringifyTaskDuration(1, "H"),
  stringifyTaskDuration(2, "H"),
  stringifyTaskDuration(4, "H"),
  stringifyTaskDuration(1, "D"),
  stringifyTaskDuration(1, "W"),
].filter((value): value is string => value != null);

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
  detailsTo?: string;
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
      onDragEnd={canDrag ? props.onDragEnd : undefined}
      onMouseDown={canDrag ? () => setIsMouseDown(true) : undefined}
      onMouseUp={canDrag ? () => setIsMouseDown(false) : undefined}
      onMouseLeave={canDrag ? () => setIsMouseDown(false) : undefined}
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
            <TaskListItemAssignee
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

function TaskListItemPriority({
  priority,
  canUpdate,
  isUpdating,
  slots,
  onPriorityChange,
}: {
  priority: TaskPriority;
  canUpdate: boolean;
  isUpdating?: boolean;
  slots: ReturnType<typeof taskListItem>;
  onPriorityChange: (priority: TaskPriority) => void;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const [open, setOpen] = useState(false);
  const label = t(`priorities.${taskPriorityKeys[priority]}`);
  const icon = <TaskPriorityIcon priority={priority} />;

  if (!canUpdate) {
    return (
      <Tooltip>
        <TooltipTrigger
          render={(
            <span aria-label={label}>
              {icon}
            </span>
          )}
        />
        <TooltipPopup>{label}</TooltipPopup>
      </Tooltip>
    );
  }

  return (
    <div className={slots.state()}>
      <Dropdown open={open} onOpenChange={setOpen}>
        <Tooltip>
          <TooltipTrigger
            render={(
              <DropdownTrigger
                disabled={isUpdating}
                render={(
                  <button
                    type="button"
                    className={slots.chipTrigger()}
                    aria-label={label}
                  >
                    {icon}
                  </button>
                )}
              />
            )}
          />
          <TooltipPopup>{label}</TooltipPopup>
        </Tooltip>
        <DropdownPopup size={1} align="start">
          <TaskPriorityMenu
            priority={priority}
            slots={slots}
            onChange={(next) => {
              onPriorityChange(next);
              setOpen(false);
            }}
          />
        </DropdownPopup>
      </Dropdown>
    </div>
  );
}

function TaskListItemState({
  displayState,
  taskState,
  canUpdate,
  isUpdating,
  slots,
  onStateChange,
}: {
  displayState: TaskState;
  taskState: TaskState;
  canUpdate: boolean;
  isUpdating?: boolean;
  slots: ReturnType<typeof taskListItem>;
  onStateChange: (state: TaskState) => void;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const [open, setOpen] = useState(false);
  const label = t(`states.${taskStateKeys[displayState]}`);
  const icon = <TaskStateIcon state={displayState} />;

  if (!canUpdate) {
    return (
      <Tooltip>
        <TooltipTrigger
          render={(
            <span className={slots.state()} aria-label={label}>
              {icon}
            </span>
          )}
        />
        <TooltipPopup>{label}</TooltipPopup>
      </Tooltip>
    );
  }

  return (
    <div className={slots.state()}>
      <Dropdown open={open} onOpenChange={setOpen}>
        <Tooltip>
          <TooltipTrigger
            render={(
              <DropdownTrigger
                disabled={isUpdating}
                render={(
                  <button
                    type="button"
                    className={slots.chipTrigger()}
                    aria-label={label}
                  >
                    {icon}
                  </button>
                )}
              />
            )}
          />
          <TooltipPopup>{label}</TooltipPopup>
        </Tooltip>
        <DropdownPopup size={1} align="start">
          <TaskStateMenu
            displayState={displayState}
            taskState={taskState}
            slots={slots}
            onChange={(state) => {
              onStateChange(state);
              setOpen(false);
            }}
          />
        </DropdownPopup>
      </Dropdown>
    </div>
  );
}

function TaskListItemEstimate({
  timeEstimate,
  label,
  canUpdate,
  isUpdating,
  slots,
  onTimeEstimateChange,
}: {
  timeEstimate?: string | null;
  label: string | null;
  canUpdate: boolean;
  isUpdating?: boolean;
  slots: ReturnType<typeof taskListItem>;
  onTimeEstimateChange: (timeEstimate: string | null) => void;
}): ReactNode {
  const { t, i18n } = useTranslation(TASKS_NAMESPACE);
  const [open, setOpen] = useState(false);
  const chip = (
    <span className={slots.meta()}>
      <ClockIcon className={slots.metaIcon()} weight="bold" />
      {label && (
        <Text size={1} color="current">
          {label}
        </Text>
      )}
    </span>
  );

  if (!canUpdate) {
    return (
      <span className={slots.meta()} title={t("metadata.timeEstimate")}>
        <ClockIcon className={slots.metaIcon()} weight="bold" />
        {label && (
          <Text size={1} color="current">
            {label}
          </Text>
        )}
      </span>
    );
  }

  const options = estimateOptions(timeEstimate);

  return (
    <div className={slots.state()}>
      <Dropdown open={open} onOpenChange={setOpen}>
        <Tooltip>
          <TooltipTrigger
            render={(
              <DropdownTrigger
                disabled={isUpdating}
                render={(
                  <button
                    type="button"
                    className={`${slots.meta()} ${slots.chipTrigger()}`}
                    aria-label={t("metadata.timeEstimate")}
                  >
                    {chip}
                  </button>
                )}
              />
            )}
          />
          <TooltipPopup>{t("metadata.timeEstimate")}</TooltipPopup>
        </Tooltip>
        <DropdownPopup size={1} align="end">
          <TaskEstimateMenu
            timeEstimate={timeEstimate}
            options={options}
            locale={i18n.language}
            onChange={(next) => {
              onTimeEstimateChange(next);
              setOpen(false);
            }}
          />
        </DropdownPopup>
      </Dropdown>
    </div>
  );
}

function TaskListItemDeadline({
  deadline,
  isOverdue,
  canUpdate,
  isUpdating,
  slots,
  onDeadlineChange,
}: {
  deadline?: string | null;
  isOverdue: boolean;
  canUpdate: boolean;
  isUpdating?: boolean;
  slots: ReturnType<typeof taskListItem>;
  onDeadlineChange: (deadline: string | null) => void;
}): ReactNode {
  const { t, i18n } = useTranslation(TASKS_NAMESPACE);
  const [open, setOpen] = useState(false);
  const label = deadline == null ? null : deadlineDateFormat(i18n.language, deadline);
  const icon = isOverdue
    ? <CalendarXIcon className={slots.deadlineIcon()} weight="bold" />
    : <CalendarBlankIcon className={slots.deadlineIcon()} weight="bold" />;

  function commitDeadline(next: string | null) {
    if (toDateInput(next) === toDateInput(deadline)) {
      return;
    }
    onDeadlineChange(next);
    setOpen(false);
  }

  if (!canUpdate) {
    return (
      <span className={slots.deadline()} title={t("metadata.deadline")}>
        {icon}
        {label && (
          <Text size={1} color="current">
            <time dateTime={deadline ?? undefined}>
              {label}
            </time>
          </Text>
        )}
      </span>
    );
  }

  return (
    <Dropdown
      open={open}
      onOpenChange={(next, details) => {
        if (!next && isNestedMenuDismiss(details)) {
          return;
        }
        setOpen(next);
      }}
    >
      <Tooltip>
        <TooltipTrigger
          render={(
            <DropdownTrigger
              disabled={isUpdating}
              render={(
                <button
                  type="button"
                  className={`${slots.deadline()} ${slots.chipTrigger()}`}
                  aria-label={t("metadata.deadline")}
                >
                  {icon}
                  {label && (
                    <Text size={1} color="current">
                      <time dateTime={deadline ?? undefined}>
                        {label}
                      </time>
                    </Text>
                  )}
                </button>
              )}
            />
          )}
        />
        <TooltipPopup>{t("metadata.deadline")}</TooltipPopup>
      </Tooltip>
      <DropdownPopup size={1} align="end">
        <TaskDeadlineMenu
          deadline={deadline}
          locale={i18n.language}
          slots={slots}
          onCommit={commitDeadline}
        />
      </DropdownPopup>
    </Dropdown>
  );
}

function TaskListItemAssignee({
  assignee,
  canUpdate,
  isUpdating,
  className,
  renderAssigneeMenu,
}: {
  assignee?: TaskListItemAssignee | null;
  canUpdate: boolean;
  isUpdating?: boolean;
  className: string;
  renderAssigneeMenu?: () => ReactNode;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const [open, setOpen] = useState(false);
  const label = assignee?.fullName ?? t("unassigned");
  const avatar = assignee == null
    ? (
        <UserCircleDashedIcon
          className="size-6 shrink-0 text-sand-8"
          aria-hidden
        />
      )
    : (
        <Avatar
          size={1}
          radius="full"
          name={assignee.fullName}
          email={assignee.emailAddress}
          src={assignee.avatarUrl}
        />
      );

  if (canUpdate && renderAssigneeMenu != null) {
    return (
      <div className={className}>
        <Dropdown
          open={open}
          onOpenChange={(next, details) => {
            if (!next && isNestedMenuDismiss(details)) {
              return;
            }
            setOpen(next);
          }}
        >
          <Tooltip>
            <TooltipTrigger
              render={(
                <DropdownTrigger
                  disabled={isUpdating}
                  render={(
                    <button
                      type="button"
                      className="appearance-none cursor-pointer border-0 bg-transparent p-0 text-inherit"
                      aria-label={label}
                    >
                      {avatar}
                    </button>
                  )}
                />
              )}
            />
            <TooltipPopup>{label}</TooltipPopup>
          </Tooltip>
          <DropdownPopup size={1} align="end">
            {renderAssigneeMenu()}
          </DropdownPopup>
        </Dropdown>
      </div>
    );
  }

  return (
    <Tooltip>
      <TooltipTrigger
        render={(
          <span className={className} aria-label={label}>
            {avatar}
          </span>
        )}
      />
      <TooltipPopup>{label}</TooltipPopup>
    </Tooltip>
  );
}

function TaskListItemMenu({
  displayState,
  taskState,
  priority,
  deadline,
  timeEstimate,
  canUpdate,
  canUpdateAssignee,
  isUpdating,
  slots,
  renderAssigneeMenu,
  extraMenu,
  onStateChange,
  onPriorityChange,
  onDeadlineChange,
  onTimeEstimateChange,
  onDelete,
}: {
  displayState: TaskState;
  taskState: TaskState;
  priority: TaskPriority;
  deadline?: string | null;
  timeEstimate?: string | null;
  canUpdate: boolean;
  canUpdateAssignee: boolean;
  isUpdating?: boolean;
  slots: ReturnType<typeof taskListItem>;
  renderAssigneeMenu?: () => ReactNode;
  extraMenu?: ReactNode;
  onStateChange: (state: TaskState) => void;
  onPriorityChange: (priority: TaskPriority) => void;
  onDeadlineChange: (deadline: string | null) => void;
  onTimeEstimateChange: (timeEstimate: string | null) => void;
  onDelete?: () => void;
}): ReactNode {
  const { t, i18n } = useTranslation(TASKS_NAMESPACE);
  const [open, setOpen] = useState(false);
  const showAssigneeMenu = canUpdateAssignee && renderAssigneeMenu != null;
  const showFields = canUpdate || showAssigneeMenu;
  const showDelete = onDelete != null;
  const showExtra = extraMenu != null;
  if (!showFields && !showDelete && !showExtra) {
    return null;
  }

  const estimateValues = estimateOptions(timeEstimate);

  function commitDeadline(next: string | null) {
    if (toDateInput(next) === toDateInput(deadline)) {
      return;
    }
    onDeadlineChange(next);
    setOpen(false);
  }

  return (
    <div className={slots.more()}>
      <Dropdown
        open={open}
        onOpenChange={(next, details) => {
          if (!next && isNestedMenuDismiss(details)) {
            return;
          }
          setOpen(next);
        }}
      >
        <DropdownTrigger
          disabled={isUpdating}
          render={(
            <IconButton
              size={1}
              variant="ghost"
              color="neutral"
              aria-label={t("actions.more")}
            >
              <DotsThreeVerticalIcon />
            </IconButton>
          )}
        />
        <DropdownPopup size={1} align="end">
          {canUpdate && (
            <>
              <DropdownSubmenu>
                <DropdownSubmenuTrigger>{t("fields.state")}</DropdownSubmenuTrigger>
                <DropdownPopup size={1} align="start" side="right">
                  <TaskStateMenu
                    displayState={displayState}
                    taskState={taskState}
                    slots={slots}
                    onChange={(state) => {
                      onStateChange(state);
                      setOpen(false);
                    }}
                  />
                </DropdownPopup>
              </DropdownSubmenu>
              <DropdownSubmenu>
                <DropdownSubmenuTrigger>{t("fields.priority")}</DropdownSubmenuTrigger>
                <DropdownPopup size={1} align="start" side="right">
                  <TaskPriorityMenu
                    priority={priority}
                    slots={slots}
                    onChange={(next) => {
                      onPriorityChange(next);
                      setOpen(false);
                    }}
                  />
                </DropdownPopup>
              </DropdownSubmenu>
              <DropdownSubmenu>
                <DropdownSubmenuTrigger>{t("fields.deadline")}</DropdownSubmenuTrigger>
                <DropdownPopup size={1} align="start" side="right">
                  <TaskDeadlineMenu
                    deadline={deadline}
                    locale={i18n.language}
                    slots={slots}
                    onCommit={commitDeadline}
                  />
                </DropdownPopup>
              </DropdownSubmenu>
              <DropdownSubmenu>
                <DropdownSubmenuTrigger>{t("fields.timeEstimate")}</DropdownSubmenuTrigger>
                <DropdownPopup size={1} align="start" side="right">
                  <TaskEstimateMenu
                    timeEstimate={timeEstimate}
                    options={estimateValues}
                    locale={i18n.language}
                    onChange={(next) => {
                      onTimeEstimateChange(next);
                      setOpen(false);
                    }}
                  />
                </DropdownPopup>
              </DropdownSubmenu>
            </>
          )}
          {showAssigneeMenu && (
            <DropdownSubmenu>
              <DropdownSubmenuTrigger>{t("fields.assignee")}</DropdownSubmenuTrigger>
              <DropdownPopup size={1} align="start" side="right">
                {renderAssigneeMenu()}
              </DropdownPopup>
            </DropdownSubmenu>
          )}
          {(showFields && (showExtra || showDelete)) && <DropdownSeparator />}
          {extraMenu}
          {showDelete && (
            <DropdownItem
              color="error"
              iconStart={<TrashIcon />}
              onClick={() => {
                onDelete();
              }}
            >
              {t("actions.delete")}
            </DropdownItem>
          )}
        </DropdownPopup>
      </Dropdown>
    </div>
  );
}

function TaskDeadlineMenu({
  deadline,
  locale,
  slots,
  onCommit,
}: {
  deadline?: string | null;
  locale: string;
  slots: ReturnType<typeof taskListItem>;
  onCommit: (next: string | null) => void;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);

  return (
    <>
      {deadlinePresets.map(preset => (
        <DropdownItem
          key={preset}
          iconStart={<CalendarBlankIcon className="size-4" weight="bold" />}
          shortcut={(
            <span className="ml-6">
              {compactDateFormat(locale, deadlinePresetDate(preset, locale))}
            </span>
          )}
          onClick={() => {
            onCommit(deadlineFromPreset(preset, locale));
          }}
        >
          {t(`deadlinePicker.${preset}`)}
        </DropdownItem>
      ))}
      <DropdownSeparator />
      <DropdownSubmenu>
        <DropdownSubmenuTrigger
          iconStart={<CalendarBlankIcon className="size-4" weight="bold" />}
        >
          {t("deadlinePicker.custom")}
        </DropdownSubmenuTrigger>
        <DropdownPopup side="right" align="start" size={1}>
          <div className={slots.deadlineCustom()}>
            <DateCalendar
              size={1}
              locale={locale}
              value={toDateInput(deadline)}
              onValueChange={(next) => {
                onCommit(next === "" ? null : formatDatetime(next) ?? null);
              }}
            />
          </div>
        </DropdownPopup>
      </DropdownSubmenu>
      <DropdownSeparator />
      <DropdownItem
        iconStart={<CalendarXIcon className="size-4" weight="bold" />}
        onClick={() => {
          onCommit(null);
        }}
      >
        {t("deadlinePicker.none")}
      </DropdownItem>
    </>
  );
}

function priorityOption(
  item: TaskPriority,
  slots: ReturnType<typeof taskListItem>,
  t: (key: string) => string,
): ReactNode {
  return (
    <span className={slots.stateOption()}>
      <TaskPriorityIcon priority={item} />
      {t(`priorities.${taskPriorityKeys[item]}`)}
    </span>
  );
}

function stateOption(
  state: TaskState,
  slots: ReturnType<typeof taskListItem>,
  t: (key: string) => string,
): ReactNode {
  return (
    <span className={slots.stateOption()}>
      <TaskStateIcon state={state} />
      {t(`states.${taskStateKeys[state]}`)}
    </span>
  );
}

function TaskStateMenu({
  displayState,
  taskState,
  slots,
  onChange,
}: {
  displayState: TaskState;
  taskState: TaskState;
  slots: ReturnType<typeof taskListItem>;
  onChange: (state: TaskState) => void;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);

  return (
    <DropdownRadioGroup
      value={displayState}
      onValueChange={(state) => {
        if (typeof state !== "string" || state === taskState) {
          return;
        }
        onChange(state as TaskState);
      }}
    >
      {taskStates.map(state => (
        <DropdownRadioItem key={state} value={state}>
          {stateOption(state, slots, t)}
        </DropdownRadioItem>
      ))}
    </DropdownRadioGroup>
  );
}

function TaskPriorityMenu({
  priority,
  slots,
  onChange,
}: {
  priority: TaskPriority;
  slots: ReturnType<typeof taskListItem>;
  onChange: (priority: TaskPriority) => void;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);

  return (
    <DropdownRadioGroup
      value={priority}
      onValueChange={(next) => {
        if (typeof next !== "string" || next === priority) {
          return;
        }
        onChange(next as TaskPriority);
      }}
    >
      {taskPriorities.map(item => (
        <DropdownRadioItem key={item} value={item}>
          {priorityOption(item, slots, t)}
        </DropdownRadioItem>
      ))}
    </DropdownRadioGroup>
  );
}

function TaskEstimateMenu({
  timeEstimate,
  options,
  locale,
  onChange,
}: {
  timeEstimate?: string | null;
  options: readonly string[];
  locale: string;
  onChange: (timeEstimate: string | null) => void;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const selected = matchingEstimatePreset(timeEstimate) ?? noneEstimateValue;

  function commit(next: string) {
    const resolved = next === noneEstimateValue ? null : next;
    if (resolved === timeEstimate || (
      resolved != null
      && timeEstimate != null
      && taskDurationsEqual(resolved, timeEstimate)
    )) {
      return;
    }
    onChange(resolved);
  }

  return (
    <DropdownRadioGroup
      value={selected}
      onValueChange={(next) => {
        if (typeof next !== "string") {
          return;
        }
        commit(next);
      }}
    >
      <DropdownRadioItem
        value={noneEstimateValue}
        iconStart={<ProhibitIcon className="size-4" weight="bold" />}
      >
        {t("durationPicker.none")}
      </DropdownRadioItem>
      {options.map(value => (
        <DropdownRadioItem
          key={value}
          value={value}
          iconStart={<ClockIcon className="size-4" weight="bold" />}
        >
          {formatDuration(value, t, { style: "short", locale })}
        </DropdownRadioItem>
      ))}
    </DropdownRadioGroup>
  );
}

function matchingEstimatePreset(value?: string | null): string | null {
  if (value == null) {
    return null;
  }
  return estimatePresets.find(preset => taskDurationsEqual(preset, value)) ?? value;
}

function estimateOptions(value?: string | null): string[] {
  const selected = matchingEstimatePreset(value);
  if (selected == null || estimatePresets.includes(selected)) {
    return estimatePresets;
  }
  return [selected, ...estimatePresets];
}

const deadlinePresets = ["today", "tomorrow", "endOfWeek", "nextWeek"] as const;

function deadlinePresetDate(
  preset: (typeof deadlinePresets)[number],
  locale: string,
): string {
  return toDateInput(deadlineFromPreset(preset, locale));
}

function dateInputPlusDays(days: number): string {
  const today = parseDate(todayAsDateInput());
  const next = new Date(today.getFullYear(), today.getMonth(), today.getDate() + days);
  const year = String(next.getFullYear()).padStart(4, "0");
  const month = String(next.getMonth() + 1).padStart(2, "0");
  const day = String(next.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function deadlineFromPreset(preset: string, locale: string): string | null {
  if (preset === "none") {
    return null;
  }
  if (preset === "endOfWeek") {
    return formatDatetime(endOfWeekDateInput(locale)) ?? null;
  }
  const days = preset === "today" ? 0 : preset === "tomorrow" ? 1 : preset === "nextWeek" ? 7 : null;
  if (days == null) {
    return null;
  }
  return formatDatetime(dateInputPlusDays(days)) ?? null;
}

function endOfWeekDateInput(locale: string): string {
  const today = parseDate(todayAsDateInput());
  let delta = weekEndDay(locale) - today.getDay();
  if (delta < 0) {
    delta += 7;
  }
  return dateInputPlusDays(delta);
}

function weekEndDay(locale: string): number {
  try {
    const intlLocale = new Intl.Locale(locale);
    const withWeekInfo = intlLocale as Intl.Locale & {
      weekInfo?: { firstDay?: number };
      getWeekInfo?: () => { firstDay?: number };
    };
    const weekInfo = typeof withWeekInfo.getWeekInfo === "function"
      ? withWeekInfo.getWeekInfo()
      : withWeekInfo.weekInfo;
    const firstDay = weekInfo?.firstDay;
    if (firstDay == null) {
      return 0;
    }
    const start = firstDay === 7 ? 0 : firstDay;
    return (start + 6) % 7;
  } catch {
    return 0;
  }
}

function isNestedMenuDismiss(details: {
  reason: string;
  event?: Event;
}): boolean {
  if (details.reason !== "focus-out" && details.reason !== "outside-press") {
    return false;
  }
  const target = details.reason === "focus-out" && details.event instanceof FocusEvent
    ? details.event.relatedTarget
    : details.event?.target;
  return target instanceof Element
    && target.closest("[role='dialog'], [role='menu'], [role='listbox'], [role='option'], input, textarea") != null;
}
