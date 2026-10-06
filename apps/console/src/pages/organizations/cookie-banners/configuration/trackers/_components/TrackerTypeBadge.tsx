// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { useTranslation } from "react-i18next";

import { cookieBannerList } from "../../../variants";
import { trackerTypeBadges } from "../_lib/trackerBadges";
import type { TrackerType } from "../_lib/useTrackersListFilters";

interface TrackerTypeBadgeProps {
  trackerType: TrackerType;
  tooltip?: boolean;
}

export function TrackerTypeBadge({ trackerType, tooltip = true }: TrackerTypeBadgeProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const spec = trackerTypeBadges[trackerType];
  const name = t(`trackerPatternRow.types.${spec.labelKey}`);
  const badge = (
    <Badge variant={spec.variant} color={spec.color}>
      {t(`trackerPatternRow.types.${spec.shortKey}`)}
    </Badge>
  );

  if (!tooltip) {
    return badge;
  }

  return (
    <Tooltip>
      <TooltipTrigger render={<span tabIndex={0} className="inline-flex">{badge}</span>} />
      <TooltipPopup>{name}</TooltipPopup>
    </Tooltip>
  );
}

interface TrackerTypeOptionProps {
  trackerType: TrackerType;
}

export function TrackerTypeOption({ trackerType }: TrackerTypeOptionProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { typeOption } = cookieBannerList();
  const spec = trackerTypeBadges[trackerType];

  return (
    <span className={typeOption()}>
      <TrackerTypeBadge trackerType={trackerType} tooltip={false} />
      {t(`trackerPatternRow.types.${spec.labelKey}`)}
    </span>
  );
}
