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

import { ignoredErrorPatterns } from "./ignore";

// Faro matches ignoreErrors against message, name, and stack joined by spaces.
// A non-Error rejection is reported with this message prefix.
const safeLinksReport
  = "Non-Error promise rejection captured with value: "
    + "Object Not Found Matching Id:8, MethodName:update, ParamCount:4";

function ignored(report: string): boolean {
  return ignoredErrorPatterns.some(pattern => pattern.test(report));
}

describe("ignoredErrorPatterns", () => {
  it("matches the Outlook SafeLinks rejection Faro reports", () => {
    expect(ignored(safeLinksReport)).toBe(true);
  });

  it("does not match a viewer error", () => {
    expect(ignored(
      "TypeError: Promise.try is not a function TypeError at DocumentRequestPanel",
    )).toBe(false);
  });
});
