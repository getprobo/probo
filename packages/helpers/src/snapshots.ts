type Translator = (s: string) => string;

export const snapshotTypes = [
  "DATA",
  "NONCONFORMITY_REGISTRIES",
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
    default:
      return __("Unknown");
  }
}

export function getSnapshotTypeUrlPath(type?: string): string {
  switch (type) {
    case "DATA":
      return "/data";
    case "NONCONFORMITY_REGISTRIES":
      return "/nonconformity-registries";
    default:
      return "";
  }
}
