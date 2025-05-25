type Translator = (s: string) => string;

export const measureStates = [
    "IMPLEMENTED",
    "IN_PROGRESS",
    "NOT_APPLICABLE",
    "NOT_STARTED",
] as const;

export function getMeasureStateLabel(__: Translator, state: string) {
    switch (state) {
        case "IMPLEMENTED":
            return __("Implemented");
        case "IN_PROGRESS":
            return __("In Progress");
        case "NOT_APPLICABLE":
            return __("Not Applicable");
        case "NOT_STARTED":
            return __("Not Started");
        default:
            return __("Unknown");
    }
}
