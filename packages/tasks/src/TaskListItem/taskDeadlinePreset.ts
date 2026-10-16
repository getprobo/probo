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

import { formatDatetime, parseDate, toDateInput, todayAsDateInput } from "@probo/helpers";
import { weekStartDay } from "@probo/ui/src/v2/form/dateModel";

export const deadlinePresets = ["today", "tomorrow", "endOfWeek", "nextWeek"] as const;

export function deadlinePresetDate(
  preset: (typeof deadlinePresets)[number],
  locale: string,
): string {
  return toDateInput(deadlineFromPreset(preset, locale));
}

function dateInputPlusDays(days: number): string {
  const today = parseDate(todayAsDateInput());
  const next = new Date(today.getFullYear(), today.getMonth(), today.getDate() + days);
  const year = String(next.getFullYear()).padStart(4, "0");
  const month = String(next.getMonth() + 1).padStart(2, "0");
  const day = String(next.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

export function deadlineFromPreset(preset: string, locale: string): string | null {
  if (preset === "none") {
    return null;
  }
  if (preset === "endOfWeek") {
    return formatDatetime(endOfWeekDateInput(locale)) ?? null;
  }
  const days = preset === "today" ? 0 : preset === "tomorrow" ? 1 : preset === "nextWeek" ? 7 : null;
  if (days == null) {
    return null;
  }
  return formatDatetime(dateInputPlusDays(days)) ?? null;
}

function endOfWeekDateInput(locale: string): string {
  const today = parseDate(todayAsDateInput());
  const weekEnd = (weekStartDay(locale) + 6) % 7;
  let delta = weekEnd - today.getDay();
  if (delta < 0) {
    delta += 7;
  }
  return dateInputPlusDays(delta);
}
