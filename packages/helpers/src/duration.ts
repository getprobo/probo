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
