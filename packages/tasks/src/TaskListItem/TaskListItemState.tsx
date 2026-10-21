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
import { type TaskState, taskStateKeys } from "../taskState";
import { TaskStateIcon } from "../TaskStateIcon";
import { taskListItem } from "../variants";

import { TaskStateMenu } from "./TaskStateMenu";

function stateValue(
  icon: ReactNode,
  label: string,
  showLabel: boolean,
  slots: ReturnType<typeof taskListItem>,
): ReactNode {
  if (!showLabel) {
    return icon;
  }

  return (
    <span className={slots.stateOption()}>
      {icon}
      <Text size={2}>{label}</Text>
    </span>
  );
}

export function TaskListItemState({
  displayState,
  taskState,
  canUpdate,
  isUpdating,
  showLabel = false,
  placeholder,
  slots,
  onStateChange,
}: {
  displayState: TaskState;
  taskState: TaskState;
  canUpdate: boolean;
  isUpdating?: boolean;
  showLabel?: boolean;
  placeholder?: string;
  slots: ReturnType<typeof taskListItem>;
  onStateChange: (state: TaskState) => void;
}): ReactNode {
  const { t } = useTranslation(TASKS_NAMESPACE);
  const [open, setOpen] = useState(false);
  const label = t(`states.${taskStateKeys[displayState]}`);
  const tooltip = placeholder ?? label;
  const icon = <TaskStateIcon state={displayState} />;
  const value = stateValue(icon, label, showLabel, slots);
  const frame = showLabel ? slots.stateOption() : slots.state();

  if (!canUpdate) {
    return (
      <Tooltip>
        <TooltipTrigger
          render={(
            <span className={frame} aria-label={label}>
              {value}
            </span>
          )}
        />
        <TooltipPopup>{tooltip}</TooltipPopup>
      </Tooltip>
    );
  }

  return (
    <div className={frame}>
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
                    {value}
                  </button>
                )}
              />
            )}
          />
          <TooltipPopup>{tooltip}</TooltipPopup>
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
