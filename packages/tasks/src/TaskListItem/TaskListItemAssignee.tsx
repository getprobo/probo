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

import { UserCircleDashedIcon } from "@phosphor-icons/react";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "../i18n";
import { taskListItem } from "../variants";

import { isNestedMenuDismiss } from "./isNestedMenuDismiss";
import type { TaskListItemAssignee as Assignee } from "./TaskListItem";

export function TaskListItemAssignee({
  assignee,
  canUpdate,
  isUpdating,
  showLabel = false,
  placeholder,
  className,
  renderAssigneeMenu,
}: {
  assignee?: Assignee | null;
  canUpdate: boolean;
  isUpdating?: boolean;
  showLabel?: boolean;
  placeholder?: string;
  className: string;
  renderAssigneeMenu?: () => ReactNode;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const [open, setOpen] = useState(false);
  const slots = taskListItem({ field: showLabel });
  const name = assignee?.fullName ?? null;
  const label = name ?? (showLabel ? placeholder : null) ?? t("unassigned");
  const tooltip = showLabel && placeholder != null ? placeholder : label;
  const avatar = assignee == null
    ? (
        <UserCircleDashedIcon
          className={showLabel ? slots.assigneeIcon() : "size-6 shrink-0 text-sand-8"}
          aria-hidden
        />
      )
    : (
        <Avatar
          size={1}
          radius="full"
          className={showLabel ? slots.assigneeAvatar() : undefined}
          name={assignee.fullName}
          email={assignee.emailAddress}
          src={assignee.avatarUrl}
        />
      );
  const value = showLabel
    ? (
        <span className={slots.stateOption()}>
          {avatar}
          <Text size={2} color={name == null ? "faint" : undefined} className={slots.assigneeLabel()}>{label}</Text>
        </span>
      )
    : avatar;

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
                      className={showLabel
                        ? slots.chipTrigger()
                        : "appearance-none cursor-pointer border-0 bg-transparent p-0 text-inherit"}
                      aria-label={label}
                    >
                      {value}
                    </button>
                  )}
                />
              )}
            />
            <TooltipPopup>{tooltip}</TooltipPopup>
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
            {value}
          </span>
        )}
      />
      <TooltipPopup>{tooltip}</TooltipPopup>
    </Tooltip>
  );
}
