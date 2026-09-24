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

import { CalendarBlankIcon, CaretLeftIcon, CaretRightIcon, XIcon } from "@phosphor-icons/react";
import { type ChangeEvent, type KeyboardEvent, useState } from "react";

import { IconButton } from "../IconButton/IconButton";
import { Popover, type PopoverProps } from "../Popover/Popover";
import { PopoverPopup } from "../Popover/PopoverPopup";
import { PopoverTrigger } from "../Popover/PopoverTrigger";
import { Select } from "../Select/Select";
import { SelectItem } from "../Select/SelectItem";
import { SelectPopup } from "../Select/SelectPopup";
import { SelectTrigger } from "../Select/SelectTrigger";

import { dateField, dateFieldDay } from "./variants";

const YEAR_PAST = 40;
const YEAR_FUTURE = 10;
const ISO_DATE = /^(\d{4})-(\d{2})-(\d{2})$/;

type DatePart = "day" | "month" | "year";

export interface DateFieldProps {
  // Applied to the trigger row (the top-level element).
  "className"?: string;
  // Forwards onto the typed input so Field can associate its label.
  "id"?: string;
  // Native form field name; written to a hidden input.
  "name"?: string;
  // ISO `YYYY-MM-DD`, or `""` when empty.
  "value": string;
  // `Intl` locale for typed mask, month names, and weekdays.
  "locale"?: string;
  // Shown on the input when empty. Defaults to the locale mask (`mm/dd/yyyy`).
  "placeholder"?: string;
  "size"?: 1 | 2;
  // Surface treatment (defaults to "surface").
  "variant"?: "classic" | "surface";
  "disabled"?: boolean;
  // Inclusive ISO bound; days before this are not selectable.
  "min"?: string;
  // Inclusive ISO bound; days after this are not selectable.
  "max"?: string;
  // When true, a clear control appears while a date is selected.
  "nullable"?: boolean;
  "aria-invalid"?: boolean;
  "aria-describedby"?: string;
  "aria-required"?: boolean;
  "onValueChange"?: (value: string) => void;
}

// TextField-like date control: type digits and separators are inserted, or
// pick a day from the Popover calendar. Value is always an ISO date string.
export function DateField(props: DateFieldProps) {
  const {
    className,
    id,
    name,
    value,
    locale,
    placeholder,
    size,
    variant,
    disabled = false,
    min,
    max,
    nullable = false,
    "aria-invalid": ariaInvalid,
    "aria-describedby": ariaDescribedBy,
    "aria-required": ariaRequired,
    onValueChange,
  } = props;

  const pattern = datePattern(locale);
  const [open, setOpen] = useState(false);
  const [visible, setVisible] = useState(() => visibleFrom(value));
  const [draft, setDraft] = useState(() => isoToMasked(value, pattern));
  const [source, setSource] = useState({ value, locale });
  const {
    root,
    surface,
    iconTrigger,
    input,
    clear,
    calendar,
    header,
    monthSelect,
    yearSelect,
    weekdays,
    weekday,
    grid,
  } = dateField({ size, variant });

  if (value !== source.value || locale !== source.locale) {
    setSource({ value, locale });
    setDraft(isoToMasked(value, pattern));
  }

  const minDate = parseIsoDate(min ?? "");
  const maxDate = parseIsoDate(max ?? "");
  const months = monthLabels(locale);
  const weekStartsOn = weekStartDay(locale);
  const days = monthGrid(visible.year, visible.month, weekStartsOn);
  const years = yearRange(visible.year, minDate, maxDate);
  const todayIso = toIsoDate(new Date());

  function handleOpenChange(
    nextOpen: boolean,
    eventDetails: Parameters<NonNullable<PopoverProps["onOpenChange"]>>[1],
  ) {
    if (!nextOpen && isNestedSelectDismiss(eventDetails)) {
      return;
    }
    setOpen(nextOpen);
    if (nextOpen) {
      setVisible(visibleFrom(value));
    }
  }

  function commitIso(iso: string) {
    onValueChange?.(iso);
    setVisible(visibleFrom(iso));
  }

  function selectDay(iso: string) {
    commitIso(iso);
    setOpen(false);
  }

  function clearValue() {
    onValueChange?.("");
  }

  function handleInputChange(event: ChangeEvent<HTMLInputElement>) {
    const raw = event.currentTarget.value;
    const trimmed = raw.trim();
    if (parseIsoDate(trimmed) != null && !isOutOfRange(trimmed, minDate, maxDate)) {
      commitIso(trimmed);
      return;
    }

    const digits = raw.replace(/\D/g, "").slice(0, 8);
    setDraft(maskDigits(digits, pattern));

    if (digits.length === 0) {
      if (nullable) {
        onValueChange?.("");
      }
      return;
    }

    if (digits.length !== 8) {
      return;
    }

    const iso = digitsToIso(digits, pattern);
    if (iso != null && !isOutOfRange(iso, minDate, maxDate)) {
      commitIso(iso);
    }
  }

  function handleInputBlur() {
    setDraft(isoToMasked(value, pattern));
  }

  function handleInputKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== "Backspace") {
      return;
    }
    const field = event.currentTarget;
    const start = field.selectionStart ?? 0;
    const end = field.selectionEnd ?? 0;
    if (start !== end || start === 0) {
      return;
    }
    const charBefore = draft[start - 1];
    if (charBefore == null || /\d/.test(charBefore)) {
      return;
    }
    event.preventDefault();
    const prefixDigits = draft.slice(0, start).replace(/\D/g, "").slice(0, -1);
    const suffixDigits = draft.slice(start).replace(/\D/g, "");
    const digits = `${prefixDigits}${suffixDigits}`.slice(0, 8);
    setDraft(maskDigits(digits, pattern));
    if (digits.length === 0 && nullable) {
      onValueChange?.("");
    }
  }

  return (
    <div className={root({ className })}>
      {name != null && <input type="hidden" name={name} value={value} disabled={disabled} />}
      <div className={surface()}>
        <Popover open={open} onOpenChange={handleOpenChange}>
          <PopoverTrigger
            disabled={disabled}
            className={iconTrigger()}
            aria-label="Choose date"
          >
            <CalendarBlankIcon />
          </PopoverTrigger>
          <PopoverPopup>
            <div className={calendar()}>
              <div className={header()}>
                <IconButton
                  type="button"
                  size={1}
                  variant="ghost"
                  color="neutral"
                  aria-label="Previous month"
                  disabled={monthBeforeMin(visible, minDate)}
                  onClick={() => setVisible(shiftMonth(visible, -1))}
                >
                  <CaretLeftIcon />
                </IconButton>
                <div className={monthSelect()}>
                  <Select
                    value={visible.month}
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
                  disabled={monthAfterMax(visible, maxDate)}
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
                  const isDisabled = isOutOfRange(iso, minDate, maxDate);

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
                      onClick={() => selectDay(iso)}
                    >
                      {day.getDate()}
                    </button>
                  );
                })}
              </div>
            </div>
          </PopoverPopup>
        </Popover>
        <input
          id={id}
          type="text"
          inputMode="numeric"
          autoComplete="off"
          spellCheck={false}
          disabled={disabled}
          maxLength={10}
          placeholder={placeholder ?? maskPlaceholder(pattern)}
          value={draft}
          aria-invalid={ariaInvalid}
          aria-describedby={ariaDescribedBy}
          aria-required={ariaRequired}
          className={input()}
          onChange={handleInputChange}
          onBlur={handleInputBlur}
          onKeyDown={handleInputKeyDown}
        />
        {nullable && value !== "" && !disabled && (
          <button
            type="button"
            className={clear()}
            aria-label="Clear date"
            onMouseDown={event => event.preventDefault()}
            onClick={clearValue}
          >
            <XIcon />
          </button>
        )}
      </div>
    </div>
  );
}

