// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
