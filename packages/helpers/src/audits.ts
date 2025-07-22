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
