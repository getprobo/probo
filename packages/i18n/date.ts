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
