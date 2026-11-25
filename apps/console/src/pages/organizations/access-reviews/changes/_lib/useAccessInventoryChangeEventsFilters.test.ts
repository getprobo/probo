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

import { afterAll, beforeAll, describe, expect, it } from "vitest";

import {
  accessInventoryChangeEventLocalDayEnd,
  accessInventoryChangeEventLocalDayStart,
  accessInventoryChangeEventsGraphqlFilter,
} from "./useAccessInventoryChangeEventsFilters";

describe("access inventory change event local day bounds", () => {
  // UTC midnight and local midnight are the same instant when the
  // process zone is UTC, which is how CI runs. Pin a non-UTC zone so a
  // revert to `T00:00:00Z` fails here.
  const previousTz = process.env.TZ;

  beforeAll(() => {
    process.env.TZ = "Europe/Paris";
  });

  afterAll(() => {
    if (previousTz === undefined) {
      delete process.env.TZ;
    } else {
      process.env.TZ = previousTz;
    }
  });

  it("starts at local midnight of the chosen day", () => {
    const start = accessInventoryChangeEventLocalDayStart("2026-03-10");

    expect(start).toBe(new Date(2026, 2, 10, 0, 0, 0, 0).toISOString());
    expect(start).not.toBe("2026-03-10T00:00:00.000Z");
  });

  it("ends at local end of the chosen day", () => {
    const end = accessInventoryChangeEventLocalDayEnd("2026-03-10");

    expect(end).toBe(new Date(2026, 2, 10, 23, 59, 59, 999).toISOString());
    expect(end).not.toBe("2026-03-10T23:59:59.999Z");
  });

  it("rejects a calendar-invalid day", () => {
    expect(accessInventoryChangeEventLocalDayStart("2026-02-31")).toBeNull();
    expect(accessInventoryChangeEventLocalDayEnd("2026-02-31")).toBeNull();
  });

  it("maps filter inputs onto GraphQL filter fields", () => {
    const withSourceAndType = accessInventoryChangeEventsGraphqlFilter({
      sourceId: "source-gid",
      changeType: "APPEARED",
      observedAfter: "2026-03-10",
      observedBefore: "2026-03-12",
    });

    expect(withSourceAndType.accessReviewSourceId).toBe("source-gid");
    expect(withSourceAndType.changeTypes).toEqual(["APPEARED"]);
    expect(withSourceAndType.observedAtOnOrAfter).toBe(
      accessInventoryChangeEventLocalDayStart("2026-03-10"),
    );
    expect(withSourceAndType.observedAtOnOrBefore).toBe(
      accessInventoryChangeEventLocalDayEnd("2026-03-12"),
    );

    const withoutSourceOrType = accessInventoryChangeEventsGraphqlFilter({
      sourceId: null,
      changeType: null,
      observedAfter: null,
      observedBefore: null,
    });

    expect(withoutSourceOrType.accessReviewSourceId).toBeNull();
    expect(withoutSourceOrType.changeTypes).toBeNull();
    expect(withoutSourceOrType.observedAtOnOrAfter).toBeNull();
    expect(withoutSourceOrType.observedAtOnOrBefore).toBeNull();
  });
});
