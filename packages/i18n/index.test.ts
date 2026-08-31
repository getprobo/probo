// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { describe, expect, it } from "vitest";

import { fileSize, formatDuration, humanizeSeconds } from "./index";

describe("fileSize", () => {
  it("formats byte counts with translated units", () => {
    const t = (key: string) => key.replace("size.", "");

    expect(fileSize(4911, t)).toBe("4.8 KB");
    expect(fileSize(20, t)).toBe("20 B");
  });
});

describe("humanizeSeconds", () => {
  const t = (key: string, options?: { count?: number }) => {
    const unit = key.replace("duration.", "");
    if (unit === "session" || unit === "persistent") return unit;
    return options?.count === 1 ? unit.slice(0, -1) : unit;
  };

  it("formats durations with translated plural units", () => {
    expect(humanizeSeconds(4570, t)).toBe("1 hour, 16 minutes, 10 seconds");
    expect(humanizeSeconds(120, t)).toBe("2 minutes");
  });

  it("formats trackers without a max age as session or persistent", () => {
    expect(humanizeSeconds(null, t)).toBe("session");
    expect(humanizeSeconds(0, t, "LOCAL_STORAGE")).toBe("persistent");
  });
});

describe("formatDuration", () => {
  const countedUnits: Record<string, { one: string; other: string }> = {
    "duration.sec": { one: "{{count}} Second", other: "{{count}} Seconds" },
    "duration.min": { one: "{{count}} Minute", other: "{{count}} Minutes" },
    "duration.hour": { one: "{{count}} Hour", other: "{{count}} Hours" },
  };
  const unitWords: Record<string, { one: string; other: string }> = {
    "duration.years": { one: "year", other: "years" },
    "duration.months": { one: "month", other: "months" },
    "duration.weeks": { one: "week", other: "weeks" },
    "duration.days": { one: "day", other: "days" },
  };

  function t(key: string, options?: { count?: number }) {
    const count = options?.count;
    const form = count === 1 ? "one" : "other";
    const counted = countedUnits[key];
    if (counted && count != null) {
      return counted[form].replaceAll("{{count}}", String(count));
    }
    const unit = unitWords[key];
    if (unit && count != null) {
      return unit[form];
    }
    return key;
  }

  it("formats a single ISO time or date component", () => {
    expect(formatDuration("PT30S", t)).toBe("30 Seconds");
    expect(formatDuration("PT5M", t)).toBe("5 Minutes");
    expect(formatDuration("PT1H", t)).toBe("1 Hour");
    expect(formatDuration("P2D", t)).toBe("2 days");
    expect(formatDuration("P1W", t)).toBe("1 week");
    expect(formatDuration("P7D", t)).toBe("1 week");
    expect(formatDuration("P6M", t)).toBe("6 months");
    expect(formatDuration("P1Y", t)).toBe("1 year");
  });

  it("keeps every component of a composite ISO duration", () => {
    expect(formatDuration("PT1H30M", t)).toBe("1 Hour, 30 Minutes");
    expect(formatDuration("PT1M30S", t)).toBe("1 Minute, 30 Seconds");
    expect(formatDuration("P1DT2H", t)).toBe("1 day, 2 Hours");
    expect(formatDuration("PT1H30M5S", t)).toBe("1 Hour, 30 Minutes, 5 Seconds");
    expect(formatDuration("P1Y6M", t)).toBe("1 year, 6 months");
    expect(formatDuration("P1YT2H", t)).toBe("1 year, 2 Hours");
  });
});
