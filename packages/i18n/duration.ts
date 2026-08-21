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

type Translator = (key: string, options?: { count?: number }) => string;

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

function parseAmount(raw: string | undefined): number {
  if (raw == null) {
    return 0;
  }

  const amount = Number(raw);
  return Number.isFinite(amount) ? amount : 0;
}

function parseIsoDuration(value: string): IsoDurationComponents | null {
  const match = (value.startsWith("-") ? value.slice(1) : value).match(
    isoDurationPattern,
  );
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

  return {
    years: parseAmount(years),
    months: parseAmount(months),
    weeks: parseAmount(weeks),
    days: parseAmount(days),
    hours: parseAmount(hours),
    minutes: parseAmount(minutes),
    seconds: parseAmount(seconds),
  };
}

const DURATION_UNITS = [
  { value: "seconds", seconds: 1, snap: 0 },
  { value: "minutes", seconds: 60, snap: 5 },
  { value: "hours", seconds: 3_600, snap: 5 * 60 },
  { value: "days", seconds: 86_400, snap: 2 * 3_600 },
  { value: "weeks", seconds: 604_800, snap: 12 * 3_600 },
  { value: "months", seconds: 2_592_000, snap: 2 * 24 * 3_600 },
  { value: "years", seconds: 31_536_000, snap: 21 * 24 * 3_600 },
] as const;

export function humanizeSeconds(
  seconds: number | null,
  t: Translator,
  storageType?: string,
): string {
  if (seconds === null || seconds <= 0) {
    return storageType === "LOCAL_STORAGE"
      ? t("duration.persistent")
      : t("duration.session");
  }

  let remaining = seconds;
  const parts: string[] = [];

  for (const { value, seconds: durationInSeconds, snap } of [...DURATION_UNITS].reverse()) {
    if (remaining >= durationInSeconds - snap) {
      let count = Math.floor(remaining / durationInSeconds);
      const leftover = remaining - count * durationInSeconds;

      if (leftover >= durationInSeconds - snap) {
        count++;
        remaining = 0;
      } else if (leftover <= snap) {
        remaining = 0;
      } else {
        remaining = leftover;
      }

      parts.push(`${count} ${t(`duration.${value}`, { count })}`);
    }
  }

  return parts.length > 0 ? parts.join(", ") : t("duration.session");
}

export type FormatDurationOptions = {
  style?: "long" | "short";
  locale?: string;
};

type DurationFormatCtor = new (
  locales?: string | string[],
  options?: { style?: "long" | "short" | "narrow" },
) => { format: (value: Record<string, number>) => string };

function formatDurationShortIntl(
  locale: string,
  components: IsoDurationComponents,
): string | null {
  const DurationFormat = (Intl as typeof Intl & {
    DurationFormat?: DurationFormatCtor;
  }).DurationFormat;
  if (DurationFormat == null) {
    return null;
  }

  const value: Record<string, number> = {};
  for (const [unit, amount] of Object.entries(components)) {
    if (amount > 0) {
      value[unit] = amount;
    }
  }
  if (Object.keys(value).length === 0) {
    return null;
  }

  try {
    return new DurationFormat(locale, { style: "short" }).format(value);
  } catch {
    return null;
  }
}

function normalizeDuration(components: IsoDurationComponents): IsoDurationComponents {
  let { weeks, days } = components;
  const { years, months, hours, minutes, seconds } = components;
  if (
    weeks === 0
    && days > 0
    && days % 7 === 0
    && hours === 0
    && minutes === 0
    && seconds === 0
    && years === 0
    && months === 0
  ) {
    weeks = days / 7;
    days = 0;
  }

  return { years, months, weeks, days, hours, minutes, seconds };
}

export function formatDuration(
  duration?: string | null,
  t?: Translator,
  options?: FormatDurationOptions,
): string | null {
  if (!duration) return null;

  const style = options?.style ?? "long";
  if (style === "long" && !t) return null;

  const parsed = parseIsoDuration(duration);
  if (!parsed) {
    return null;
  }

  const components = normalizeDuration(parsed);
  if (style === "short") {
    if (options?.locale != null) {
      const intl = formatDurationShortIntl(options.locale, components);
      if (intl != null) {
        return intl;
      }
    }
    if (t == null) {
      return null;
    }

    const parts: string[] = [];
    if (components.years > 0) {
      parts.push(t("duration.yearShort", { count: components.years }));
    }
    if (components.months > 0) {
      parts.push(t("duration.monthShort", { count: components.months }));
    }
    if (components.weeks > 0) {
      parts.push(t("duration.weekShort", { count: components.weeks }));
    }
    if (components.days > 0) {
      parts.push(t("duration.dayShort", { count: components.days }));
    }
    if (components.hours > 0) {
      parts.push(t("duration.hourShort", { count: components.hours }));
    }
    if (components.minutes > 0) {
      parts.push(t("duration.minShort", { count: components.minutes }));
    }
    if (components.seconds > 0) {
      parts.push(t("duration.secShort", { count: components.seconds }));
    }

    return parts.length > 0 ? parts.join(", ") : null;
  }

  if (t == null) {
    return null;
  }

  const parts: string[] = [];
  if (components.years > 0) {
    parts.push(`${components.years} ${t("duration.years", { count: components.years })}`);
  }
  if (components.months > 0) {
    parts.push(`${components.months} ${t("duration.months", { count: components.months })}`);
  }
  if (components.weeks > 0) {
    parts.push(`${components.weeks} ${t("duration.weeks", { count: components.weeks })}`);
  }
  if (components.days > 0) {
    parts.push(`${components.days} ${t("duration.days", { count: components.days })}`);
  }
  if (components.hours > 0) {
    parts.push(t("duration.hour", { count: components.hours }));
  }
  if (components.minutes > 0) {
    parts.push(t("duration.min", { count: components.minutes }));
  }
  if (components.seconds > 0) {
    parts.push(t("duration.sec", { count: components.seconds }));
  }

  return parts.length > 0 ? parts.join(", ") : null;
}
