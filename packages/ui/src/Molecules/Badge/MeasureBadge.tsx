import type { measureStates } from "@probo/helpers";
import { getMeasureStateLabel } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import type { ComponentProps } from "react";

import { Badge } from "../../Atoms/Badge/Badge";

type Props = {
  state: (typeof measureStates)[number];
};

const stateToVariant: Record<
  Props["state"],
  ComponentProps<typeof Badge>["variant"]
> = {
  IMPLEMENTED: "success",
  IN_PROGRESS: "info",
  NOT_APPLICABLE: "neutral",
  NOT_STARTED: "neutral",
};

export function MeasureBadge({ state }: Props) {
  const { __ } = useTranslate();
  return (
    <Badge variant={stateToVariant[state]}>
      {getMeasureStateLabel(__, state)}
    </Badge>
  );
}
