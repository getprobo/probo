import { getDocumentClassificationLabel } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";

import { Badge } from "../../Atoms/Badge/Badge";

type Props = {
  classification: string;
};

export function DocumentClassificationBadge({ classification }: Props) {
  const { __ } = useTranslate();

  // Choose badge variant based on classification level
  const variant = (() => {
    switch (classification) {
      case "PUBLIC":
        return "success" as const;
      case "INTERNAL":
        return "neutral" as const;
      case "CONFIDENTIAL":
        return "warning" as const;
      case "SECRET":
        return "danger" as const;
      default:
        return "neutral" as const;
    }
  })();

  return (
    <Badge variant={variant}>
      {getDocumentClassificationLabel(__, classification)}
    </Badge>
  );
}
