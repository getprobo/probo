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

import { useTranslate } from "@probo/i18n";
import { Card, Tbody, Th, Thead, Tr } from "@probo/ui";
import { type ComponentProps } from "react";
import { graphql, usePaginationFragment } from "react-relay";

import type { TrackerPatternDetectedTrackersSection_trackerPattern$key } from "#/__generated__/core/TrackerPatternDetectedTrackersSection_trackerPattern.graphql";
import type {
  DetectedTrackerOrderField,
  TrackerPatternDetectedTrackersSectionRefetchQuery,
} from "#/__generated__/core/TrackerPatternDetectedTrackersSectionRefetchQuery.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";

import { DetectedTrackerRow } from "./DetectedTrackerRow";

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
    detectedTrackers(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "TrackerPatternDetectedTrackersSection_detectedTrackers", filters: ["orderBy"]) {
      __id
      edges {
        node {
          id
          ...DetectedTrackerRow_detectedTracker
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
  const { __ } = useTranslate();

  const { data, ...pagination } = usePaginationFragment<
    TrackerPatternDetectedTrackersSectionRefetchQuery,
    TrackerPatternDetectedTrackersSection_trackerPattern$key
  >(trackerPatternDetectedTrackersSectionFragment, trackerPatternKey);

  const trackers = data.detectedTrackers?.edges.map(edge => edge.node) ?? [];

  const refetchWithOrder: ComponentProps<typeof SortableTable>["refetch"] = ({ order }) => {
    pagination.refetch({
      order: { direction: order.direction, field: order.field as DetectedTrackerOrderField },
    });
  };

  return (
    <>
      <h3 className="text-lg font-semibold">{__("Detected Trackers")}</h3>

      {trackers.length > 0
        ? (
            <SortableTable
              {...pagination}
              refetch={refetchWithOrder}
              pageSize={50}
            >
              <Thead>
                <Tr>
                  <Th>{__("Identifier")}</Th>
                  <SortableTh field="INITIATOR_URL">{__("Initiator URL")}</SortableTh>
                  <Th>{__("Max Age (s)")}</Th>
                  <Th>{__("Source")}</Th>
                  <SortableTh field="LAST_DETECTED_AT">{__("Detection Time")}</SortableTh>
                </Tr>
              </Thead>
              <Tbody>
                {trackers.map(tracker => (
                  <DetectedTrackerRow
                    key={tracker.id}
                    detectedTrackerKey={tracker}
                  />
                ))}
              </Tbody>
            </SortableTable>
          )
        : (
            <Card padded>
              <div className="text-center py-12">
                <p className="text-txt-tertiary">
                  {__("No detected trackers for this pattern yet.")}
                </p>
              </div>
            </Card>
          )}
    </>
  );
}
