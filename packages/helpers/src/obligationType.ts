type Translator = (s: string) => string;

export type ObligationType = "LEGAL" | "CONTRACTUAL";

export const obligationTypes = [
  "LEGAL",
  "CONTRACTUAL",
] as const;

export const getObligationTypeLabel = (type: ObligationType) => {
  switch (type) {
    case "LEGAL":
      return "Legal";
    case "CONTRACTUAL":
      return "Contractual";
    default:
      return type;
  }
};

export function getObligationTypeOptions(__: Translator) {
  return obligationTypes.map((type) => ({
    value: type,
    label: __({
      "LEGAL": "Legal",
      "CONTRACTUAL": "Contractual",
    }[type]),
  }));
}
