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

import { todayAsDateInput } from "@probo/helpers";

export const matrixAsOfParam = "asOf";

const dateInputPattern = /^\d{4}-\d{2}-\d{2}$/;

function isCalendarDateInput(value: string): boolean {
  if (!dateInputPattern.test(value)) {
    return false;
  }

  const year = Number(value.slice(0, 4));
  const month = Number(value.slice(5, 7));
  const day = Number(value.slice(8, 10));
  const parsed = new Date(Date.UTC(year, month - 1, day));

  return (
    parsed.getUTCFullYear() === year
    && parsed.getUTCMonth() === month - 1
    && parsed.getUTCDate() === day
  );
}

function localDateInput(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");

  return `${year}-${month}-${day}`;
}

function createdOnDateInput(createdAt: string): string {
  const parsed = new Date(createdAt);
  if (Number.isNaN(parsed.getTime())) {
    return todayAsDateInput();
  }

  return localDateInput(parsed);
}

export function parseAsOfDate(params: URLSearchParams): string {
  const today = todayAsDateInput();
  const value = params.get(matrixAsOfParam);
  if (!value || !isCalendarDateInput(value) || value >= today) {
    return today;
  }

  return value;
}

export function clampDateInput(dateInput: string, min: string, max: string): string {
  if (!dateInput || dateInput > max) {
    return max;
  }
  if (dateInput < min) {
    return min;
  }
  return dateInput;
}

export function asOfDateBounds(createdAt: string): { minDate: string; maxDate: string } {
  const createdDate = createdOnDateInput(createdAt);
  const maxDate = todayAsDateInput();
  const minDate = createdDate > maxDate ? maxDate : createdDate;

  return { minDate, maxDate };
}

export function matrixAsOf(dateInput: string): string | null {
  const date = dateInput || todayAsDateInput();
  if (date === todayAsDateInput()) {
    return null;
  }

  const year = Number(date.slice(0, 4));
  const month = Number(date.slice(5, 7));
  const day = Number(date.slice(8, 10));

  return new Date(Date.UTC(year, month - 1, day + 1)).toISOString();
}
