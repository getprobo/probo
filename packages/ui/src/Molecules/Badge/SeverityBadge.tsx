import { useTranslate } from "@probo/i18n";
import { Badge } from "../../Atoms/Badge/Badge.tsx";

type Props = {
    score: number;
};

const badgeVariant = (score: number) => {
    if (score >= 15) {
        return "danger";
    }
    if (score >= 5) {
        return "warning";
    }
    return "neutral";
};

export function SeverityBadge({ score }: Props) {
    const { __ } = useTranslate();
    const label = () => {
        if (score >= 15) {
            return __("Critical");
        }
        if (score >= 5) {
            return __("High");
        }
        return __("Negligible");
    };
    return <Badge variant={badgeVariant(score)}>{label()}</Badge>;
}
