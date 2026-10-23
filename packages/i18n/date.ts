// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { parseDate } from "@probo/helpers";

const relativeFormat = [
  { limit: 1000 * 60 * 60 * 24 * 365, unit: "years" },
  { limit: 1000 * 60 * 60 * 24 * 30, unit: "months" },
  { limit: 1000 * 60 * 60 * 24 * 7, unit: "weeks" },
  { limit: 1000 * 60 * 60 * 24, unit: "days" },
  { limit: 1000 * 60 * 60, unit: "hours" },
  { limit: 1000 * 60, unit: "minutes" },
  { limit: 1000, unit: "seconds" },
] as const;

export function relativeDateFormat(
  language: string,
  date: Date | string | null | undefined,
  options: Intl.RelativeTimeFormatOptions = { style: "long" },
): string {
  if (!date) return "";

  const dateValue = typeof date === "string" ? parseDate(date) : date;
  const distanceInMilliseconds = dateValue.getTime() - Date.now();
  const formatter = new Intl.RelativeTimeFormat(language, options);

  for (const { limit, unit } of relativeFormat) {
    if (Math.abs(distanceInMilliseconds) >= limit) {
      return formatter.format(Math.round(distanceInMilliseconds / limit), unit);
    }
  }

  return "";
}

export function dateFormat(
  language: string,
  date: Date | string | null | undefined,
  options: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "short",
    day: "numeric",
    weekday: "short",
  },
): string {
  if (!date) return "";

  const dateValue = typeof date === "string" ? parseDate(date) : date;
  return new Intl.DateTimeFormat(language, options).format(dateValue);
}

export function dateTimeFormat(
  language: string,
  date: Date | string | null | undefined,
  options: Intl.DateTimeFormatOptions = {
    hour: "2-digit",
    hour12: false,
    minute: "2-digit",
    day: "numeric",
    month: "short",
    year: "numeric",
  },
): string {
  return dateFormat(language, date, options);
}

function startOfLocalDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

function calendarDayDiff(date: Date, now = new Date()): number {
  const msPerDay = 24 * 60 * 60 * 1000;
  return Math.round(
    (startOfLocalDay(date).getTime() - startOfLocalDay(now).getTime()) / msPerDay,
  );
}

export type CompactDateFormatOptions = {
  // Another year reads "Sep 2027" instead of "Sep 30, 2027".
  omitDayWithYear?: boolean;
};

export function compactDateFormat(
  language: string,
  date: Date | string | null | undefined,
  now = new Date(),
  format: CompactDateFormatOptions = {},
): string {
  if (!date) return "";

  const dateValue = typeof date === "string" ? parseDate(date) : date;
  const sameYear = dateValue.getFullYear() === now.getFullYear();
  const options: Intl.DateTimeFormatOptions = { month: "short" };
  if (sameYear || format.omitDayWithYear !== true) {
    options.day = "numeric";
  }
  if (!sameYear) {
    options.year = "numeric";
  }

  return dateFormat(language, dateValue, options);
}

function asCalendarDate(date: Date | string): Date {
  if (typeof date !== "string") {
    return date;
  }

  const match = /^(\d{4})-(\d{2})-(\d{2})/.exec(date);
  if (match == null) {
    return parseDate(date);
  }

  return new Date(Number(match[1]), Number(match[2]) - 1, Number(match[3]));
}

export function deadlineDateFormat(
  language: string,
  date: Date | string | null | undefined,
  now = new Date(),
): string {
  if (!date) return "";

  const dateValue = asCalendarDate(date);
  const diff = calendarDayDiff(dateValue, now);
  if (diff >= -1 && diff <= 1) {
    return new Intl.RelativeTimeFormat(language, { numeric: "auto" }).format(diff, "day");
  }

  return compactDateFormat(language, dateValue, now);
}
