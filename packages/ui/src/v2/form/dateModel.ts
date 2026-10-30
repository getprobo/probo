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

const YEAR_PAST = 40;
const YEAR_FUTURE = 10;
const ISO_DATE = /^(\d{4})-(\d{2})-(\d{2})$/;

export type DatePart = "day" | "month" | "year";

export type DatePattern = {
  order: DatePart[];
  separator: string;
};

export type YearMonth = {
  year: number;
  month: number;
};

export function datePattern(locale: string | undefined): DatePattern {
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

export function maskPlaceholder({ order, separator }: DatePattern): string {
  const labels: Record<DatePart, string> = {
    day: "dd",
    month: "mm",
    year: "yyyy",
  };
  return order.map(part => labels[part]).join(separator);
}

export function maskDigits(digits: string, { order, separator }: DatePattern): string {
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

export function isoToMasked(iso: string, pattern: DatePattern): string {
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

export function digitsToIso(digits: string, { order }: DatePattern): string | null {
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
  const date = dateFromParts(values.year, values.month - 1, values.day);
  if (date == null) {
    return null;
  }
  return toIsoDate(date);
}

export function parseIsoDate(value: string): Date | null {
  const match = ISO_DATE.exec(value);
  if (match == null) {
    return null;
  }
  const year = Number(match[1]);
  const month = Number(match[2]) - 1;
  const day = Number(match[3]);
  return dateFromParts(year, month, day);
}

function dateFromParts(year: number, month: number, day: number): Date | null {
  const date = new Date(0);
  date.setFullYear(year, month, day);
  if (date.getFullYear() !== year || date.getMonth() !== month || date.getDate() !== day) {
    return null;
  }
  return date;
}

export function toIsoDate(date: Date): string {
  const year = String(date.getFullYear()).padStart(4, "0");
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export type DateDayTone = "day" | "selected" | "within" | "start" | "end" | "single";

// Highlight for one day in a start/end span. `end` may be a hover preview.
// ISO dates sort lexicographically, so an earlier second click swaps the ends.
export function dateRangeTone(iso: string, start: string, end: string): DateDayTone {
  if (start === "" && end === "") {
    return "day";
  }
  if (end === "") {
    return iso === start ? "single" : "day";
  }
  if (start === "") {
    return iso === end ? "single" : "day";
  }
  const [lo, hi] = start <= end ? [start, end] : [end, start];
  if (iso === lo && iso === hi) {
    return "single";
  }
  if (iso === lo) {
    return "start";
  }
  if (iso === hi) {
    return "end";
  }
  if (iso > lo && iso < hi) {
    return "within";
  }
  return "day";
}

export function orderedDateRange(from: string, to: string): { from: string; to: string } {
  if (from !== "" && to !== "" && from > to) {
    return { from: to, to: from };
  }
  return { from, to };
}

const RANGE_DASH = "–";

function formatIsoDay(iso: string, locale: string | undefined, withYear: boolean): string {
  const date = parseIsoDate(iso);
  if (date == null) {
    return "";
  }
  return new Intl.DateTimeFormat(locale, withYear
    ? { day: "numeric", month: "short", year: "numeric" }
    : { day: "numeric", month: "short" }).format(date);
}

// Empty when both ends are blank. A missing end keeps the dash so a
// half-chosen span still reads as a range.
export function formatDateRange(from: string, to: string, locale: string | undefined): string {
  if (from !== "" && to !== "") {
    const start = parseIsoDate(from);
    const end = parseIsoDate(to);
    const sameYear = start != null && end != null && start.getFullYear() === end.getFullYear();
    return `${formatIsoDay(from, locale, !sameYear)} ${RANGE_DASH} ${formatIsoDay(to, locale, true)}`;
  }
  if (from !== "") {
    return `${formatIsoDay(from, locale, true)} ${RANGE_DASH}`;
  }
  if (to !== "") {
    return `${RANGE_DASH} ${formatIsoDay(to, locale, true)}`;
  }
  return "";
}

export function visibleFrom(value: string): YearMonth {
  const date = parseIsoDate(value) ?? new Date();
  return { year: date.getFullYear(), month: date.getMonth() };
}

export function shiftMonth({ year, month }: YearMonth, delta: number): YearMonth {
  const date = new Date(year, month + delta, 1);
  return { year: date.getFullYear(), month: date.getMonth() };
}

export function monthGrid(year: number, month: number, weekStartsOn: number): Date[] {
  const first = new Date(year, month, 1);
  const offset = (first.getDay() - weekStartsOn + 7) % 7;
  const start = new Date(year, month, 1 - offset);
  return Array.from({ length: 42 }, (_, index) => (
    new Date(start.getFullYear(), start.getMonth(), start.getDate() + index)
  ));
}

export function yearRange(
  visibleYear: number,
  minDate: Date | null,
  maxDate: Date | null,
): number[] {
  const todayYear = new Date().getFullYear();
  const start = Math.min(
    visibleYear,
    minDate?.getFullYear() ?? todayYear - YEAR_PAST,
    todayYear - YEAR_PAST,
  );
  const end = Math.max(
    visibleYear,
    maxDate?.getFullYear() ?? todayYear + YEAR_FUTURE,
    todayYear + YEAR_FUTURE,
  );
  return Array.from({ length: end - start + 1 }, (_, index) => start + index);
}

export function isOutOfRange(iso: string, minDate: Date | null, maxDate: Date | null): boolean {
  if (minDate != null && iso < toIsoDate(minDate)) {
    return true;
  }
  if (maxDate != null && iso > toIsoDate(maxDate)) {
    return true;
  }
  return false;
}

export function monthBeforeMin(visible: YearMonth, minDate: Date | null): boolean {
  if (minDate == null) {
    return false;
  }
  const previous = shiftMonth(visible, -1);
  return previous.year < minDate.getFullYear()
    || (previous.year === minDate.getFullYear() && previous.month < minDate.getMonth());
}

export function monthAfterMax(visible: YearMonth, maxDate: Date | null): boolean {
  if (maxDate == null) {
    return false;
  }
  const next = shiftMonth(visible, 1);
  return next.year > maxDate.getFullYear()
    || (next.year === maxDate.getFullYear() && next.month > maxDate.getMonth());
}

export function monthLabels(locale: string | undefined): string[] {
  const format = new Intl.DateTimeFormat(locale, { month: "short" });
  return Array.from({ length: 12 }, (_, month) => format.format(new Date(2020, month, 1)));
}

export function weekdayLabels(locale: string | undefined, weekStartsOn: number): string[] {
  const format = new Intl.DateTimeFormat(locale, { weekday: "narrow" });
  return Array.from({ length: 7 }, (_, index) => (
    format.format(new Date(2020, 0, 5 + weekStartsOn + index))
  ));
}

export function weekStartDay(locale: string | undefined): number {
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
