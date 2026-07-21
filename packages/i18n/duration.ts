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
): string {
  if (seconds === null || seconds <= 0) {
    return '';
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

export function formatDuration(
  duration?: string | null,
  t?: Translator,
): string | null {
  if (!duration || !t) return null;

  const timeMatch = duration.match(/PT(\d+)([MH])/);
  if (timeMatch) {
    const amount = parseInt(timeMatch[1], 10) || 0;
    const unit = timeMatch[2];
    if (unit === "M") return t("duration.min", { count: amount });
    if (unit === "H") return t("duration.hour", { count: amount });
  }

  const dateMatch = duration.match(/P(\d+)([DW])/);
  if (dateMatch) {
    const amount = parseInt(dateMatch[1], 10) || 0;
    const unit = dateMatch[2];
    if (unit === "W") {
      return `${amount} ${amount === 1 ? t("Week") : t("Weeks")}`;
    }
    if (unit === "D") {
      if (amount % 7 === 0 && amount > 0) {
        const weeks = amount / 7;
        return `${weeks} ${weeks === 1 ? t("Week") : t("Weeks")}`;
      }
      return `${amount} ${amount === 1 ? t("Day") : t("Days")}`;
    }
  }

  return null;
}
