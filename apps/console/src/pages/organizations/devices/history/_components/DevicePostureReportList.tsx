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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";

import type { DevicePostureReportListFragment$key } from "#/__generated__/core/DevicePostureReportListFragment.graphql";
import type { DevicePostureReportListPaginationQuery } from "#/__generated__/core/DevicePostureReportListPaginationQuery.graphql";

import { devicesPage } from "../../_components/variants";

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
  const { t } = useTranslation("organizations/devices");
  const { more } = devicesPage();
  const reportsPagination = usePaginationFragment<
    DevicePostureReportListPaginationQuery,
    DevicePostureReportListFragment$key
  >(deviceFragment, fKey);
  const edges = reportsPagination.data.postureReports.edges;

  return (
    <div className={devicesPage().root()}>
      <Table size={2} variant="surface">
        <TableHeader>
          <TableRow>
            <TableColumnHeaderCell>{t("devices.history.columns.time")}</TableColumnHeaderCell>
            <TableColumnHeaderCell>{t("devices.history.columns.correlationId")}</TableColumnHeaderCell>
            <TableColumnHeaderCell>{t("devices.history.columns.checks")}</TableColumnHeaderCell>
          </TableRow>
        </TableHeader>
        <TableBody>
          {edges.length === 0
            ? (
                <TableRow>
                  <TableCell colSpan={3} justify="center">
                    <Text size={2} color="faint">{t("devices.history.empty")}</Text>
                  </TableCell>
                </TableRow>
              )
            : edges.map(({ node: report }) => (
                <DevicePostureReportListItem key={report.id} fKey={report} />
              ))}
        </TableBody>
      </Table>
      {reportsPagination.hasNext && (
        <div className={more()}>
          <Button
            variant="soft"
            color="neutral"
            loading={reportsPagination.isLoadingNext}
            onClick={() => reportsPagination.loadNext(20)}
          >
            {t("sortableTable.actions.showMore")}
          </Button>
        </div>
      )}
    </div>
  );
}
