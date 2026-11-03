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
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";

import type { DetectedTrackerListItem_detectedTracker$key } from "#/__generated__/core/DetectedTrackerListItem_detectedTracker.graphql";
import type { ScanDetectedTrackers_cookies$key } from "#/__generated__/core/ScanDetectedTrackers_cookies.graphql";
import type { ScanDetectedTrackers_storage$key } from "#/__generated__/core/ScanDetectedTrackers_storage.graphql";
import type { ScanDetectedTrackersCookiesPaginationQuery } from "#/__generated__/core/ScanDetectedTrackersCookiesPaginationQuery.graphql";
import type { ScanDetectedTrackersStoragePaginationQuery } from "#/__generated__/core/ScanDetectedTrackersStoragePaginationQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";
import { DetectedTrackerListItem } from "#/pages/organizations/cookie-banners/configuration/trackers/_components/DetectedTrackerListItem";

import { cookieBannerList } from "../../../variants";
import { SCAN_FINDINGS_PAGE_SIZE } from "../_lib/pageSize";

import { FindingsPagination } from "./FindingsPagination";

const scanCookiesFragment = graphql`
  fragment ScanDetectedTrackers_cookies on CookieBannerScan
  @refetchable(queryName: "ScanDetectedTrackersCookiesPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 15 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    cookies: detectedTrackers(
      first: $first
      after: $after
      last: $last
      before: $before
      filter: { trackerTypes: [COOKIE] }
      orderBy: { field: CREATED_AT, direction: DESC }
    ) {
      edges {
        node {
          id
          ...DetectedTrackerListItem_detectedTracker
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

const scanStorageFragment = graphql`
  fragment ScanDetectedTrackers_storage on CookieBannerScan
  @refetchable(queryName: "ScanDetectedTrackersStoragePaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 15 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
  ) {
    storage: detectedTrackers(
      first: $first
      after: $after
      last: $last
      before: $before
      filter: { trackerTypes: [LOCAL_STORAGE, SESSION_STORAGE] }
      orderBy: { field: CREATED_AT, direction: DESC }
    ) {
      edges {
        node {
          id
          ...DetectedTrackerListItem_detectedTracker
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

function TrackerTable<T extends DetectedTrackerListItem_detectedTracker$key & { readonly id: string }>({
  trackers,
  empty,
  hasPrevious,
  hasNext,
  loading,
  onPrevious,
  onNext,
}: {
  trackers: readonly T[];
  empty: string;
  hasPrevious: boolean;
  hasNext: boolean;
  loading: boolean;
  onPrevious: () => void;
  onNext: () => void;
}) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { results } = cookieBannerList({ pending: loading });

  if (trackers.length === 0) {
    return <Text size={2} color="faint">{empty}</Text>;
  }

  return (
    <div className="flex flex-col gap-3">
      <div aria-busy={loading} className={results()}>
        <Table variant="surface" layout="fixed">
          <TableHeader>
            <TableRow>
              <TableColumnHeaderCell overflow="truncate">
                {t("detectedTrackersSection.columns.identifier")}
              </TableColumnHeaderCell>
              <TableColumnHeaderCell overflow="truncate">
                {t("detectedTrackersSection.columns.domain")}
              </TableColumnHeaderCell>
              <TableColumnHeaderCell overflow="truncate">
                {t("detectedTrackersSection.columns.initiatorUrl")}
              </TableColumnHeaderCell>
              <TableColumnHeaderCell width="7rem">
                {t("detectedTrackersSection.columns.maxAge")}
              </TableColumnHeaderCell>
              <TableColumnHeaderCell width="7rem">
                {t("detectedTrackersSection.columns.source")}
              </TableColumnHeaderCell>
              <TableColumnHeaderCell width="8rem">
                {t("detectedTrackersSection.columns.seenBy")}
              </TableColumnHeaderCell>
              <TableColumnHeaderCell width="10rem">
                {t("detectedTrackersSection.columns.firstSeen")}
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
      <FindingsPagination
        hasPrevious={hasPrevious}
        hasNext={hasNext}
        disabled={loading}
        onPrevious={onPrevious}
        onNext={onNext}
      />
    </div>
  );
}

export function ScanCookies({ scanKey }: { scanKey: ScanDetectedTrackers_cookies$key }) {
  const { t } = useTranslation("organizations/cookie-banners");
  const [data, refetch] = useRefetchableFragment<
    ScanDetectedTrackersCookiesPaginationQuery,
    ScanDetectedTrackers_cookies$key
  >(scanCookiesFragment, scanKey);
  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);
  const { isPending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    data.cookies.pageInfo,
    SCAN_FINDINGS_PAGE_SIZE,
  );

  return (
    <TrackerTable
      trackers={data.cookies.edges.map(edge => edge.node)}
      empty={t("scanPage.findings.cookiesEmpty")}
      hasPrevious={data.cookies.pageInfo.hasPreviousPage}
      hasNext={data.cookies.pageInfo.hasNextPage}
      loading={isPending}
      onPrevious={goPrevious}
      onNext={goNext}
    />
  );
}

export function ScanStorage({ scanKey }: { scanKey: ScanDetectedTrackers_storage$key }) {
  const { t } = useTranslation("organizations/cookie-banners");
  const [data, refetch] = useRefetchableFragment<
    ScanDetectedTrackersStoragePaginationQuery,
    ScanDetectedTrackers_storage$key
  >(scanStorageFragment, scanKey);
  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);
  const { isPending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    data.storage.pageInfo,
    SCAN_FINDINGS_PAGE_SIZE,
  );

  return (
    <TrackerTable
      trackers={data.storage.edges.map(edge => edge.node)}
      empty={t("scanPage.findings.storageEmpty")}
      hasPrevious={data.storage.pageInfo.hasPreviousPage}
      hasNext={data.storage.pageInfo.hasNextPage}
      loading={isPending}
      onPrevious={goPrevious}
      onNext={goNext}
    />
  );
}
