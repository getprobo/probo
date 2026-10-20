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

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { isPastDueDate, todayAsDateInput } from "./date";

describe("isPastDueDate", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 9, 7, 15, 30, 0));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("is false for empty values", () => {
    expect(isPastDueDate(null)).toBe(false);
    expect(isPastDueDate(undefined)).toBe(false);
    expect(isPastDueDate("")).toBe(false);
  });

  it("is false for today", () => {
    expect(isPastDueDate(todayAsDateInput())).toBe(false);
    expect(isPastDueDate("2026-10-07T00:00:00Z")).toBe(false);
  });

  it("is false for future calendar days", () => {
    expect(isPastDueDate("2026-10-08")).toBe(false);
    expect(isPastDueDate("2026-10-08T00:00:00Z")).toBe(false);
  });

  it("is true for past calendar days", () => {
    expect(isPastDueDate("2026-10-06")).toBe(true);
    expect(isPastDueDate("2026-10-06T23:59:59Z")).toBe(true);
  });
});
