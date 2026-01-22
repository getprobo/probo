import { useTranslate } from "@probo/i18n";

import { Badge } from "../../Atoms/Badge/Badge";

type Props = {
  score: number;
};

const badgeVariant = (score: number) => {
  if (score >= 15) {
    return "danger";
  }
  if (score > 6) {
    return "warning";
  }
  return "success";
};

export function SeverityBadge({ score }: Props) {
  const { __ } = useTranslate();
  const label = () => {
    if (score >= 15) {
      return __("High");
    }
    if (score > 6) {
      return __("Medium");
    }
    return __("Low");
  };
  return <Badge variant={badgeVariant(score)}>{label()}</Badge>;
}
