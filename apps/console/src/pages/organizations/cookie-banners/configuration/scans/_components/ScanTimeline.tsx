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

import { CheckCircleIcon, ClockIcon, PlayIcon, WarningCircleIcon } from "@phosphor-icons/react";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Timeline } from "@probo/ui/src/v2/Timeline/Timeline";
import { TimelineContent } from "@probo/ui/src/v2/Timeline/TimelineContent";
import { TimelineItem } from "@probo/ui/src/v2/Timeline/TimelineItem";
import { TimelineMarker } from "@probo/ui/src/v2/Timeline/TimelineMarker";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ScanTimeline_scan$key } from "#/__generated__/core/ScanTimeline_scan.graphql";

import { scanSection, scanTimeline } from "../../../variants";
import { formatScanDuration, formatScanTime, scanLifecycleDurations } from "../_lib/formatScanTime";
import { scanFailureReasonKey } from "../_lib/scanStates";

const scanTimelineFragment = graphql`
  fragment ScanTimeline_scan on CookieBannerScan {
    state
    failureReason
    createdAt
    startedAt
    completedAt
  }
`;

interface ScanTimelineProps {
  scanKey: ScanTimeline_scan$key;
}

export function ScanTimeline({ scanKey }: ScanTimelineProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const { t: tRoot } = useTranslation();
  const scan = useFragment(scanTimelineFragment, scanKey);
  const { root } = scanSection();
  const { frame, columns, durations, duration, value } = scanTimeline();
  const times = scanLifecycleDurations(scan);
  const durationItems = [
    { label: t("scanPage.timeline.totalElapsed"), milliseconds: times.totalMs },
    { label: t("scanPage.timeline.queueTime"), milliseconds: times.queueMs },
    { label: t("scanPage.timeline.runningTime"), milliseconds: times.runningMs },
  ];
  const failed = scan.state === "FAILED";
  const events: {
    label: string;
    at: string;
    icon: ReactNode;
    color: "neutral" | "red";
  }[] = [
    {
      label: t("scanPage.timeline.queued"),
      at: scan.createdAt,
      icon: <ClockIcon aria-hidden />,
      color: "neutral",
    },
  ];
  if (scan.startedAt != null) {
    events.push({
      label: t("scanPage.timeline.started"),
      at: scan.startedAt,
      icon: <PlayIcon aria-hidden />,
      color: "neutral",
    });
  }
  if (scan.completedAt != null) {
    events.push({
      label: t("scanPage.timeline.finished"),
      at: scan.completedAt,
      icon: failed ? <WarningCircleIcon aria-hidden /> : <CheckCircleIcon aria-hidden />,
      color: failed ? "red" : "neutral",
    });
  }
  const reasonKey = scan.failureReason == null ? null : scanFailureReasonKey(scan.failureReason);

  return (
    <section className={root()}>
      <Heading level={2} size={4} weight="medium" highContrast>
        {t("scanPage.timeline.title")}
      </Heading>
      {failed
        ? (
            <Callout color="red">
              {reasonKey == null
                ? t("scanPage.failure.unknown")
                : t(`scanPage.failure.${reasonKey}`)}
            </Callout>
          )
        : null}
      <Card variant="soft" size={3} padding={4} className={frame()}>
        <div className={columns()}>
          <div className={durations()}>
            {durationItems.map(item => (
              <div key={item.label} className={duration()}>
                <Text size={1} color="faint">{item.label}</Text>
                <Text size={2} highContrast className={value()}>
                  {formatScanDuration(i18n.language, item.milliseconds, tRoot)}
                </Text>
              </div>
            ))}
          </div>
          <Timeline>
            {events.map(event => (
              <TimelineItem key={event.label}>
                <TimelineMarker color={event.color}>{event.icon}</TimelineMarker>
                <TimelineContent>
                  <Text size={1} highContrast>{event.label}</Text>
                  <Text size={1} color="faint" className="shrink-0 whitespace-nowrap">
                    {formatScanTime(i18n.language, event.at, true)}
                  </Text>
                </TimelineContent>
              </TimelineItem>
            ))}
          </Timeline>
        </div>
      </Card>
    </section>
  );
}
