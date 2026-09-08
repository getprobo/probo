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

import { dateTimeFormat } from "@probo/i18n";

import { parseTaskDuration } from "./taskDuration";

export type TaskActivityType = "CREATED" | "UPDATED";

export type TaskActivityField
  = | "NAME"
    | "DESCRIPTION"
    | "STATE"
    | "PRIORITY"
    | "ASSIGNED_TO"
    | "MEASURE"
    | "DEADLINE"
    | "TIME_ESTIMATE";

const stateLabelKeys = {
  BACKLOG: "detailsPage.states.backlog",
  TODO: "detailsPage.states.todo",
  IN_PROGRESS: "detailsPage.states.inProgress",
  DONE: "detailsPage.states.done",
  CANCELED: "detailsPage.states.canceled",
  DUPLICATE: "detailsPage.states.duplicate",
} as const;

const priorityLabelKeys = {
  URGENT: "detailsPage.priorities.urgent",
  HIGH: "detailsPage.priorities.high",
  MEDIUM: "detailsPage.priorities.medium",
  LOW: "detailsPage.priorities.low",
} as const;

export function formatTaskActivityValue(
  field: TaskActivityField | null | undefined,
  value: string | null | undefined,
  language: string,
  t: (key: string) => string,
): string {
  if (value == null || value === "") {
    return "";
  }

  if (field === "STATE") {
    const key = stateLabelKeys[value as keyof typeof stateLabelKeys];
    return key ? t(key) : value;
  }

  if (field === "PRIORITY") {
    const key = priorityLabelKeys[value as keyof typeof priorityLabelKeys];
    return key ? t(key) : value;
  }

  if (field === "DEADLINE") {
    const formatted = dateTimeFormat(language, value);
    return formatted || value;
  }

  if (field === "TIME_ESTIMATE") {
    const parsed = parseTaskDuration(value);
    if (parsed == null) {
      return value;
    }

    return `${parsed.amount} ${t(`detailsPage.duration.${parsed.unit}`)}`;
  }

  return value;
}

export function taskActivitySentenceKey(
  activityType: TaskActivityType,
  field: TaskActivityField | null | undefined,
  oldValue: string | null | undefined,
  newValue: string | null | undefined,
): string {
  if (activityType === "CREATED") {
    return "detailsPage.activity.sentences.created";
  }

  switch (field) {
    case "NAME":
      return "detailsPage.activity.sentences.renamed";
    case "DESCRIPTION":
      return "detailsPage.activity.sentences.descriptionUpdated";
    case "STATE":
      return "detailsPage.activity.sentences.stateChanged";
    case "PRIORITY":
      return "detailsPage.activity.sentences.priorityChanged";
    case "ASSIGNED_TO":
      if (oldValue && newValue) {
        return "detailsPage.activity.sentences.reassigned";
      }
      if (newValue) {
        return "detailsPage.activity.sentences.assigned";
      }
      return "detailsPage.activity.sentences.unassigned";
    case "MEASURE":
      if (oldValue && newValue) {
        return "detailsPage.activity.sentences.measureChanged";
      }
      if (newValue) {
        return "detailsPage.activity.sentences.measureLinked";
      }
      return "detailsPage.activity.sentences.measureUnlinked";
    case "DEADLINE":
      if (oldValue && newValue) {
        return "detailsPage.activity.sentences.deadlineChanged";
      }
      if (newValue) {
        return "detailsPage.activity.sentences.deadlineSet";
      }
      return "detailsPage.activity.sentences.deadlineRemoved";
    case "TIME_ESTIMATE":
      if (oldValue && newValue) {
        return "detailsPage.activity.sentences.timeEstimateChanged";
      }
      if (newValue) {
        return "detailsPage.activity.sentences.timeEstimateSet";
      }
      return "detailsPage.activity.sentences.timeEstimateRemoved";
    default:
      return "detailsPage.activity.sentences.updated";
  }
}
