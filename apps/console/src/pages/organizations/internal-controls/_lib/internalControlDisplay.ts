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

import { isPastDueDate } from "@probo/helpers";

export type ImplementationStatusColor = "neutral" | "amber" | "sky" | "green";

export function implementationStatusColor(status: string): ImplementationStatusColor {
  switch (status) {
    case "IN_PROGRESS":
      return "amber";
    case "IMPLEMENTED":
      return "sky";
    case "OPERATING":
      return "green";
    default:
      return "neutral";
  }
}

export interface NextDue {
  kind: "evidence" | "test";
  at: string;
  overdue: boolean;
}

export function nextDue(
  evidenceDue: string | null | undefined,
  testDue: string | null | undefined,
): NextDue | null {
  const evidenceTime = evidenceDue ? Date.parse(evidenceDue) : Number.NaN;
  const testTime = testDue ? Date.parse(testDue) : Number.NaN;
  const hasEvidence = evidenceDue != null && !Number.isNaN(evidenceTime);
  const hasTest = testDue != null && !Number.isNaN(testTime);

  if (!hasEvidence && !hasTest) {
    return null;
  }

  if (!hasTest || (hasEvidence && evidenceTime <= testTime)) {
    return {
      kind: "evidence",
      at: evidenceDue ?? "",
      overdue: isPastDueDate(evidenceDue),
    };
  }

  return {
    kind: "test",
    at: testDue ?? "",
    overdue: isPastDueDate(testDue),
  };
}
