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
import { taskPriorityKeys, taskStateKeys } from "./taskState";

export type TaskActivityType = "CREATED" | "UPDATED";

export type TaskActivityField
  = | "NAME"
    | "DESCRIPTION"
    | "STATE"
    | "PRIORITY"
    | "ASSIGNED_TO"
    | "INTERNAL_CONTROL"
    | "DEADLINE"
    | "TIME_ESTIMATE";

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
    const key = taskStateKeys[value as keyof typeof taskStateKeys];
    return key ? t(`states.${key}`) : value;
  }

  if (field === "PRIORITY") {
    const key = taskPriorityKeys[value as keyof typeof taskPriorityKeys];
    return key ? t(`priorities.${key}`) : value;
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

    const unitNames = {
      S: "seconds",
      M: "minutes",
      H: "hours",
      D: "days",
      W: "weeks",
      MO: "months",
      Y: "years",
    } as const;
    const form = parsed.amount === 1 ? "one" : "other";
    return `${parsed.amount} ${t(`duration.${unitNames[parsed.unit]}_${form}`)}`;
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
    return "activity.sentences.created";
  }

  switch (field) {
    case "NAME":
      return "activity.sentences.renamed";
    case "DESCRIPTION":
      return "activity.sentences.descriptionUpdated";
    case "STATE":
      return "activity.sentences.stateChanged";
    case "PRIORITY":
      return "activity.sentences.priorityChanged";
    case "ASSIGNED_TO":
      if (oldValue && newValue) {
        return "activity.sentences.reassigned";
      }
      if (newValue) {
        return "activity.sentences.assigned";
      }
      return "activity.sentences.unassigned";
    case "INTERNAL_CONTROL":
      if (oldValue && newValue) {
        return "activity.sentences.internalControlChanged";
      }
      if (newValue) {
        return "activity.sentences.internalControlLinked";
      }
      return "activity.sentences.internalControlUnlinked";
    case "DEADLINE":
      if (oldValue && newValue) {
        return "activity.sentences.deadlineChanged";
      }
      if (newValue) {
        return "activity.sentences.deadlineSet";
      }
      return "activity.sentences.deadlineRemoved";
    case "TIME_ESTIMATE":
      if (oldValue && newValue) {
        return "activity.sentences.timeEstimateChanged";
      }
      if (newValue) {
        return "activity.sentences.timeEstimateSet";
      }
      return "activity.sentences.timeEstimateRemoved";
    default:
      return "activity.sentences.updated";
  }
}
