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

import { dateTimeFormat } from "@probo/i18n";

const scanTimeWithSeconds = {
  hour: "2-digit",
  hour12: false,
  minute: "2-digit",
  second: "2-digit",
  day: "numeric",
  month: "short",
  year: "numeric",
} as const;

export function formatScanTime(
  language: string,
  value: string | null | undefined,
  withSeconds = false,
): string {
  const formatted = withSeconds
    ? dateTimeFormat(language, value, scanTimeWithSeconds)
    : dateTimeFormat(language, value);
  return formatted === "" ? "—" : formatted;
}

export function formatElapsedSeconds(language: string, elapsedMs: number): string {
  return new Intl.NumberFormat(language, { maximumFractionDigits: 1 }).format(elapsedMs / 1000);
}

const EMPTY_DURATION = "—";

type DurationFormatCtor = new (
  locales?: string | string[],
  options?: { style?: "long" | "short" | "narrow" },
) => { format: (value: Record<string, number>) => string };

export function formatScanDuration(language: string, milliseconds: number | null): string {
  if (milliseconds == null || !Number.isFinite(milliseconds)) {
    return EMPTY_DURATION;
  }

  const totalSeconds = Math.max(0, Math.round(milliseconds / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  const DurationFormat = (Intl as typeof Intl & { DurationFormat?: DurationFormatCtor }).DurationFormat;
  if (DurationFormat == null) {
    return `${totalSeconds}s`;
  }

  const value: Record<string, number> = {};
  if (hours > 0) {
    value.hours = hours;
  }
  if (minutes > 0) {
    value.minutes = minutes;
  }
  if (seconds > 0 || (hours === 0 && minutes === 0)) {
    value.seconds = seconds;
  }

  try {
    return new DurationFormat(language, { style: "short" }).format(value);
  } catch {
    return `${totalSeconds}s`;
  }
}

function timestampMs(value: string): number | null {
  const parsed = Date.parse(value);
  return Number.isFinite(parsed) ? parsed : null;
}

export function scanLifecycleDurations(scan: {
  createdAt: string;
  startedAt: string | null;
  completedAt: string | null;
}, now = Date.now()): { totalMs: number; queueMs: number | null; runningMs: number | null } {
  const createdMs = timestampMs(scan.createdAt);
  const startedMs = scan.startedAt == null ? null : timestampMs(scan.startedAt);
  const completedMs = scan.completedAt == null ? null : timestampMs(scan.completedAt);
  if (createdMs == null) {
    return { totalMs: 0, queueMs: null, runningMs: null };
  }

  const endMs = completedMs ?? now;
  const totalMs = Math.max(0, endMs - createdMs);
  if (startedMs == null) {
    return { totalMs, queueMs: totalMs, runningMs: null };
  }

  return {
    totalMs,
    queueMs: Math.max(0, startedMs - createdMs),
    runningMs: Math.max(0, endMs - startedMs),
  };
}
