import { useTranslate } from "@probo/i18n";
import { Badge } from "../../Atoms/Badge/Badge.tsx";

type Props = {
    severity: number;
};

const badgeVariant = (severity: number) => {
    if (severity >= 75) {
        return "danger";
    }
    if (severity >= 50) {
        return "warning";
    }
    if (severity >= 25) {
        return "info";
    }
    return "success";
};

export function SeverityBadge({ severity }: Props) {
    const { __ } = useTranslate();
    const label = () => {
        if (severity >= 75) {
            return __("Critical");
        }
        if (severity >= 50) {
            return __("High");
        }
        if (severity >= 25) {
            return __("Low");
        }
        return __("None");
    };
    return <Badge variant={badgeVariant(severity)}>{label()}</Badge>;
}
