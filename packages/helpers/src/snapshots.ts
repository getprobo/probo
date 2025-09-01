type Translator = (s: string) => string;

export const snapshotTypes = [
  "VENDORS",
  "ASSETS",
  "DATA",
  "NONCONFORMITY_REGISTRIES",
  "COMPLIANCE_REGISTRIES",
  "CONTINUAL_IMPROVEMENT_REGISTRIES",
  "PROCESSING_ACTIVITY_REGISTRIES",
] as const;

export function getSnapshotTypeLabel(__: Translator, type: string | null | undefined) {
  if (!type) {
    return __("Unknown");
  }

  switch (type) {
    case "RISKS":
      return __("Risks");
    case "VENDORS":
      return __("Vendors");
    case "ASSETS":
      return __("Assets");
    case "DATA":
      return __("Data");
    case "NONCONFORMITY_REGISTRIES":
      return __("Nonconformity Registries");
    case "COMPLIANCE_REGISTRIES":
      return __("Compliance Registries");
    case "CONTINUAL_IMPROVEMENT_REGISTRIES":
      return __("Continual Improvement Registries");
    case "PROCESSING_ACTIVITY_REGISTRIES":
      return __("Processing Activity Registries");
    default:
      return __("Unknown");
  }
}

export function getSnapshotTypeUrlPath(type?: string): string {
  switch (type) {
    case "VENDORS":
      return "/vendors";
    case "ASSETS":
      return "/assets";
    case "DATA":
      return "/data";
    case "NONCONFORMITY_REGISTRIES":
      return "/nonconformity-registries";
    case "COMPLIANCE_REGISTRIES":
      return "/compliance-registries";
    case "CONTINUAL_IMPROVEMENT_REGISTRIES":
      return "/continual-improvement-registries";
    case "PROCESSING_ACTIVITY_REGISTRIES":
      return "/processing-activity-registries";
    default:
      return "";
  }
}

export interface SnapshotableResource {
  snapshotId?: string | null | undefined;
}

export function validateSnapshotConsistency(
  resource: SnapshotableResource | null | undefined,
  urlSnapshotId?: string | null | undefined
): void {
  if (resource && resource.snapshotId !== (urlSnapshotId ?? null)) {
    throw new Error("PAGE_NOT_FOUND");
  }
}