type YearMonth = {
  year: number;
  month: number;
};

type DatePattern = {
  order: DatePart[];
  separator: string;
};

function datePattern(locale: string | undefined): DatePattern {
  const parts = new Intl.DateTimeFormat(locale, {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).formatToParts(new Date(2020, 11, 31));

  const order: DatePart[] = [];
  let separator = "/";
  for (const part of parts) {
    if (part.type === "day" || part.type === "month" || part.type === "year") {
      order.push(part.type);
    } else if (part.type === "literal") {
      const trimmed = part.value.trim();
      if (trimmed !== "") {
        separator = trimmed;
      }
    }
  }

  if (order.length !== 3) {
    return { order: ["month", "day", "year"], separator: "/" };
  }
  return { order, separator };
}

function maskPlaceholder({ order, separator }: DatePattern): string {
  const labels: Record<DatePart, string> = {
    day: "dd",
    month: "mm",
    year: "yyyy",
  };
  return order.map(part => labels[part]).join(separator);
}

function maskDigits(digits: string, { order, separator }: DatePattern): string {
  let remaining = digits;
  let masked = "";
  for (let index = 0; index < order.length; index += 1) {
    if (remaining.length === 0) {
      break;
    }
    const part = order[index];
    if (part == null) {
      break;
    }
    const length = part === "year" ? 4 : 2;
    const chunk = remaining.slice(0, length);
    if (index > 0) {
      masked += separator;
    }
    masked += chunk;
    remaining = remaining.slice(length);
    if (chunk.length === length && index < order.length - 1 && remaining.length === 0) {
      masked += separator;
    }
  }
  return masked;
}

function isoToMasked(iso: string, pattern: DatePattern): string {
  const date = parseIsoDate(iso);
  if (date == null) {
    return "";
  }
  const values: Record<DatePart, string> = {
    day: String(date.getDate()).padStart(2, "0"),
    month: String(date.getMonth() + 1).padStart(2, "0"),
    year: String(date.getFullYear()).padStart(4, "0"),
  };
  return pattern.order.map(part => values[part]).join(pattern.separator);
}

function digitsToIso(digits: string, { order }: DatePattern): string | null {
  if (digits.length !== 8) {
    return null;
  }
  const values: Record<DatePart, number> = { day: 0, month: 0, year: 0 };
  let index = 0;
  for (const part of order) {
    const length = part === "year" ? 4 : 2;
    values[part] = Number(digits.slice(index, index + length));
    index += length;
  }
  const date = new Date(values.year, values.month - 1, values.day);
  if (
    date.getFullYear() !== values.year
    || date.getMonth() !== values.month - 1
    || date.getDate() !== values.day
  ) {
    return null;
  }
  return toIsoDate(date);
}

function isNestedSelectDismiss(
  eventDetails: Parameters<NonNullable<PopoverProps["onOpenChange"]>>[1],
): boolean {
  if (eventDetails.reason === "focus-out") {
    const relatedTarget = eventDetails.event instanceof FocusEvent
      ? eventDetails.event.relatedTarget
      : null;
    return relatedTarget instanceof Element
      && relatedTarget.closest("[role='listbox'], [role='option']") != null;
  }
  if (eventDetails.reason !== "outside-press") {
    return false;
  }
  const target = eventDetails.event.target;
  return target instanceof Element
    && target.closest("[role='listbox'], [role='option']") != null;
}

function parseIsoDate(value: string): Date | null {
  const match = ISO_DATE.exec(value);
  if (match == null) {
    return null;
  }
  const year = Number(match[1]);
  const month = Number(match[2]) - 1;
  const day = Number(match[3]);
  const date = new Date(year, month, day);
  if (date.getFullYear() !== year || date.getMonth() !== month || date.getDate() !== day) {
    return null;
  }
  return date;
}

function toIsoDate(date: Date): string {
  const year = String(date.getFullYear()).padStart(4, "0");
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function visibleFrom(value: string): YearMonth {
  const date = parseIsoDate(value) ?? new Date();
  return { year: date.getFullYear(), month: date.getMonth() };
}

function shiftMonth({ year, month }: YearMonth, delta: number): YearMonth {
  const date = new Date(year, month + delta, 1);
  return { year: date.getFullYear(), month: date.getMonth() };
}

function monthGrid(year: number, month: number, weekStartsOn: number): Date[] {
  const first = new Date(year, month, 1);
  const offset = (first.getDay() - weekStartsOn + 7) % 7;
  const start = new Date(year, month, 1 - offset);
  return Array.from({ length: 42 }, (_, index) => (
    new Date(start.getFullYear(), start.getMonth(), start.getDate() + index)
  ));
}

function yearRange(visibleYear: number, minDate: Date | null, maxDate: Date | null): number[] {
  const todayYear = new Date().getFullYear();
  const start = Math.min(visibleYear, minDate?.getFullYear() ?? todayYear - YEAR_PAST, todayYear - YEAR_PAST);
  const end = Math.max(visibleYear, maxDate?.getFullYear() ?? todayYear + YEAR_FUTURE, todayYear + YEAR_FUTURE);
  return Array.from({ length: end - start + 1 }, (_, index) => start + index);
}

function isOutOfRange(iso: string, minDate: Date | null, maxDate: Date | null): boolean {
  if (minDate != null && iso < toIsoDate(minDate)) {
    return true;
  }
  if (maxDate != null && iso > toIsoDate(maxDate)) {
    return true;
  }
  return false;
}

function monthBeforeMin(visible: YearMonth, minDate: Date | null): boolean {
  if (minDate == null) {
    return false;
  }
  const previous = shiftMonth(visible, -1);
  return previous.year < minDate.getFullYear()
    || (previous.year === minDate.getFullYear() && previous.month < minDate.getMonth());
}

function monthAfterMax(visible: YearMonth, maxDate: Date | null): boolean {
  if (maxDate == null) {
    return false;
  }
  const next = shiftMonth(visible, 1);
  return next.year > maxDate.getFullYear()
    || (next.year === maxDate.getFullYear() && next.month > maxDate.getMonth());
}

function monthLabels(locale: string | undefined): string[] {
  const format = new Intl.DateTimeFormat(locale, { month: "short" });
  return Array.from({ length: 12 }, (_, month) => format.format(new Date(2020, month, 1)));
}

function weekdayLabels(locale: string | undefined, weekStartsOn: number): string[] {
  const format = new Intl.DateTimeFormat(locale, { weekday: "narrow" });
  return Array.from({ length: 7 }, (_, index) => (
    format.format(new Date(2020, 0, 5 + weekStartsOn + index))
  ));
}

function weekStartDay(locale: string | undefined): number {
  if (locale == null) {
    return 0;
  }
  try {
    const intlLocale = new Intl.Locale(locale);
    const weekInfo = getWeekInfo(intlLocale);
    const firstDay = weekInfo?.firstDay;
    if (firstDay == null) {
      return 0;
    }
    return firstDay === 7 ? 0 : firstDay;
  } catch {
    return 0;
  }
}

function getWeekInfo(locale: Intl.Locale): { firstDay?: number } | undefined {
  const withWeekInfo = locale as Intl.Locale & {
    weekInfo?: { firstDay?: number };
    getWeekInfo?: () => { firstDay?: number };
  };
  if (typeof withWeekInfo.getWeekInfo === "function") {
    return withWeekInfo.getWeekInfo();
  }
  return withWeekInfo.weekInfo;
}
