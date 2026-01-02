type Translator = (s: string) => string;

export const snapshotTypes = [
  "RISKS",
  "VENDORS",
  "ASSETS",
  "DATA",
  "NONCONFORMITIES",
  "OBLIGATIONS",
  "CONTINUAL_IMPROVEMENTS",
  "PROCESSING_ACTIVITIES",
  "STATES_OF_APPLICABILITY",
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
    case "NONCONFORMITIES":
      return __("Nonconformities");
    case "OBLIGATIONS":
      return __("Obligations");
    case "CONTINUAL_IMPROVEMENTS":
      return __("Continual Improvements");
    case "PROCESSING_ACTIVITIES":
      return __("Processing Activities");
    case "STATES_OF_APPLICABILITY":
      return __("States of Applicability");
    default:
      return __("Unknown");
  }
}

export function getSnapshotTypeUrlPath(type?: string): string {
  switch (type) {
    case "RISKS":
      return "/risks";
    case "VENDORS":
      return "/vendors";
    case "ASSETS":
      return "/assets";
    case "DATA":
      return "/data";
    case "NONCONFORMITIES":
      return "/nonconformities";
    case "OBLIGATIONS":
      return "/obligations";
    case "CONTINUAL_IMPROVEMENTS":
      return "/continual-improvements";
    case "PROCESSING_ACTIVITIES":
      return "/processing-activities";
    case "STATES_OF_APPLICABILITY":
      return "/states-of-applicability";
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
