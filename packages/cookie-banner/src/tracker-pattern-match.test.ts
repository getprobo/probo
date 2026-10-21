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

import { describe, expect, it } from "vitest";

import { selectTrackerCategory } from "./tracker-pattern-match";
import type { Category, TrackerType } from "./types";

function category(
  slug: string,
  names: string[],
  trackerType: TrackerType = "COOKIE",
): Category {
  return {
    name: slug,
    slug,
    description: "",
    kind: "NORMAL",
    cookies: names.map(name => ({
      name,
      tracker_type: trackerType,
      max_age_seconds: null,
      description: "",
    })),
    gcm_consent_types: [],
    posthog_consent: false,
  };
}

describe("selectTrackerCategory", () => {
  it("prefers an exact name over a glob", () => {
    const selected = selectTrackerCategory(
      [
        category("analytics", ["_ga_*"]),
        category("session", ["_ga_123"]),
      ],
      "cookie",
      "_ga_123",
    );

    expect(selected?.slug).toBe("session");
  });

  it("matches a prefix glob", () => {
    const selected = selectTrackerCategory(
      [category("analytics", ["_ga_*"])],
      "cookie",
      "_ga_123",
    );

    expect(selected?.slug).toBe("analytics");
  });

  it("matches a suffix glob", () => {
    const selected = selectTrackerCategory(
      [category("analytics", ["*_posthog"])],
      "cookie",
      "ph_phc_abc_posthog",
    );

    expect(selected?.slug).toBe("analytics");
  });

  it("prefers the glob with the longer fixed text", () => {
    const selected = selectTrackerCategory(
      [category("broad", ["_ga_*"]), category("narrow", ["_ga_1*"])],
      "cookie",
      "_ga_123",
    );

    expect(selected?.slug).toBe("narrow");
  });

  it("matches a storage pattern and ignores a cookie with the same name", () => {
    const selected = selectTrackerCategory(
      [
        category("cookie", ["_ga_*"], "COOKIE"),
        category("storage", ["_ga_*"], "LOCAL_STORAGE"),
      ],
      "storage",
      "_ga_123",
    );

    expect(selected?.slug).toBe("storage");
  });
});
