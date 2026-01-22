import { useTranslate } from "@probo/i18n";

import { Badge } from "../../Atoms/Badge/Badge";

type Props = {
  state: "DRAFT" | "PUBLISHED";
};

export function DocumentVersionBadge(props: Props) {
  const { __ } = useTranslate();
  return (
    <Badge variant={props.state === "DRAFT" ? "neutral" : "success"}>
      {props.state === "DRAFT" ? __("Draft") : __("Published")}
    </Badge>
  );
}
