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

const rightsRequestTypeKeys: Record<RightsRequestType, string> = {
  "ACCESS": "helpers.rightsRequestType.access",
  "DELETION": "helpers.rightsRequestType.deletion",
  "RECTIFICATION": "helpers.rightsRequestType.rectification",
  "PORTABILITY": "helpers.rightsRequestType.portability",
  "OBJECTION": "helpers.rightsRequestType.objection",
  "COMPLAINT": "helpers.rightsRequestType.complaint",
};

export function getRightsRequestTypeLabel(t: Translator, type: RightsRequestType) {
  return t(rightsRequestTypeKeys[type]);
}

export function getRightsRequestTypeOptions(t: Translator) {
  return rightsRequestTypes.map((type) => ({
    value: type,
    label: t(rightsRequestTypeKeys[type]),
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

const rightsRequestStateKeys: Record<RightsRequestState, string> = {
  "TODO": "helpers.rightsRequestState.todo",
  "IN_PROGRESS": "helpers.rightsRequestState.inProgress",
  "DONE": "helpers.rightsRequestState.done",
  "REJECTED": "helpers.rightsRequestState.rejected",
};

export function getRightsRequestStateLabel(t: Translator, state: RightsRequestState) {
  return t(rightsRequestStateKeys[state]);
}

export function getRightsRequestStateOptions(t: Translator) {
  return rightsRequestStates.map((state) => ({
    value: state,
    label: t(rightsRequestStateKeys[state]),
  }));
}
