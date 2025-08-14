export type NonconformityRegistryStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";

export const getStatusVariant = (status: NonconformityRegistryStatus) => {
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

export const getStatusLabel = (status: NonconformityRegistryStatus) => {
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
