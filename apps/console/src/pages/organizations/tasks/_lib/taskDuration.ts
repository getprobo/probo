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

export const taskDurationUnits = ["S", "M", "H", "D", "W"] as const;

export type TaskDurationUnit = (typeof taskDurationUnits)[number];

export type ParsedTaskDuration = {
  amount: number;
  unit: TaskDurationUnit;
};

const secondsPerMinute = 60;
const secondsPerHour = 60 * secondsPerMinute;
const secondsPerDay = 24 * secondsPerHour;
const secondsPerWeek = 7 * secondsPerDay;
const secondsPerYear = 365 * secondsPerDay;
const secondsPerMonth = secondsPerYear / 12;
const secondsPerUnit: Record<TaskDurationUnit, number> = {
  S: 1,
  M: secondsPerMinute,
  H: secondsPerHour,
  D: secondsPerDay,
  W: secondsPerWeek,
};

const maxTaskEstimateSeconds = 1000 * secondsPerHour;

const isoNumber = String.raw`(\d+(?:\.\d+)?)`;
const isoDurationPattern = new RegExp(
  `^P(?:${isoNumber}Y)?(?:${isoNumber}M)?(?:${isoNumber}W)?(?:${isoNumber}D)?`
  + `(?:T(?:${isoNumber}H)?(?:${isoNumber}M)?(?:${isoNumber}S)?)?$`,
);

function isoDurationSeconds(value: string): number | null {
  const negative = value.startsWith("-");
  const match = (negative ? value.slice(1) : value).match(isoDurationPattern);
  if (!match) {
    return null;
  }

  const [, years, months, weeks, days, hours, minutes, seconds] = match;
  if (
    years == null
    && months == null
    && weeks == null
    && days == null
    && hours == null
    && minutes == null
    && seconds == null
  ) {
    return null;
  }

  const total
    = Number(years ?? 0) * secondsPerYear
      + Number(months ?? 0) * secondsPerMonth
      + Number(weeks ?? 0) * secondsPerWeek
      + Number(days ?? 0) * secondsPerDay
      + Number(hours ?? 0) * secondsPerHour
      + Number(minutes ?? 0) * secondsPerMinute
      + Number(seconds ?? 0);

  if (!Number.isFinite(total) || total <= 0) {
    return null;
  }

  return negative ? -total : total;
}

export function parseTaskDurationAmount(value: string): number | null {
  const trimmed = value.trim();
  if (!/^[1-9]\d*$/.test(trimmed)) {
    return null;
  }

  const amount = Number.parseInt(trimmed, 10);
  if (!Number.isSafeInteger(amount)) {
    return null;
  }

  return amount;
}

export function stringifyTaskDuration(
  amount: number,
  unit: TaskDurationUnit,
): string | null {
  if (!Number.isFinite(amount) || amount <= 0) {
    return null;
  }

  const unitSeconds = secondsPerUnit[unit];
  if (amount > Math.floor(maxTaskEstimateSeconds / unitSeconds)) {
    return null;
  }

  switch (unit) {
    case "S":
      return `PT${amount}S`;
    case "M":
      return `PT${amount}M`;
    case "H":
      return `PT${amount}H`;
    case "D":
      return `P${amount}D`;
    case "W":
      return `P${amount * 7}D`;
  }
}

export function parseTaskDuration(value: string): ParsedTaskDuration | null {
  const raw = isoDurationSeconds(value);
  if (raw == null || raw <= 0) {
    return null;
  }

  const seconds = Math.round(raw);
  if (seconds <= 0 || !Number.isSafeInteger(seconds)) {
    return null;
  }

  if (seconds % secondsPerWeek === 0) {
    return { amount: seconds / secondsPerWeek, unit: "W" };
  }
  if (seconds % secondsPerDay === 0) {
    return { amount: seconds / secondsPerDay, unit: "D" };
  }
  if (seconds % secondsPerHour === 0) {
    return { amount: seconds / secondsPerHour, unit: "H" };
  }
  if (seconds % secondsPerMinute === 0) {
    return { amount: seconds / secondsPerMinute, unit: "M" };
  }

  return { amount: seconds, unit: "S" };
}

export function taskDurationsEqual(left: string, right: string): boolean {
  const parsedLeft = parseTaskDuration(left);
  const parsedRight = parseTaskDuration(right);
  return (
    parsedLeft != null
    && parsedRight != null
    && parsedLeft.amount === parsedRight.amount
    && parsedLeft.unit === parsedRight.unit
  );
}
