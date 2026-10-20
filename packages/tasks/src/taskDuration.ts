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
    if (weeks > 0 || hasClock) {
      return null;
    }

    const totalMonths = years * 12 + months + days / 30;
    if (!Number.isFinite(totalMonths) || totalMonths <= 0) {
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
  if (!Number.isFinite(totalSeconds) || totalSeconds <= 0) {
    return null;
  }

  if (Number.isInteger(totalSeconds) && totalSeconds % secondsPerWeek === 0) {
    return { amount: totalSeconds / secondsPerWeek, unit: "W" };
  }

  if (Number.isInteger(totalSeconds) && totalSeconds % secondsPerDay === 0) {
    return { amount: totalSeconds / secondsPerDay, unit: "D" };
  }

  if (Number.isInteger(totalSeconds) && totalSeconds % secondsPerHour === 0) {
    return { amount: totalSeconds / secondsPerHour, unit: "H" };
  }

  if (Number.isInteger(totalSeconds) && totalSeconds % secondsPerMinute === 0) {
    return { amount: totalSeconds / secondsPerMinute, unit: "M" };
  }

  return { amount: totalSeconds, unit: "S" };
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
