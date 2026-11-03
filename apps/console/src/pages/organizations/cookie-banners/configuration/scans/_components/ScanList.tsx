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

import { Card } from "@probo/ui/src/v2/Card/Card";
import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ScanList_cookieBanner$key } from "#/__generated__/core/ScanList_cookieBanner.graphql";
import type { ScanListRefetchQuery } from "#/__generated__/core/ScanListRefetchQuery.graphql";

import { cookieBannerList, trackerPatternListItem } from "../../../variants";
import { SCAN_POLL_MS } from "../_lib/scanPoll";
import { isActiveScanState } from "../_lib/scanStates";
import { useScansListFilters } from "../_lib/useScansListFilters";

import { ScanListItem } from "./ScanListItem";
import { ScansListFilters } from "./ScansListFilters";

export const scanListFragment = graphql`
  fragment ScanList_cookieBanner on CookieBanner
  @refetchable(queryName: "ScanListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 15 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "CookieBannerScanFilter", defaultValue: null }
    order: { type: "CookieBannerScanOrder", defaultValue: { field: CREATED_AT, direction: DESC } }
  ) {
    activeScan {
      id
      state
    }
    scans(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @required(action: THROW) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...ScanListItem_scan
        }
      }
    }
  }
`;

interface ScanListProps {
  cookieBannerKey: ScanList_cookieBanner$key;
  onRefetchReady: (refetch: () => void) => void;
}

export function ScanList({ cookieBannerKey, onRefetchReady }: ScanListProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const {
    graphqlFilter,
    graphqlOrder,
    graphqlPagination,
    hasStateFilter,
    direction,
    toggleDirection,
    setAfter,
    setBefore,
  } = useScansListFilters();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const prevFilterOrderRef = useRef({ graphqlFilter, graphqlOrder });
  const [cookieBanner, refetch] = useRefetchableFragment<
    ScanListRefetchQuery,
    ScanList_cookieBanner$key
  >(scanListFragment, cookieBannerKey);

  const refetchCurrentPage = useCallback(() => {
    startRefetchTransition(() => {
      refetch(
        { ...graphqlPagination, filter: graphqlFilter, order: graphqlOrder },
        { fetchPolicy: "network-only" },
      );
    });
  }, [graphqlFilter, graphqlOrder, graphqlPagination, refetch]);

  useEffect(() => {
    onRefetchReady(refetchCurrentPage);
  }, [onRefetchReady, refetchCurrentPage]);

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      prevFilterOrderRef.current = { graphqlFilter, graphqlOrder };
      return;
    }

    const filterOrderChanged = prevFilterOrderRef.current.graphqlFilter !== graphqlFilter
      || prevFilterOrderRef.current.graphqlOrder !== graphqlOrder;
    prevFilterOrderRef.current = { graphqlFilter, graphqlOrder };

    startRefetchTransition(() => {
      refetch(
        { ...graphqlPagination, filter: graphqlFilter, order: graphqlOrder },
        { fetchPolicy: filterOrderChanged ? "network-only" : "store-or-network" },
      );
    });
  }, [graphqlFilter, graphqlOrder, graphqlPagination, refetch]);

  const activeScanId = cookieBanner.activeScan != null && isActiveScanState(cookieBanner.activeScan.state)
    ? cookieBanner.activeScan.id
    : null;

  useEffect(() => {
    if (activeScanId == null) {
      return;
    }
    const interval = setInterval(() => {
      if (document.hidden) {
        return;
      }
      startRefetchTransition(() => {
        refetch(
          { ...graphqlPagination, filter: graphqlFilter, order: graphqlOrder },
          { fetchPolicy: "network-only" },
        );
      });
    }, SCAN_POLL_MS);
    return () => clearInterval(interval);
  }, [activeScanId, graphqlFilter, graphqlOrder, graphqlPagination, refetch]);

  const edges = cookieBanner.scans.edges;
  const pageInfo = cookieBanner.scans.pageInfo;
  const { root, results, pager, empty } = cookieBannerList({ pending: isRefetchPending });
  const { optional } = trackerPatternListItem();
  const requestedSort = direction === "ASC" ? "ascending" : "descending";

  return (
    <div className={root()}>
      <ScansListFilters />
      {edges.length === 0
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                {hasStateFilter
                  ? (
                      <Text size={2} color="faint">
                        {t("scansPage.emptyFiltered")}
                      </Text>
                    )
                  : (
                      <>
                        <Text size={3} weight="medium" highContrast>
                          {t("scansPage.empty.title")}
                        </Text>
                        <Text size={2} color="faint">
                          {t("scansPage.empty.description")}
                        </Text>
                      </>
                    )}
              </div>
            </Card>
          )
        : (
            <>
              <div aria-busy={isRefetchPending} className={results()}>
                <Table variant="surface" layout="fixed">
                  <TableHeader>
                    <TableRow>
                      <TableColumnHeaderCell
                        overflow="break"
                        sort={requestedSort}
                        onSort={toggleDirection}
                        aria-label={t("scansPage.sort.requested")}
                      >
                        {t("scansPage.columns.requested")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="11rem">
                        {t("scansPage.columns.state")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="11rem" className={optional()}>
                        {t("scansPage.columns.started")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="11rem" className={optional()}>
                        {t("scansPage.columns.finished")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="5.5rem">
                        {t("scansPage.columns.pages")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="5.5rem">
                        {t("scansPage.columns.cookies")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="5.5rem">
                        {t("scansPage.columns.storage")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="6rem">
                        {t("scansPage.columns.resources")}
                      </TableColumnHeaderCell>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {edges.map(({ node }) => (
                      <ScanListItem key={node.id} scanKey={node} />
                    ))}
                  </TableBody>
                </Table>
              </div>
              <div className={pager()}>
                <Pagination
                  hasPrevious={pageInfo.hasPreviousPage}
                  hasNext={pageInfo.hasNextPage}
                  previousLabel={t("scansPage.actions.previous")}
                  nextLabel={t("scansPage.actions.next")}
                  showLabels
                  disabled={isRefetchPending}
                  onPrevious={() => {
                    if (pageInfo.startCursor != null) {
                      setBefore(pageInfo.startCursor);
                    }
                  }}
                  onNext={() => {
                    if (pageInfo.endCursor != null) {
                      setAfter(pageInfo.endCursor);
                    }
                  }}
                />
              </div>
            </>
          )}
    </div>
  );
}
