import { useTranslate } from "@probo/i18n";

import { Badge } from "../../Atoms/Badge/Badge";

type Props = {
  level: number | string;
};

const badgeVariant = (level: string | number) => {
  if (typeof level === "number") {
    if (level >= 15) {
      level = "CRITICAL";
    } else if (level >= 8) {
      level = "HIGH";
    } else {
      level = "LOW";
    }
  }
  switch (level) {
    case "CRITICAL":
      return "danger";
    case "HIGH":
      return "warning";
    case "LOW":
      return "success";
    case "MEDIUM":
      return "info";
    default:
      return "neutral";
  }
};

export function RiskBadge({ level }: Props) {
  const { __ } = useTranslate();
  const label = () => {
    if (typeof level === "number") {
      if (level >= 15) {
        return __("High");
      }
      if (level >= 8) {
        return __("Medium");
      }
      return __("Low");
    }
    switch (level) {
      case "CRITICAL":
        return __("Critical");
      case "HIGH":
        return __("High");
      case "LOW":
        return __("Low");
      case "MEDIUM":
        return __("Medium");
      case "NONE":
        return __("None");
      default:
        return __("Low");
    }
  };
  return <Badge variant={badgeVariant(level)}>{label()}</Badge>;
}
