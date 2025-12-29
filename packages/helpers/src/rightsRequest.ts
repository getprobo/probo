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
