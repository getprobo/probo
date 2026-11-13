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

export const findingKinds = [
  "MINOR_NONCONFORMITY",
  "MAJOR_NONCONFORMITY",
  "OBSERVATION",
  "EXCEPTION",
] as const;

export const findingStatuses = [
  "OPEN",
  "IN_PROGRESS",
  "CLOSED",
  "RISK_ACCEPTED",
  "MITIGATED",
  "FALSE_POSITIVE",
] as const;

export const findingPriorities = ["LOW", "MEDIUM", "HIGH"] as const;

export type FindingKind = (typeof findingKinds)[number];
export type FindingStatus = (typeof findingStatuses)[number];
export type FindingPriority = (typeof findingPriorities)[number];

export function isFindingKind(value: string | null): value is FindingKind {
  return value != null && (findingKinds as readonly string[]).includes(value);
}

export function isFindingStatus(value: string | null): value is FindingStatus {
  return value != null && (findingStatuses as readonly string[]).includes(value);
}

export function isFindingPriority(value: string | null): value is FindingPriority {
  return value != null && (findingPriorities as readonly string[]).includes(value);
}

export function findingKindKey(kind: string) {
  return kind.toLowerCase();
}

export function dateInputValue(value: string | null | undefined) {
  if (!value) {
    return "";
  }
  return value.slice(0, 10);
}

export function dateTimeValue(value: string) {
  if (value === "") {
    return null;
  }
  return `${value}T00:00:00.000Z`;
}

const closedStatuses = new Set<FindingStatus>([
  "CLOSED",
  "MITIGATED",
  "FALSE_POSITIVE",
  "RISK_ACCEPTED",
]);

export function findingDueIsOverdue(dueDate: string | null | undefined, status: string) {
  return Boolean(dueDate) && !closedStatuses.has(status as FindingStatus) && isPastDueDate(dueDate);
}

export function findingChoice<T extends string>(value: unknown, allowed: readonly T[]): T | null {
  if (typeof value !== "string") {
    return null;
  }
  return (allowed as readonly string[]).includes(value) ? value as T : null;
}
