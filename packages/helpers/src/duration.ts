// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

const UNITS: [number, string, string][] = [
  [365 * 24 * 3600, "year", "years"],
  [30 * 24 * 3600, "month", "months"],
  [7 * 24 * 3600, "week", "weeks"],
  [24 * 3600, "day", "days"],
  [3600, "hour", "hours"],
  [60, "minute", "minutes"],
];

export const DURATION_UNITS: { value: string; label: string; seconds: number }[] = [
  { value: "seconds", label: "seconds", seconds: 1 },
  { value: "minutes", label: "minutes", seconds: 60 },
  { value: "hours", label: "hours", seconds: 3600 },
  { value: "days", label: "days", seconds: 86400 },
  { value: "weeks", label: "weeks", seconds: 604800 },
  { value: "months", label: "months", seconds: 2592000 },
  { value: "years", label: "years", seconds: 31536000 },
];

export function humanizeSeconds(seconds: number | null): string {
  if (seconds === null || seconds <= 0) return "session";
  for (const [unit, singular, plural] of UNITS) {
    if (seconds >= unit && seconds % unit === 0) {
      const count = seconds / unit;
      return `${count} ${count === 1 ? singular : plural}`;
    }
  }
  return `${seconds} ${seconds === 1 ? "second" : "seconds"}`;
}

export function toMaxAgeSeconds(value: string, unit: string): number | null {
  const trimmed = value.trim();
  if (trimmed === "" || !/^\d+(\.\d+)?$/.test(trimmed)) return null;
  const num = Number(trimmed);
  if (!Number.isFinite(num) || num <= 0) return null;
  const u = DURATION_UNITS.find(u => u.value === unit);
  if (!u) return null;
  const rounded = Math.round(num * u.seconds);
  if (rounded <= 0) return null;
  return rounded;
}

export function fromMaxAgeSeconds(seconds: number | null): { value: string; unit: string } {
  if (seconds === null || seconds <= 0) return { value: "", unit: "days" };
  for (const u of [...DURATION_UNITS].reverse()) {
    if (seconds >= u.seconds && seconds % u.seconds === 0) {
      return { value: String(seconds / u.seconds), unit: u.value };
    }
  }
  return { value: String(seconds), unit: "seconds" };
}
