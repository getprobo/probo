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

export const DURATION_UNITS: { value: string; label: string; singular: string; seconds: number, snap: number }[] = [
  { value: "seconds", label: "seconds", singular: "second", seconds: 1, snap: 0 },
  { value: "minutes", label: "minutes", singular: "minute", seconds: 60, snap: 5 },
  { value: "hours", label: "hours", singular: "hour", seconds: 3600, snap: 5 * 60 },
  { value: "days", label: "days", singular: "day", seconds: 86400, snap: 2 * 3600 },
  { value: "weeks", label: "weeks", singular: "week", seconds: 604800, snap: 12 * 3600 },
  { value: "months", label: "months", singular: "month", seconds: 2592000, snap: 2 * 24 *  3600 },
  { value: "years", label: "years", singular: "year", seconds: 31536000, snap: 21 * 24 *  3600 },
] as const;

export function humanizeSeconds(seconds: number | null): string {
  if (seconds === null || seconds <= 0) return "session";

  let remaining = seconds;
  const parts: string[] = [];

  for (const {label, singular, seconds: durationInSeconds, snap} of [...DURATION_UNITS].reverse()) {
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

      parts.push(`${count} ${count === 1 ? singular : label}`);
    }
  }

  return parts.length > 0 ? parts.join(", ") : "session";
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
