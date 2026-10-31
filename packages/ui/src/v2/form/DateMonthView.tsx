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
  dateRangeTone,
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

export interface DateMonthRange {
  start: string;
  end: string;
  // Hovered ISO day while the end is still open. Ignored once `end` is set.
  preview?: string;
}

export interface DateMonthViewProps {
  // Selected ISO `YYYY-MM-DD`, or `""` when empty. Ignored for highlighting
  // when `range` is set; still chooses the month the grid opens on.
  value: string;
  locale?: string;
  min?: string;
  max?: string;
  disabled?: boolean;
  size?: 1 | 2;
  // When this changes, snap the visible month back to `value` (or today).
  resetKey?: string | number | boolean;
  // Start/end highlighting. Omit for a single-day picker.
  range?: DateMonthRange;
  onDaySelect?: (iso: string) => void;
  onDayHover?: (iso: string) => void;
}

// Month/year header and day grid shared by DateField, DateCalendar, and DateRangeField.
export function DateMonthView(props: DateMonthViewProps) {
  const {
    value,
    locale,
    min,
    max,
    disabled = false,
    size = 2,
    resetKey,
    range,
    onDaySelect,
    onDayHover,
  } = props;

  const [visible, setVisible] = useState(() => visibleFrom(value));
  const [source, setSource] = useState({ value, resetKey });
  const { calendar, header, monthSelect, yearSelect, weekdays, weekday, grid } = dateField({ size });

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
              {months.map((label, month) => {
                const start = new Date(visible.year, month, 1);
                const end = new Date(visible.year, month + 1, 0);
                if (minDate != null && end < minDate) {
                  return null;
                }
                if (maxDate != null && start > maxDate) {
                  return null;
                }
                return (
                  <SelectItem key={label} value={month}>
                    {label}
                  </SelectItem>
                );
              })}
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
              {years.map((year) => {
                if (minDate != null && year < minDate.getFullYear()) {
                  return null;
                }
                if (maxDate != null && year > maxDate.getFullYear()) {
                  return null;
                }
                return (
                  <SelectItem key={year} value={year}>
                    {year}
                  </SelectItem>
                );
              })}
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
      <div className={grid()} onMouseLeave={() => onDayHover?.("")}>
        {days.map((day) => {
          const iso = toIsoDate(day);
          const isOutside = day.getMonth() !== visible.month;
          const tone = range == null
            ? (iso === value ? "selected" : "day")
            : dateRangeTone(
                iso,
                range.start,
                range.end !== "" ? range.end : (range.preview ?? ""),
              );
          const isDisabled = disabled || isOutOfRange(iso, minDate, maxDate);
          const pressed = tone === "selected" || tone === "within" || tone === "start" || tone === "end" || tone === "single";

          return (
            <button
              key={iso}
              type="button"
              disabled={isDisabled}
              aria-label={day.toLocaleDateString(locale, { dateStyle: "full" })}
              aria-pressed={pressed}
              aria-current={iso === todayIso ? "date" : undefined}
              className={dateFieldDay({
                size,
                tone,
                today: iso === todayIso,
                outside: isOutside && !pressed,
              })}
              onMouseEnter={() => onDayHover?.(iso)}
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
