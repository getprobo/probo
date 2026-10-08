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
import { useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { TrackerPatternList_cookieBanner$key } from "#/__generated__/core/TrackerPatternList_cookieBanner.graphql";
import type { TrackerPatternListRefetchQuery } from "#/__generated__/core/TrackerPatternListRefetchQuery.graphql";

import { cookieBannerList } from "../../../variants";
import {
  trackersListHeaderSort,
  useTrackersListFilters,
} from "../_lib/useTrackersListFilters";

import { TrackerPatternListItem } from "./TrackerPatternListItem";
import { TrackersListFilters } from "./TrackersListFilters";

export const trackerPatternListFragment = graphql`
  fragment TrackerPatternList_cookieBanner on CookieBanner
  @refetchable(queryName: "TrackerPatternListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 15 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "TrackerPatternFilter", defaultValue: null }
    order: { type: "TrackerPatternOrder", defaultValue: { field: NAME, direction: ASC } }
  ) {
    ...TrackersListFilters_cookieBanner
    ...MoveToCategorySelect_cookieBanner
    trackerPatterns(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    )
      @required(action: THROW) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...TrackerPatternListItem_trackerPattern
        }
      }
    }
  }
`;

interface TrackerPatternListProps {
  cookieBannerKey: TrackerPatternList_cookieBanner$key;
}

export function TrackerPatternList({ cookieBannerKey }: TrackerPatternListProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const {
    view,
    graphqlFilter,
    graphqlOrder,
    graphqlPagination,
    hasActiveFilters,
    setOrder,
    setAfter,
    setBefore,
  } = useTrackersListFilters();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [cookieBanner, refetch] = useRefetchableFragment<
    TrackerPatternListRefetchQuery,
    TrackerPatternList_cookieBanner$key
  >(trackerPatternListFragment, cookieBannerKey);

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    startRefetchTransition(() => {
      refetch(
        { ...graphqlPagination, filter: graphqlFilter, order: graphqlOrder },
        { fetchPolicy: "network-only" },
      );
    });
  }, [graphqlFilter, graphqlOrder, graphqlPagination, refetch]);

  const edges = cookieBanner.trackerPatterns.edges;
  const pageInfo = cookieBanner.trackerPatterns.pageInfo;
  const { root, results, pager, empty } = cookieBannerList({ pending: isRefetchPending });

  function refetchCurrentPage() {
    startRefetchTransition(() => {
      refetch(
        { ...graphqlPagination, filter: graphqlFilter, order: graphqlOrder },
        { fetchPolicy: "network-only" },
      );
    });
  }

  function handleRemoved() {
    if (edges.length === 1 && pageInfo.hasPreviousPage && pageInfo.startCursor != null) {
      setBefore(pageInfo.startCursor);
      return;
    }
    refetchCurrentPage();
  }

  return (
    <div className={root()}>
      <TrackersListFilters cookieBannerKey={cookieBanner} />
      {edges.length === 0
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                {hasActiveFilters
                  ? (
                      <Text size={2} color="faint">
                        {t("trackersPage.emptyFiltered")}
                      </Text>
                    )
                  : view === "on-banner"
                    ? (
                        <Text size={2} color="faint">
                          {t("trackersPage.empty.onBanner")}
                        </Text>
                      )
                    : (
                        <>
                          <Text size={3} weight="medium" highContrast>
                            {t("trackersPage.empty.title")}
                          </Text>
                          <Text size={2} color="faint">
                            {t("trackersPage.empty.description")}
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
                        sort={trackersListHeaderSort("NAME", graphqlOrder)}
                        onSort={() => setOrder("NAME")}
                        aria-label={t("trackersPage.sort.name")}
                      >
                        {t("trackersPage.columns.name")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="10rem">
                        {t("trackersPage.columns.attribution")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell
                        width="7rem"
                        sort={trackersListHeaderSort("SOURCE", graphqlOrder)}
                        onSort={() => setOrder("SOURCE")}
                        aria-label={t("trackersPage.sort.source")}
                      >
                        {t("trackersPage.columns.source")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="10rem">
                        {t("trackersPage.columns.category")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="7rem">
                        {t("trackersPage.columns.maxAge")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell
                        width="10rem"
                        sort={trackersListHeaderSort("LAST_MATCHED_AT", graphqlOrder)}
                        onSort={() => setOrder("LAST_MATCHED_AT")}
                        aria-label={t("trackersPage.sort.lastMatched")}
                      >
                        {t("trackersPage.columns.lastMatched")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="3rem" />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {edges.map(({ node }) => (
                      <TrackerPatternListItem
                        key={node.id}
                        patternKey={node}
                        cookieBannerKey={cookieBanner}
                        onMoved={refetchCurrentPage}
                        onRemoved={handleRemoved}
                      />
                    ))}
                  </TableBody>
                </Table>
              </div>
              <div className={pager()}>
                <Pagination
                  hasPrevious={pageInfo.hasPreviousPage}
                  hasNext={pageInfo.hasNextPage}
                  previousLabel={t("trackersPage.actions.previous")}
                  nextLabel={t("trackersPage.actions.next")}
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
