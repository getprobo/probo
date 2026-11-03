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

import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableCell } from "@probo/ui/src/v2/Table/TableCell";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";

import type { ScanResources_scan$key } from "#/__generated__/core/ScanResources_scan.graphql";
import type { ScanResourcesPaginationQuery } from "#/__generated__/core/ScanResourcesPaginationQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import { cookieBannerList } from "../../../variants";
import { resourceTypeBadges } from "../../resources/_lib/resourceBadges";
import { SCAN_FINDINGS_PAGE_SIZE } from "../_lib/pageSize";

import { FindingsPagination } from "./FindingsPagination";

const scanResourcesFragment = graphql`
  fragment ScanResources_scan on CookieBannerScan
  @refetchable(queryName: "ScanResourcesPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 15 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    trackerResources(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: CREATED_AT, direction: DESC }
    ) {
      edges {
        node {
          id
          type
          origin
          path
        }
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
    }
  }
`;

export function ScanResources({ scanKey }: { scanKey: ScanResources_scan$key }) {
  const { t } = useTranslation("organizations/cookie-banners");
  const [data, refetch] = useRefetchableFragment<
    ScanResourcesPaginationQuery,
    ScanResources_scan$key
  >(scanResourcesFragment, scanKey);
  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);
  const { isPending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    data.trackerResources.pageInfo,
    SCAN_FINDINGS_PAGE_SIZE,
  );
  const { results } = cookieBannerList({ pending: isPending });
  const resources = data.trackerResources.edges.map(edge => edge.node);

  if (resources.length === 0) {
    return <Text size={2} color="faint">{t("scanPage.findings.resourcesEmpty")}</Text>;
  }

  return (
    <div className="flex flex-col gap-3">
      <div aria-busy={isPending} className={results()}>
        <Table variant="surface" layout="fixed">
          <TableHeader>
            <TableRow>
              <TableColumnHeaderCell width="9rem">
                {t("resourcesPage.columns.type")}
              </TableColumnHeaderCell>
              <TableColumnHeaderCell overflow="truncate">
                {t("resourcesPage.columns.origin")}
              </TableColumnHeaderCell>
              <TableColumnHeaderCell overflow="truncate">
                {t("resourcesPage.columns.path")}
              </TableColumnHeaderCell>
            </TableRow>
          </TableHeader>
          <TableBody>
            {resources.map(resource => (
              <TableRow key={resource.id}>
                <TableCell>
                  <Text size={2}>{t(`resourcesPage.types.${resourceTypeBadges[resource.type].labelKey}`)}</Text>
                </TableCell>
                <TableCell overflow="truncate">
                  <Text size={2} className="break-all font-mono">{resource.origin}</Text>
                </TableCell>
                <TableCell overflow="truncate">
                  <Text size={2} className="break-all font-mono">{resource.path}</Text>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <FindingsPagination
        hasPrevious={data.trackerResources.pageInfo.hasPreviousPage}
        hasNext={data.trackerResources.pageInfo.hasNextPage}
        disabled={isPending}
        onPrevious={goPrevious}
        onNext={goNext}
      />
    </div>
  );
}
