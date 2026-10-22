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

import { CalendarBlankIcon, CalendarXIcon } from "@phosphor-icons/react";
import { toDateInput } from "@probo/helpers";
import { deadlineDateFormat } from "@probo/i18n";
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
import { TaskDeadlineMenu } from "./TaskDeadlineMenu";

export function TaskListItemDeadline({
  deadline,
  isOverdue,
  canUpdate,
  isUpdating,
  size = 1,
  placeholder,
  slots,
  onDeadlineChange,
}: {
  deadline?: string | null;
  isOverdue: boolean;
  canUpdate: boolean;
  isUpdating?: boolean;
  size?: 1 | 2;
  placeholder?: string;
  slots: ReturnType<typeof taskListItem>;
  onDeadlineChange: (deadline: string | null) => void;
}): ReactNode {
  const { t, i18n } = useTranslation(TASKS_NAMESPACE);
  const [open, setOpen] = useState(false);
  const label = deadline == null ? null : deadlineDateFormat(i18n.language, deadline);
  const tooltip = label ?? placeholder ?? t("metadata.deadline");
  const shown = label ?? placeholder ?? null;
  const icon = isOverdue
    ? <CalendarXIcon className={slots.deadlineIcon()} weight="bold" />
    : <CalendarBlankIcon className={slots.deadlineIcon()} weight="bold" />;

  function commitDeadline(next: string | null) {
    if (toDateInput(next) === toDateInput(deadline)) {
      setOpen(false);
      return;
    }
    onDeadlineChange(next);
    setOpen(false);
  }

  if (!canUpdate) {
    return (
      <span className={slots.deadline()} title={tooltip}>
        {icon}
        {shown != null && (
          <Text size={size} color={label == null ? "faint" : "current"}>
            {label == null
              ? shown
              : (
                  <time dateTime={deadline ?? undefined}>
                    {shown}
                  </time>
                )}
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
                  aria-label={label == null ? tooltip : `${tooltip}, ${label}`}
                >
                  {icon}
                  {shown != null && (
                    <Text size={size} color={label == null ? "faint" : "current"}>
                      {label == null
                        ? shown
                        : (
                            <time dateTime={deadline ?? undefined}>
                              {shown}
                            </time>
                          )}
                    </Text>
                  )}
                </button>
              )}
            />
          )}
        />
        <TooltipPopup>{tooltip}</TooltipPopup>
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
