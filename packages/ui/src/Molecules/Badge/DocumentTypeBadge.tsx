import { Badge } from "../../Atoms/Badge/Badge";
import { getDocumentTypeLabel } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";

type Props = {
    type: string;
};

export function DocumentTypeBadge({ type }: Props) {
    const { __ } = useTranslate();
    return <Badge variant="neutral">{getDocumentTypeLabel(__, type)}</Badge>;
}
