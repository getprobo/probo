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

export type RightsRequestType =
  | "ACCESS"
  | "DELETION"
  | "RECTIFICATION"
  | "PORTABILITY"
  | "OBJECTION"
  | "COMPLAINT";

export const rightsRequestTypes = [
  "ACCESS",
  "DELETION",
  "RECTIFICATION",
  "PORTABILITY",
  "OBJECTION",
  "COMPLAINT",
] as const;

export type RightsRequestState = "TODO" | "IN_PROGRESS" | "DONE" | "REJECTED";

export const rightsRequestStates = [
  "TODO",
  "IN_PROGRESS",
  "DONE",
  "REJECTED",
] as const;

const rightsRequestTypeLabels: Record<RightsRequestType, string> = {
  "ACCESS": "Access",
  "DELETION": "Deletion",
  "RECTIFICATION": "Rectification",
  "PORTABILITY": "Portability",
  "OBJECTION": "Objection",
  "COMPLAINT": "Complaint",
};

export function getRightsRequestTypeLabel(__: Translator, type: RightsRequestType) {
  return __(rightsRequestTypeLabels[type] ?? type);
}

export function getRightsRequestTypeOptions(__: Translator) {
  return rightsRequestTypes.map((type) => ({
    value: type,
    label: __(rightsRequestTypeLabels[type]),
  }));
}

export const getRightsRequestStateVariant = (
  state: RightsRequestState
): "danger" | "warning" | "success" | "neutral" | "info" | "outline" | "highlight" => {
  switch (state) {
    case "TODO":
      return "warning" as const;
    case "IN_PROGRESS":
      return "info" as const;
    case "DONE":
      return "success" as const;
    case "REJECTED":
      return "danger" as const;
    default:
      return "neutral" as const;
  }
};

const rightsRequestStateLabels: Record<RightsRequestState, string> = {
  "TODO": "To Do",
  "IN_PROGRESS": "In Progress",
  "DONE": "Done",
  "REJECTED": "Rejected",
};

export function getRightsRequestStateLabel(__: Translator, state: RightsRequestState) {
  return __(rightsRequestStateLabels[state] ?? state);
}

export function getRightsRequestStateOptions(__: Translator) {
  return rightsRequestStates.map((state) => ({
    value: state,
    label: __(rightsRequestStateLabels[state]),
  }));
}
