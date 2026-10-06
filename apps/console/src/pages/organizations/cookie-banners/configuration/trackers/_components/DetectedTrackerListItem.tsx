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

import { dateTimeFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DetectedTrackerListItem_detectedTracker$key } from "#/__generated__/core/DetectedTrackerListItem_detectedTracker.graphql";

import { detectedTrackerListItem } from "../../../variants";
import { cookieSourceBadges } from "../_lib/trackerBadges";

const detectedTrackerFragment = graphql`
  fragment DetectedTrackerListItem_detectedTracker on DetectedTracker {
    identifier
    initiatorUrl
    maxAgeSeconds
    source
    lastDetectedAt
  }
`;

interface DetectedTrackerListItemProps {
  detectedTrackerKey: DetectedTrackerListItem_detectedTracker$key;
}

export function DetectedTrackerListItem({ detectedTrackerKey }: DetectedTrackerListItemProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const tracker = useFragment(detectedTrackerFragment, detectedTrackerKey);
  const { identifier, url, date } = detectedTrackerListItem();
  const sourceBadge = tracker.source == null ? null : cookieSourceBadges[tracker.source];

  return (
    <TableRow align="center">
      <TableRowHeaderCell overflow="truncate">
        <Text size={2} className={identifier()}>
          {tracker.identifier}
        </Text>
      </TableRowHeaderCell>
      <TableCell overflow="truncate">
        {tracker.initiatorUrl == null
          ? <Text size={2} color="faint">-</Text>
          : (
              <Text size={2} className={url()}>
                {tracker.initiatorUrl}
              </Text>
            )}
      </TableCell>
      <TableCell>
        {tracker.maxAgeSeconds == null
          ? <Text size={2} color="faint">-</Text>
          : (
              <Text size={2}>
                {t("detectedTrackerRow.duration.second", { count: tracker.maxAgeSeconds })}
              </Text>
            )}
      </TableCell>
      <TableCell>
        {sourceBadge == null
          ? <Text size={2} color="faint">-</Text>
          : (
              <Badge variant={sourceBadge.variant} color={sourceBadge.color}>
                {t(`trackerPatternRow.sources.${sourceBadge.labelKey}`)}
              </Badge>
            )}
      </TableCell>
      <TableCell>
        <time dateTime={tracker.lastDetectedAt} className={date()}>
          <Text size={2}>
            {dateTimeFormat(i18n.language, tracker.lastDetectedAt)}
          </Text>
        </time>
      </TableCell>
    </TableRow>
  );
}
