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

export const taskDurationUnits = ["S", "M", "H", "D", "W", "MO", "Y"] as const;

export type TaskDurationUnit = (typeof taskDurationUnits)[number];

export const taskEstimateDurationUnits = ["S", "M", "H", "D", "W"] as const;

export const taskRecurrenceDurationUnits = ["S", "M", "H", "D", "W", "MO", "Y"] as const;

export type ParsedTaskDuration = {
  amount: number;
  unit: TaskDurationUnit;
};

const secondsPerMinute = 60;
const secondsPerHour = 60 * secondsPerMinute;
const secondsPerDay = 24 * secondsPerHour;
const secondsPerWeek = 7 * secondsPerDay;

const isoNumber = String.raw`(\d+(?:\.\d+)?)`;
const isoDurationPattern = new RegExp(
  `^P(?:${isoNumber}Y)?(?:${isoNumber}M)?(?:${isoNumber}W)?(?:${isoNumber}D)?`
  + `(?:T(?:${isoNumber}H)?(?:${isoNumber}M)?(?:${isoNumber}S)?)?$`,
);

type IsoDurationComponents = {
  years: number;
  months: number;
  weeks: number;
  days: number;
  hours: number;
  minutes: number;
  seconds: number;
};

function parseIsoDuration(value: string): IsoDurationComponents | null {
  const negative = value.startsWith("-");
  const match = (negative ? value.slice(1) : value).match(isoDurationPattern);
  if (!match || negative) {
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

  return {
    years: Number(years ?? 0),
    months: Number(months ?? 0),
    weeks: Number(weeks ?? 0),
    days: Number(days ?? 0),
    hours: Number(hours ?? 0),
    minutes: Number(minutes ?? 0),
    seconds: Number(seconds ?? 0),
  };
}

function isWholePositive(value: number): boolean {
  return Number.isSafeInteger(value) && value > 0;
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
    case "MO":
      return `P${amount}M`;
    case "Y":
      return `P${amount}Y`;
  }
}

export function parseTaskDuration(value: string): ParsedTaskDuration | null {
  const components = parseIsoDuration(value);
  if (components == null) {
    return null;
  }

  const { years, months, weeks, days, hours, minutes, seconds } = components;
  const hasClock = hours > 0 || minutes > 0 || seconds > 0;

  if (years > 0 || months > 0) {
    if (weeks > 0 || days > 0 || hasClock) {
      return null;
    }

    const totalMonths = years * 12 + months;
    if (!isWholePositive(totalMonths)) {
      return null;
    }

    if (totalMonths % 12 === 0) {
      return { amount: totalMonths / 12, unit: "Y" };
    }

    return { amount: totalMonths, unit: "MO" };
  }

  const totalSeconds
    = weeks * secondsPerWeek
      + days * secondsPerDay
      + hours * secondsPerHour
      + minutes * secondsPerMinute
      + seconds;
  const normalized = Math.round(totalSeconds);
  if (!Number.isFinite(totalSeconds) || !isWholePositive(normalized)) {
    return null;
  }

  if (normalized % secondsPerWeek === 0) {
    return { amount: normalized / secondsPerWeek, unit: "W" };
  }

  if (normalized % secondsPerDay === 0) {
    return { amount: normalized / secondsPerDay, unit: "D" };
  }

  if (normalized % secondsPerHour === 0) {
    return { amount: normalized / secondsPerHour, unit: "H" };
  }

  if (normalized % secondsPerMinute === 0) {
    return { amount: normalized / secondsPerMinute, unit: "M" };
  }

  return { amount: normalized, unit: "S" };
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
