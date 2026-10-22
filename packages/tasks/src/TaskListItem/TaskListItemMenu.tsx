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

import { DotsThreeVerticalIcon, TrashIcon } from "@phosphor-icons/react";
import { toDateInput } from "@probo/helpers";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { DropdownSubmenu } from "@probo/ui/src/v2/Dropdown/DropdownSubmenu";
import { DropdownSubmenuTrigger } from "@probo/ui/src/v2/Dropdown/DropdownSubmenuTrigger";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import type { ReactNode } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "../i18n";
import { type TaskPriority, type TaskState } from "../taskState";
import { taskListItem } from "../variants";

import { isNestedMenuDismiss } from "./isNestedMenuDismiss";
import { TaskDeadlineMenu } from "./TaskDeadlineMenu";
import { TaskEstimateCustomDialog } from "./TaskEstimateCustomDialog";
import { TaskEstimateMenu } from "./TaskEstimateMenu";
import { TaskPriorityMenu } from "./TaskPriorityMenu";
import { TaskStateMenu } from "./TaskStateMenu";

export function TaskListItemMenu({
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
  const [customOpen, setCustomOpen] = useState(false);
  const showAssigneeMenu = canUpdateAssignee && renderAssigneeMenu != null;
  const showFields = canUpdate || showAssigneeMenu;
  const showDelete = onDelete != null;
  const showExtra = extraMenu != null;
  if (!showFields && !showDelete && !showExtra) {
    return null;
  }

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
                    locale={i18n.language}
                    onChange={(next) => {
                      onTimeEstimateChange(next);
                      setOpen(false);
                    }}
                    onCustom={() => {
                      setOpen(false);
                      setCustomOpen(true);
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
      <TaskEstimateCustomDialog
        open={customOpen}
        timeEstimate={timeEstimate}
        onOpenChange={setCustomOpen}
        onSave={onTimeEstimateChange}
      />
    </div>
  );
}
