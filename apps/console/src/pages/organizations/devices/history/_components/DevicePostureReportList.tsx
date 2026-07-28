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

import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { DevicePostureReportListFragment$key } from "#/__generated__/core/DevicePostureReportListFragment.graphql";
import type { DevicePostureReportListPaginationQuery } from "#/__generated__/core/DevicePostureReportListPaginationQuery.graphql";
import { SortableTable } from "#/components/SortableTable";

import { DevicePostureReportListItem } from "./DevicePostureReportListItem";

const deviceFragment = graphql`
  fragment DevicePostureReportListFragment on Device
  @refetchable(queryName: "DevicePostureReportListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    postureReports(
      first: $first
      after: $after
      last: $last
      before: $before
    ) @connection(key: "DevicePostureReportListFragment_postureReports") {
      edges {
        node {
          createdAt
          id
          ...DevicePostureReportListItemFragment
        }
      }
    }
  }
`;

interface DevicePostureReportListProps {
  fKey: DevicePostureReportListFragment$key;
}

export function DevicePostureReportList({
  fKey,
}: DevicePostureReportListProps) {
  const { t } = useTranslation();

  const reportsPagination = usePaginationFragment<
    DevicePostureReportListPaginationQuery,
    DevicePostureReportListFragment$key
  >(deviceFragment, fKey);

  const edges = reportsPagination.data.postureReports.edges;

  return (
    <SortableTable
      {...reportsPagination}
      refetch={() => {
        reportsPagination.refetch({}, { fetchPolicy: "network-only" });
      }}
      pageSize={20}
    >
      <Thead>
        <Tr>
          <Th>{t("devices.history.columns.time")}</Th>
          <Th className="text-txt-tertiary font-normal">
            {t("devices.history.columns.correlationId")}
          </Th>
          <Th>{t("devices.history.columns.checks")}</Th>
        </Tr>
      </Thead>
      <Tbody>
        {edges.length === 0
          ? (
              <Tr>
                <Td colSpan={3} className="text-center text-txt-secondary">
                  {t("devices.history.empty")}
                </Td>
              </Tr>
            )
          : (
              edges.map(({ node: report }) => (
                <DevicePostureReportListItem
                  key={report.id}
                  fKey={report}
                />
              ))
            )}
      </Tbody>
    </SortableTable>
  );
}
