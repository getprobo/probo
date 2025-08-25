type Translator = (s: string) => string;

export const snapshotTypes = [
  "RISKS",
  "VENDORS",
  "ASSETS",
  "DATA",
  "NON_CONFORMITY_REGISTRIES",
  "COMPLIANCE_REGISTRIES"
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
    case "NON_CONFORMITY_REGISTRIES":
      return __("Non Conformity Registries");
    case "COMPLIANCE_REGISTRIES":
      return __("Compliance Registries");
    default:
      return __("Unknown");
  }
}
