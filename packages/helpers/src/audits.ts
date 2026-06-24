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

type Translator = (s: string) => string;

export const auditStates = [
  "NOT_STARTED",
  "IN_PROGRESS",
  "COMPLETED",
  "REJECTED",
  "OUTDATED",
] as const;

export function getAuditStateLabel(__: Translator, state: (typeof auditStates)[number]) {
  switch (state) {
    case "NOT_STARTED":
      return __("Not Started");
    case "IN_PROGRESS":
      return __("In Progress");
    case "COMPLETED":
      return __("Completed");
    case "REJECTED":
      return __("Rejected");
    case "OUTDATED":
      return __("Outdated");
    default:
      return __("Unknown");
  }
}

export function getAuditStateVariant(state: (typeof auditStates)[number]) {
  switch (state) {
    case "NOT_STARTED":
      return "neutral";
    case "IN_PROGRESS":
      return "info";
    case "COMPLETED":
      return "success";
    case "REJECTED":
      return "danger";
    case "OUTDATED":
      return "warning";
    default:
      return "neutral";
  }
}
