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
import type { ReactNode } from "react";
import { useState } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "../i18n";

import { isNestedMenuDismiss } from "./isNestedMenuDismiss";
import type { TaskListItemAssignee as Assignee } from "./TaskListItem";

export function TaskListItemAssignee({
  assignee,
  canUpdate,
  isUpdating,
  className,
  renderAssigneeMenu,
}: {
  assignee?: Assignee | null;
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
