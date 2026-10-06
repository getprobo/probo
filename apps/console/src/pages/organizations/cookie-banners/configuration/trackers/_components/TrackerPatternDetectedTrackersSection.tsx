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

import { CaretDownIcon } from "@phosphor-icons/react";
import { dateTimeFormat } from "@probo/i18n";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { TrackerPatternDetectedTrackersSection_trackerPattern$key } from "#/__generated__/core/TrackerPatternDetectedTrackersSection_trackerPattern.graphql";
import type {
  DetectedTrackerOrder,
  DetectedTrackerOrderField,
  TrackerPatternDetectedTrackersSectionRefetchQuery,
} from "#/__generated__/core/TrackerPatternDetectedTrackersSectionRefetchQuery.graphql";

import { trackerPatternDetectedTrackersSection } from "../../../variants";
import {
  defaultDetectedTrackersOrder,
  detectedTrackersHeaderSort,
  nextDetectedTrackersOrder,
} from "../_lib/detectedTrackersOrder";
import { DETECTED_TRACKERS_PAGE_SIZE } from "../_lib/pageSize";

import { DetectedTrackerListItem } from "./DetectedTrackerListItem";

export const trackerPatternDetectedTrackersSectionFragment = graphql`
  fragment TrackerPatternDetectedTrackersSection_trackerPattern on TrackerPattern
  @refetchable(queryName: "TrackerPatternDetectedTrackersSectionRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "DetectedTrackerOrder", defaultValue: { field: LAST_DETECTED_AT, direction: DESC } }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    detectedCount
    lastMatchedAt
    detectedTrackers(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "TrackerPatternDetectedTrackersSection_detectedTrackers", filters: ["orderBy"]) {
      edges {
        node {
          id
          ...DetectedTrackerListItem_detectedTracker
        }
      }
    }
  }
`;

interface TrackerPatternDetectedTrackersSectionProps {
  trackerPatternKey: TrackerPatternDetectedTrackersSection_trackerPattern$key;
}

export function TrackerPatternDetectedTrackersSection({
  trackerPatternKey,
}: TrackerPatternDetectedTrackersSectionProps) {
  const { t, i18n } = useTranslation("organizations/cookie-banners");
  const [order, setOrder] = useState<DetectedTrackerOrder>(defaultDetectedTrackersOrder);
  const [isSortPending, startSortTransition] = useTransition();

  const { data, hasNext, loadNext, isLoadingNext, refetch } = usePaginationFragment<
    TrackerPatternDetectedTrackersSectionRefetchQuery,
    TrackerPatternDetectedTrackersSection_trackerPattern$key
  >(trackerPatternDetectedTrackersSectionFragment, trackerPatternKey);

  const isPending = isSortPending || isLoadingNext;
  const { root, intro, results, pager, empty } = trackerPatternDetectedTrackersSection({
    pending: isPending,
  });
  const trackers = data.detectedTrackers?.edges.map(edge => edge.node) ?? [];
  const summary = data.lastMatchedAt == null
    ? t("detectedTrackersSection.summaryNever", { count: data.detectedCount })
    : t("detectedTrackersSection.summary", {
        count: data.detectedCount,
        date: dateTimeFormat(i18n.language, data.lastMatchedAt),
      });

  function handleSort(field: DetectedTrackerOrderField) {
    const next = nextDetectedTrackersOrder(field, order);
    startSortTransition(() => {
      setOrder(next);
      refetch({
        order: next,
        first: DETECTED_TRACKERS_PAGE_SIZE,
        after: null,
        last: null,
        before: null,
      });
    });
  }

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("detectedTrackersSection.title")}
        </Heading>
        <Text size={2} color="neutral">
          {summary}
        </Text>
      </div>

      {trackers.length > 0
        ? (
            <>
              <div aria-busy={isPending} className={results()}>
                <Table variant="surface" layout="fixed">
                  <TableHeader>
                    <TableRow>
                      <TableColumnHeaderCell overflow="truncate">
                        {t("detectedTrackersSection.columns.identifier")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell
                        overflow="truncate"
                        sort={detectedTrackersHeaderSort("INITIATOR_URL", order)}
                        onSort={() => handleSort("INITIATOR_URL")}
                        aria-label={t("detectedTrackersSection.sort.initiatorUrl")}
                      >
                        {t("detectedTrackersSection.columns.initiatorUrl")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="7rem">
                        {t("detectedTrackersSection.columns.maxAge")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="7rem">
                        {t("detectedTrackersSection.columns.source")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell
                        width="10rem"
                        sort={detectedTrackersHeaderSort("LAST_DETECTED_AT", order)}
                        onSort={() => handleSort("LAST_DETECTED_AT")}
                        aria-label={t("detectedTrackersSection.sort.detectionTime")}
                      >
                        {t("detectedTrackersSection.columns.detectionTime")}
                      </TableColumnHeaderCell>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {trackers.map(tracker => (
                      <DetectedTrackerListItem
                        key={tracker.id}
                        detectedTrackerKey={tracker}
                      />
                    ))}
                  </TableBody>
                </Table>
              </div>
              {hasNext && (
                <div className={pager()}>
                  <Button
                    type="button"
                    size={2}
                    variant="soft"
                    color="neutral"
                    loading={isLoadingNext}
                    disabled={isSortPending}
                    iconStart={<CaretDownIcon />}
                    onClick={() => loadNext(DETECTED_TRACKERS_PAGE_SIZE)}
                  >
                    {t("detectedTrackersSection.actions.showMore")}
                  </Button>
                </div>
              )}
            </>
          )
        : (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                <Text size={2} color="faint">
                  {t("detectedTrackersSection.empty")}
                </Text>
              </div>
            </Card>
          )}
    </section>
  );
}
