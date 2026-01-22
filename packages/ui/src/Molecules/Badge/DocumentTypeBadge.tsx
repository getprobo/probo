import { getDocumentTypeLabel } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";

import { Badge } from "../../Atoms/Badge/Badge";

type Props = {
  type: string;
};

export function DocumentTypeBadge({ type }: Props) {
  const { __ } = useTranslate();
  return <Badge variant="neutral">{getDocumentTypeLabel(__, type)}</Badge>;
}
