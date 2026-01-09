type Status =
  | "ACTIVE"
  | "PROVISIONING"
  | "RENEWING"
  | "PENDING"
  | "FAILED"
  | "EXPIRED";

export const getCustomDomainStatusBadgeVariant = (sslStatus: Status) => {
  switch (sslStatus) {
    case "ACTIVE":
      return "success" as const;
    case "PROVISIONING":
    case "RENEWING":
    case "PENDING":
      return "warning" as const;
    case "FAILED":
    case "EXPIRED":
      return "danger" as const;
    default:
      return "neutral" as const;
  }
};

export const getCustomDomainStatusBadgeLabel = (
  sslStatus: Status,
  __: (key: string) => string,
) => {
  if (sslStatus === "ACTIVE") {
    return __("Active");
  }
  if (sslStatus === "PROVISIONING" || sslStatus === "RENEWING") {
    return __("Provisioning");
  }
  if (sslStatus === "PENDING") {
    return __("Pending");
  }
  if (sslStatus === "FAILED") {
    return __("Failed");
  }
  if (sslStatus === "EXPIRED") {
    return __("Expired");
  }
  return __("Unknown");
};
