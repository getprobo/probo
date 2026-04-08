import { useTranslate } from "@probo/i18n";
import { Badge } from "@probo/ui";

export function CookieBannerStateBadge({
  state,
}: {
  state: string;
}) {
  const { __ } = useTranslate();

  switch (state) {
    case "PUBLISHED":
      return <Badge variant="success">{__("Published")}</Badge>;
    case "DRAFT":
      return <Badge variant="warning">{__("Draft")}</Badge>;
    case "DISABLED":
      return <Badge variant="danger">{__("Disabled")}</Badge>;
    default:
      return <Badge>{state}</Badge>;
  }
}
