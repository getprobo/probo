type Translator = (s: string) => string;

export type ObligationStatus = "NON_COMPLIANT" | "PARTIALLY_COMPLIANT" | "COMPLIANT";

export const obligationStatuses = [
  "NON_COMPLIANT",
  "PARTIALLY_COMPLIANT",
  "COMPLIANT",
] as const;

export const getObligationStatusVariant = (status: ObligationStatus) => {
  switch (status) {
    case "NON_COMPLIANT":
      return "danger" as const;
    case "PARTIALLY_COMPLIANT":
      return "warning" as const;
    case "COMPLIANT":
      return "success" as const;
    default:
      return "neutral" as const;
  }
};

export const getObligationStatusLabel = (status: ObligationStatus) => {
  switch (status) {
    case "NON_COMPLIANT":
      return "Non-compliant";
    case "PARTIALLY_COMPLIANT":
      return "Partially compliant";
    case "COMPLIANT":
      return "Compliant";
    default:
      return status;
  }
};

export function getObligationStatusOptions(__: Translator) {
  return obligationStatuses.map((status) => ({
    value: status,
    label: __({
      "NON_COMPLIANT": "Non-compliant",
      "PARTIALLY_COMPLIANT": "Partially compliant",
      "COMPLIANT": "Compliant",
    }[status]),
  }));
}
