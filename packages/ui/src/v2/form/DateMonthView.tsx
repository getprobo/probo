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

import { CaretLeftIcon, CaretRightIcon } from "@phosphor-icons/react";
import { useState } from "react";

import { IconButton } from "../IconButton/IconButton";
import { Select } from "../Select/Select";
import { SelectItem } from "../Select/SelectItem";
import { SelectPopup } from "../Select/SelectPopup";
import { SelectTrigger } from "../Select/SelectTrigger";

import {
  isOutOfRange,
  monthAfterMax,
  monthBeforeMin,
  monthGrid,
  monthLabels,
  parseIsoDate,
  shiftMonth,
  toIsoDate,
  visibleFrom,
  weekdayLabels,
  weekStartDay,
  yearRange,
} from "./dateModel";
import { dateField, dateFieldDay } from "./variants";

export interface DateMonthViewProps {
  // Selected ISO `YYYY-MM-DD`, or `""` when empty.
  value: string;
  locale?: string;
  min?: string;
  max?: string;
  disabled?: boolean;
  // When this changes, snap the visible month back to `value` (or today).
  resetKey?: string | number | boolean;
  onDaySelect?: (iso: string) => void;
}

// Month/year header and day grid shared by DateField and DateCalendar.
export function DateMonthView(props: DateMonthViewProps) {
  const {
    value,
    locale,
    min,
    max,
    disabled = false,
    resetKey,
    onDaySelect,
  } = props;

  const [visible, setVisible] = useState(() => visibleFrom(value));
  const [source, setSource] = useState({ value, resetKey });
  const { calendar, header, monthSelect, yearSelect, weekdays, weekday, grid } = dateField();

  if (value !== source.value || resetKey !== source.resetKey) {
    setSource({ value, resetKey });
    setVisible(visibleFrom(value));
  }

  const minDate = parseIsoDate(min ?? "");
  const maxDate = parseIsoDate(max ?? "");
  const months = monthLabels(locale);
  const weekStartsOn = weekStartDay(locale);
  const days = monthGrid(visible.year, visible.month, weekStartsOn);
  const years = yearRange(visible.year, minDate, maxDate);
  const todayIso = toIsoDate(new Date());

  return (
    <div className={calendar()}>
      <div className={header()}>
        <IconButton
          type="button"
          size={1}
          variant="ghost"
          color="neutral"
          aria-label="Previous month"
          disabled={disabled || monthBeforeMin(visible, minDate)}
          onClick={() => setVisible(shiftMonth(visible, -1))}
        >
          <CaretLeftIcon />
        </IconButton>
        <div className={monthSelect()}>
          <Select
            value={visible.month}
            disabled={disabled}
            onValueChange={(month) => {
              if (month == null) {
                return;
              }
              setVisible({ year: visible.year, month });
            }}
          >
            <SelectTrigger size={1} aria-label="Month">
              {(month: number | null) => (month == null ? null : months[month])}
            </SelectTrigger>
            <SelectPopup>
              {months.map((label, month) => (
                <SelectItem key={label} value={month}>
                  {label}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <div className={yearSelect()}>
          <Select
            value={visible.year}
            disabled={disabled}
            onValueChange={(year) => {
              if (year == null) {
                return;
              }
              setVisible({ year, month: visible.month });
            }}
          >
            <SelectTrigger size={1} aria-label="Year">
              {(year: number | null) => year}
            </SelectTrigger>
            <SelectPopup>
              {years.map(year => (
                <SelectItem key={year} value={year}>
                  {year}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <IconButton
          type="button"
          size={1}
          variant="ghost"
          color="neutral"
          aria-label="Next month"
          disabled={disabled || monthAfterMax(visible, maxDate)}
          onClick={() => setVisible(shiftMonth(visible, 1))}
        >
          <CaretRightIcon />
        </IconButton>
      </div>
      <div className={weekdays()}>
        {weekdayLabels(locale, weekStartsOn).map((label, index) => (
          <span key={`${label}-${index}`} className={weekday()}>
            {label}
          </span>
        ))}
      </div>
      <div className={grid()}>
        {days.map((day) => {
          const iso = toIsoDate(day);
          const isOutside = day.getMonth() !== visible.month;
          const isSelected = iso === value;
          const isDisabled = disabled || isOutOfRange(iso, minDate, maxDate);

          return (
            <button
              key={iso}
              type="button"
              disabled={isDisabled}
              aria-pressed={isSelected}
              aria-current={iso === todayIso ? "date" : undefined}
              className={dateFieldDay({
                selected: isSelected,
                today: iso === todayIso,
                outside: isOutside,
              })}
              onClick={() => onDaySelect?.(iso)}
            >
              {day.getDate()}
            </button>
          );
        })}
      </div>
    </div>
  );
}
