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

import { taskListItem } from "@probo/tasks";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactNode, useState } from "react";

export interface FindingPropertyMenuProps {
  label: string;
  icon: ReactNode;
  canUpdate: boolean;
  disabled?: boolean;
  faint?: boolean;
  menu: ((close: () => void) => ReactNode) | null;
  onOpen?: () => void;
}

export function FindingPropertyMenu({
  label,
  icon,
  canUpdate,
  disabled,
  faint = false,
  menu,
  onOpen,
}: FindingPropertyMenuProps) {
  const slots = taskListItem({ field: true });
  const [open, setOpen] = useState(false);
  const value = (
    <span className={slots.stateOption()}>
      {icon}
      <Text size={2} color={faint ? "faint" : undefined}>{label}</Text>
    </span>
  );

  if (!canUpdate || (menu == null && onOpen == null)) {
    return (
      <Tooltip>
        <TooltipTrigger
          render={(
            <span className={slots.stateOption()} aria-label={label}>
              {value}
            </span>
          )}
        />
        <TooltipPopup>{label}</TooltipPopup>
      </Tooltip>
    );
  }

  if (menu == null) {
    return (
      <Tooltip>
        <TooltipTrigger
          render={(
            <button
              type="button"
              className={slots.chipTrigger()}
              aria-label={label}
              disabled={disabled}
              onClick={onOpen}
            >
              {value}
            </button>
          )}
        />
        <TooltipPopup>{label}</TooltipPopup>
      </Tooltip>
    );
  }

  return (
    <div className={slots.stateOption()}>
      <Dropdown open={open} onOpenChange={setOpen}>
        <Tooltip>
          <TooltipTrigger
            render={(
              <DropdownTrigger
                disabled={disabled}
                render={(
                  <button type="button" className={slots.chipTrigger()} aria-label={label}>
                    {value}
                  </button>
                )}
              />
            )}
          />
          <TooltipPopup>{label}</TooltipPopup>
        </Tooltip>
        <DropdownPopup size={1} align="start">
          {menu(() => {
            setOpen(false);
          })}
        </DropdownPopup>
      </Dropdown>
    </div>
  );
}
