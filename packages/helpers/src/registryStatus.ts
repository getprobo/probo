type Translator = (s: string) => string;

export type RegistryStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";

export const registryStatuses = [
  "OPEN",
  "IN_PROGRESS",
  "CLOSED",
] as const;

export const getStatusVariant = (status: RegistryStatus) => {
  switch (status) {
    case "OPEN":
      return "danger" as const;
    case "IN_PROGRESS":
      return "warning" as const;
    case "CLOSED":
      return "success" as const;
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
    }[status]),
  }));
}
