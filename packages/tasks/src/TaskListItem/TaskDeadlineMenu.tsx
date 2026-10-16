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
import { formatDatetime, toDateInput } from "@probo/helpers";
import { compactDateFormat } from "@probo/i18n";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { DropdownSubmenu } from "@probo/ui/src/v2/Dropdown/DropdownSubmenu";
import { DropdownSubmenuTrigger } from "@probo/ui/src/v2/Dropdown/DropdownSubmenuTrigger";
import { DateCalendar } from "@probo/ui/src/v2/form/DateCalendar";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { TASKS_NAMESPACE } from "../i18n";
import { taskListItem } from "../variants";

import { deadlineFromPreset, deadlinePresetDate, deadlinePresets } from "./taskDeadlinePreset";

export function TaskDeadlineMenu({
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
