type Translator = (s: string) => string;

export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";

export const trustCenterVisibilities = [
  "NONE",
  "PRIVATE",
  "PUBLIC",
] as const;

export const getTrustCenterVisibilityVariant = (visibility: TrustCenterVisibility) => {
  switch (visibility) {
    case "NONE":
      return "danger" as const;
    case "PRIVATE":
      return "warning" as const;
    case "PUBLIC":
      return "success" as const;
    default:
      return "neutral" as const;
  }
};

export const getTrustCenterVisibilityLabel = (visibility: TrustCenterVisibility) => {
  switch (visibility) {
    case "NONE":
      return "None";
    case "PRIVATE":
      return "Private";
    case "PUBLIC":
      return "Public";
    default:
      return visibility;
  }
};

export function getTrustCenterVisibilityOptions(__: Translator) {
  return trustCenterVisibilities.map((visibility) => ({
    value: visibility,
    label: __({
      "NONE": "None",
      "PRIVATE": "Private",
      "PUBLIC": "Public",
    }[visibility]),
    variant: getTrustCenterVisibilityVariant(visibility),
  }));
}
