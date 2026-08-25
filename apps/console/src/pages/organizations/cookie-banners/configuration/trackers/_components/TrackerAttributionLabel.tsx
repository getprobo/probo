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

import { Badge } from "@probo/ui";
import { useTranslation } from "react-i18next";

import type { CommonTrackerPatternAttribution } from "#/__generated__/core/TrackerPatternRowFragment.graphql";

interface TrackerAttributionLabelProps {
  attribution: CommonTrackerPatternAttribution | null | undefined;
}

export function TrackerAttributionLabel({ attribution }: TrackerAttributionLabelProps) {
  const { t } = useTranslation("organizations/cookie-banners");

  switch (attribution) {
    case "FIRST_PARTY":
      return <Badge variant="success">{t("trackerAttribution.firstParty")}</Badge>;
    case "NOT_ATTRIBUTABLE":
      return <Badge variant="warning">{t("trackerAttribution.visitorSoftware")}</Badge>;
    case "UNDETERMINED":
      return (
        <span className="text-txt-tertiary text-sm">
          {t("trackerAttribution.identifying")}
        </span>
      );
    default:
      return <span className="text-txt-tertiary text-sm">-</span>;
  }
}
