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

import type { TrackerResourceList_cookieBanner$key } from "#/__generated__/core/TrackerResourceList_cookieBanner.graphql";
import type { TrackerResourceListRefetchQuery } from "#/__generated__/core/TrackerResourceListRefetchQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import { cookieBannerList } from "../../../variants";
import { useResourcesListFilters } from "../_lib/useResourcesListFilters";

import { ResourcesListFilters } from "./ResourcesListFilters";
import { TrackerResourceListItem } from "./TrackerResourceListItem";

export const RESOURCES_PAGE_SIZE = 15;

export const trackerResourceListFragment = graphql`
  fragment TrackerResourceList_cookieBanner on CookieBanner
  @refetchable(queryName: "TrackerResourceListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 15 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "TrackerResourceFilter", defaultValue: null }
  ) {
    ...MoveToCategorySelect_cookieBanner
    uncategorisedTrackerResources(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: LAST_DETECTED_AT, direction: DESC }
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
          ...TrackerResourceListItem_trackerResource
        }
      }
    }
  }
`;

interface TrackerResourceListProps {
  cookieBannerKey: TrackerResourceList_cookieBanner$key;
}

export function TrackerResourceList({ cookieBannerKey }: TrackerResourceListProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { graphqlFilter, hasActiveFilters } = useResourcesListFilters();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [cookieBanner, refetch] = useRefetchableFragment<
    TrackerResourceListRefetchQuery,
    TrackerResourceList_cookieBanner$key
  >(trackerResourceListFragment, cookieBannerKey);

  const pageVariablesRef = useRef<CursorPaginationVariables>({
    first: RESOURCES_PAGE_SIZE,
    after: null,
    last: null,
    before: null,
  });

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    pageVariablesRef.current = variables;
    refetch({ ...variables, filter: graphqlFilter }, { fetchPolicy: "store-or-network" });
  }, [graphqlFilter, refetch]);

  const { isPending: isPagePending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    cookieBanner.uncategorisedTrackerResources.pageInfo,
    RESOURCES_PAGE_SIZE,
  );

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    pageVariablesRef.current = {
      first: RESOURCES_PAGE_SIZE,
      after: null,
      last: null,
      before: null,
    };
    startRefetchTransition(() => {
      refetch(
        { ...pageVariablesRef.current, filter: graphqlFilter },
        { fetchPolicy: "store-or-network" },
      );
    });
  }, [graphqlFilter, refetch]);

  const edges = cookieBanner.uncategorisedTrackerResources.edges;
  const pageInfo = cookieBanner.uncategorisedTrackerResources.pageInfo;
  const isPending = isRefetchPending || isPagePending;
  const { root, results, pager, empty } = cookieBannerList({ pending: isPending });

  function refetchCurrentPage() {
    startRefetchTransition(() => {
      refetch(
        { ...pageVariablesRef.current, filter: graphqlFilter },
        { fetchPolicy: "network-only" },
      );
    });
  }

  function handleRemoved() {
    if (edges.length === 1 && pageInfo.hasPreviousPage) {
      goPrevious();
      return;
    }
    refetchCurrentPage();
  }

  return (
    <div className={root()}>
      <ResourcesListFilters />
      {edges.length === 0
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                {hasActiveFilters
                  ? (
                      <Text size={2} color="faint">
                        {t("resourcesPage.emptyFiltered")}
                      </Text>
                    )
                  : (
                      <>
                        <Text size={3} weight="medium" highContrast>
                          {t("resourcesPage.empty.title")}
                        </Text>
                        <Text size={2} color="faint">
                          {t("resourcesPage.empty.description")}
                        </Text>
                      </>
                    )}
              </div>
            </Card>
          )
        : (
            <>
              <div aria-busy={isPending} className={results()}>
                <Table variant="surface">
                  <TableHeader>
                    <TableRow>
                      <TableColumnHeaderCell>
                        {t("resourcesPage.columns.type")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell>
                        {t("resourcesPage.columns.origin")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell>
                        {t("resourcesPage.columns.path")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell>
                        {t("resourcesPage.columns.category")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell>
                        {t("resourcesPage.columns.lastDetected")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell />
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {edges.map(({ node }) => (
                      <TrackerResourceListItem
                        key={node.id}
                        resourceKey={node}
                        cookieBannerKey={cookieBanner}
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
                  previousLabel={t("resourcesPage.actions.previous")}
                  nextLabel={t("resourcesPage.actions.next")}
                  showLabels
                  disabled={isPending}
                  onPrevious={goPrevious}
                  onNext={goNext}
                />
              </div>
            </>
          )}
    </div>
  );
}
