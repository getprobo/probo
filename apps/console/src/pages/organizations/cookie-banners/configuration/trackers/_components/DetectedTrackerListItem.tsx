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

import { InfoIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { TableRowHeaderCell } from "@probo/ui/src/v2/Table/TableRowHeaderCell";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DetectedTrackerListItem_detectedTracker$key } from "#/__generated__/core/DetectedTrackerListItem_detectedTracker.graphql";

import { detectedTrackerListItem } from "../../../variants";
import { cookieSourceBadges } from "../_lib/trackerBadges";

import { TrackerObserverBadges } from "./TrackerObserverBadges";

const sameSiteKeys = {
  STRICT: "Strict",
  LAX: "Lax",
  NONE: "None",
} as const;

function detailRow(label: string, value: string | null) {
  if (value == null) {
    return null;
  }

  return { label, value };
}

const detectedTrackerFragment = graphql`
  fragment DetectedTrackerListItem_detectedTracker on DetectedTracker {
    identifier
    cookieDomain
    hostOnly
    initiatorUrl
    maxAgeSeconds
    source
    observedBy
    valueSize
    httpOnly
    secure
    sameSite
    partitioned
    setByUrl
    createdAt
  }
`;

interface DetectedTrackerListItemProps {
  detectedTrackerKey: DetectedTrackerListItem_detectedTracker$key;
}

export function DetectedTrackerListItem({ detectedTrackerKey }: DetectedTrackerListItemProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const tracker = useFragment(detectedTrackerFragment, detectedTrackerKey);
  const { heading, identifier, info, detail, fact, url, date } = detectedTrackerListItem();
  const sourceBadge = tracker.source == null ? null : cookieSourceBadges[tracker.source];
  const yesNo = (value: boolean | null | undefined) => (
    value == null ? null : t(value ? "detectedTrackerRow.details.yes" : "detectedTrackerRow.details.no")
  );
  const sameSiteLabel = tracker.sameSite == null
    ? null
    : t(`detectedTrackerRow.details.sameSite${sameSiteKeys[tracker.sameSite]}`);
  const details = [
    detailRow(t("detectedTrackerRow.details.httpOnly"), yesNo(tracker.httpOnly)),
    detailRow(t("detectedTrackerRow.details.secure"), yesNo(tracker.secure)),
    detailRow(t("detectedTrackerRow.details.sameSite"), sameSiteLabel),
    detailRow(t("detectedTrackerRow.details.partitioned"), yesNo(tracker.partitioned)),
    detailRow(t("detectedTrackerRow.details.setBy"), tracker.setByUrl ?? null),
    detailRow(
      t("detectedTrackerRow.details.valueSize"),
      tracker.valueSize == null
        ? null
        : t("detectedTrackerRow.details.bytes", { count: tracker.valueSize }),
    ),
  ].filter(item => item != null);

  return (
    <TableRow align="center">
      <TableRowHeaderCell overflow="truncate">
        <div className={heading()}>
          <Text size={2} className={identifier()}>
            {tracker.identifier}
          </Text>
          {details.length > 0
            ? (
                <span className={info()}>
                  <Popover>
                    <PopoverTrigger
                      render={(
                        <IconButton
                          size={1}
                          variant="ghost"
                          color="neutral"
                          aria-label={t("detectedTrackerRow.details.label")}
                        />
                      )}
                    >
                      <InfoIcon />
                    </PopoverTrigger>
                    <PopoverPopup>
                      <div className={detail()}>
                        <Text size={2} weight="medium" highContrast>
                          {tracker.identifier}
                        </Text>
                        {details.map(item => (
                          <div key={item.label} className={fact()}>
                            <Text size={2} color="faint">{item.label}</Text>
                            <Text size={2} className={url()}>{item.value}</Text>
                          </div>
                        ))}
                      </div>
                    </PopoverPopup>
                  </Popover>
                </span>
              )
            : null}
        </div>
      </TableRowHeaderCell>
      <TableCell overflow="truncate">
        {tracker.hostOnly === true
          ? <Text size={2}>{t("detectedTrackerRow.hostOnly")}</Text>
          : tracker.cookieDomain == null
            ? <Text size={2} color="faint">-</Text>
            : (
                <Text size={2} className={url()}>
                  {tracker.cookieDomain}
                </Text>
              )}
      </TableCell>
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
        <TrackerObserverBadges observers={[tracker.observedBy]} />
      </TableCell>
      <TableCell>
        <time dateTime={tracker.createdAt} className={date()}>
          <Text size={2}>
            {dateTimeFormat(i18n.language, tracker.createdAt)}
          </Text>
        </time>
      </TableCell>
    </TableRow>
  );
}
