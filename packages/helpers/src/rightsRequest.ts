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

export type RightsRequestType = "ACCESS" | "DELETION" | "PORTABILITY";

export const rightsRequestTypes = [
  "ACCESS",
  "DELETION",
  "PORTABILITY",
] as const;

export type RightsRequestState = "TODO" | "IN_PROGRESS" | "DONE";

export const rightsRequestStates = [
  "TODO",
  "IN_PROGRESS",
  "DONE",
] as const;

export function getRightsRequestTypeLabel(__: Translator, type: RightsRequestType) {
  switch (type) {
    case "ACCESS":
      return __("Access");
    case "DELETION":
      return __("Deletion");
    case "PORTABILITY":
      return __("Portability");
    default:
      return type;
  }
}

export function getRightsRequestTypeOptions(__: Translator) {
  return rightsRequestTypes.map((type) => ({
    value: type,
    label: __({
      "ACCESS": "Access",
      "DELETION": "Deletion",
      "PORTABILITY": "Portability",
    }[type]),
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
    default:
      return "neutral" as const;
  }
};

export function getRightsRequestStateLabel(__: Translator, state: RightsRequestState) {
  switch (state) {
    case "TODO":
      return __("To Do");
    case "IN_PROGRESS":
      return __("In Progress");
    case "DONE":
      return __("Done");
    default:
      return state;
  }
}

export function getRightsRequestStateOptions(__: Translator) {
  return rightsRequestStates.map((state) => ({
    value: state,
    label: __({
      "TODO": "To Do",
      "IN_PROGRESS": "In Progress",
      "DONE": "Done",
    }[state]),
  }));
}
