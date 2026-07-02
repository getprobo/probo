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

export const DURATION_UNITS: { value: string; label: string; singular: string; seconds: number, snap: number }[] = [
  { value: "seconds", label: "seconds", singular: "second", seconds: 1, snap: 0 },
  { value: "minutes", label: "minutes", singular: "minute", seconds: 60, snap: 5 },
  { value: "hours", label: "hours", singular: "hour", seconds: 3600, snap: 5 * 60 },
  { value: "days", label: "days", singular: "day", seconds: 86400, snap: 2 * 3600 },
  { value: "weeks", label: "weeks", singular: "week", seconds: 604800, snap: 12 * 3600 },
  { value: "months", label: "months", singular: "month", seconds: 2592000, snap: 2 * 24 *  3600 },
  { value: "years", label: "years", singular: "year", seconds: 31536000, snap: 21 * 24 *  3600 },
] as const;

// Tracker types whose data persists until explicitly cleared. When such a
// tracker has no max-age, its lifetime is "persistent" rather than "session"
// (cookies and session storage are cleared when the session/tab ends).
const PERSISTENT_TRACKER_TYPES = new Set([
  "LOCAL_STORAGE",
  "INDEXED_DB",
  "CACHE_STORAGE",
]);

export function humanizeSeconds(
  seconds: number | null,
  trackerType?: string | null,
): string {
  if (seconds === null || seconds <= 0) {
    return trackerType && PERSISTENT_TRACKER_TYPES.has(trackerType)
      ? "persistent"
      : "session";
  }

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
