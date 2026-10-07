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

export const SCIM_EVENT_PAGE_SIZE = 20;
export const SCIM_EXPORT_DAY_MS = 24 * 60 * 60 * 1000;

const MAX_EXPORT_RANGE_YEARS = 1;

export function resultColor(statusCode: number): "green" | "red" | "amber" {
  if (statusCode >= 200 && statusCode < 300) {
    return "green";
  }
  if (statusCode >= 400) {
    return "red";
  }
  return "amber";
}

export function maxExportToDate(fromDate: string): string {
  return addUTCDays(addUTCYears(fromDate, MAX_EXPORT_RANGE_YEARS), -1);
}

export function isExportRangeTooLarge(fromDate: string, toDate: string): boolean {
  return fromDate !== "" && toDate !== "" && toDate > maxExportToDate(fromDate);
}

function parseISODate(isoDate: string): Date {
  const [year, month, day] = isoDate.split("-").map(Number);
  return new Date(Date.UTC(year, month - 1, day));
}

function formatISODate(date: Date): string {
  return date.toISOString().slice(0, 10);
}

function addUTCYears(isoDate: string, years: number): string {
  const date = parseISODate(isoDate);
  date.setUTCFullYear(date.getUTCFullYear() + years);
  return formatISODate(date);
}

function addUTCDays(isoDate: string, days: number): string {
  const date = parseISODate(isoDate);
  date.setUTCDate(date.getUTCDate() + days);
  return formatISODate(date);
}
