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

import { SCAN_MAX_PAGES } from "../_lib/pageSize";
import { scanStateBadge } from "../_lib/scanStates";

interface ScanStateBadgeProps {
  state: string;
}

export function ScanStateBadge({ state }: ScanStateBadgeProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const badge = scanStateBadge(state);

  return (
    <Badge variant={badge.variant} color={badge.color}>
      {t(`scansPage.states.${badge.labelKey}`)}
    </Badge>
  );
}

export function ScanTruncatedBadge() {
  const { t } = useTranslation("organizations/cookie-banners");

  return (
    <Tooltip>
      <TooltipTrigger
        render={(
          <span tabIndex={0} className="inline-flex">
            <Badge variant="outline" color="amber">
              {t("scansPage.badges.truncated")}
            </Badge>
          </span>
        )}
      />
      <TooltipPopup className="max-w-64">
        {t("scansPage.badges.truncatedHint", { count: SCAN_MAX_PAGES })}
      </TooltipPopup>
    </Tooltip>
  );
}
