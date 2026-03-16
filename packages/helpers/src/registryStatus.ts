type Translator = (s: string) => string;

export type RegistryStatus = "CLOSED" | "FALSE_POSITIVE" | "IN_PROGRESS" | "MITIGATED" | "OPEN" | "RISK_ACCEPTED";

export const registryStatuses = [
  "OPEN",
  "IN_PROGRESS",
  "CLOSED",
  "RISK_ACCEPTED",
  "MITIGATED",
  "FALSE_POSITIVE",
] as const;

export const getStatusVariant = (status: RegistryStatus) => {
  switch (status) {
    case "OPEN":
      return "danger" as const;
    case "IN_PROGRESS":
      return "warning" as const;
    case "CLOSED":
      return "success" as const;
    case "MITIGATED":
      return "success" as const;
    case "RISK_ACCEPTED":
      return "neutral" as const;
    case "FALSE_POSITIVE":
      return "neutral" as const;
    default:
      return "neutral" as const;
  }
};

export const getStatusLabel = (status: RegistryStatus) => {
  switch (status) {
    case "OPEN":
      return "Open";
    case "IN_PROGRESS":
      return "In Progress";
    case "CLOSED":
      return "Closed";
    case "RISK_ACCEPTED":
      return "Risk Accepted";
    case "MITIGATED":
      return "Mitigated";
    case "FALSE_POSITIVE":
      return "False Positive";
    default:
      return status;
  }
};

export function getStatusOptions(__: Translator) {
  return registryStatuses.map((status) => ({
    value: status,
    label: __({
      "OPEN": "Open",
      "IN_PROGRESS": "In Progress",
      "CLOSED": "Closed",
      "RISK_ACCEPTED": "Risk Accepted",
      "MITIGATED": "Mitigated",
      "FALSE_POSITIVE": "False Positive",
    }[status]),
  }));
}
