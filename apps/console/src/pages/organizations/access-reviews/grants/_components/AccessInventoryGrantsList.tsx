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
import { useCallback, useEffect, useMemo, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessInventoryGrantsList_organization$key } from "#/__generated__/core/AccessInventoryGrantsList_organization.graphql";
import type { AccessInventoryGrantsListRefetchQuery } from "#/__generated__/core/AccessInventoryGrantsListRefetchQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import { ACCESS_INVENTORY_GRANTS_PAGE_SIZE } from "../_lib/pageSize";
import {
  accessInventoryGrantsFilterKey,
  useAccessInventoryGrantsFilters,
} from "../_lib/useAccessInventoryGrantsFilters";

import { AccessInventoryGrantListItem } from "./AccessInventoryGrantListItem";
import { AccessInventoryGrantsFilters } from "./AccessInventoryGrantsFilters";
import { accessInventoryGrantsList } from "./variants";

export const accessInventoryGrantsListFragment = graphql`
  fragment AccessInventoryGrantsList_organization on Organization
  @refetchable(queryName: "AccessInventoryGrantsListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 25 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "AccessInventoryGrantFilter", defaultValue: null }
  ) {
    accessReviewSources(first: 100, orderBy: { direction: ASC, field: CREATED_AT }) {
      edges {
        node {
          id
          name
          connectorAccount {
            name
          }
        }
      }
    }
    accessInventoryGrants(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { direction: DESC, field: CREATED_AT }
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
          ...AccessInventoryGrantListItem_grant
        }
      }
    }
  }
`;

interface AccessInventoryGrantsListProps {
  organizationKey: AccessInventoryGrantsList_organization$key;
  loadedFilterKey: string;
}

export function AccessInventoryGrantsList({
  organizationKey,
  loadedFilterKey,
}: AccessInventoryGrantsListProps) {
  const { t } = useTranslation("organizations/access-reviews");
  const { graphqlFilter, hasActiveFilters } = useAccessInventoryGrantsFilters();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [organization, refetch] = useRefetchableFragment<
    AccessInventoryGrantsListRefetchQuery,
    AccessInventoryGrantsList_organization$key
  >(accessInventoryGrantsListFragment, organizationKey);

  const pageVariablesRef = useRef<CursorPaginationVariables>({
    first: ACCESS_INVENTORY_GRANTS_PAGE_SIZE,
    after: null,
    last: null,
    before: null,
  });

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    pageVariablesRef.current = variables;
    refetch(
      { ...variables, filter: graphqlFilter },
      { fetchPolicy: "store-or-network" },
    );
  }, [graphqlFilter, refetch]);

  const { isPending: isPagePending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    organization.accessInventoryGrants.pageInfo,
    ACCESS_INVENTORY_GRANTS_PAGE_SIZE,
  );

  useEffect(() => {
    const matchesLoaded
      = accessInventoryGrantsFilterKey(graphqlFilter) === loadedFilterKey;
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      if (matchesLoaded) {
        return;
      }
    }

    pageVariablesRef.current = {
      first: ACCESS_INVENTORY_GRANTS_PAGE_SIZE,
      after: null,
      last: null,
      before: null,
    };
    startRefetchTransition(() => {
      refetch(
        { ...pageVariablesRef.current, filter: graphqlFilter },
        { fetchPolicy: "network-only" },
      );
    });
  }, [graphqlFilter, loadedFilterKey, refetch]);

  const sources = useMemo(
    () => organization.accessReviewSources.edges.map(({ node }) => ({
      id: node.id,
      name: node.name,
      accountName: node.connectorAccount?.name,
    })),
    [organization.accessReviewSources.edges],
  );

  const edges = organization.accessInventoryGrants.edges;
  const pageInfo = organization.accessInventoryGrants.pageInfo;
  const isPending = isRefetchPending || isPagePending;
  const { root, results, empty, pager } = accessInventoryGrantsList({
    pending: isPending,
  });
  const showPager = edges.length > 0
    || pageInfo.hasPreviousPage
    || pageInfo.hasNextPage;

  return (
    <div className={root()}>
      <AccessInventoryGrantsFilters sources={sources} />
      {edges.length === 0
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                <Text size={2} color="faint">
                  {t(hasActiveFilters ? "grants.emptyFiltered" : "grants.empty")}
                </Text>
              </div>
            </Card>
          )
        : (
            <div aria-busy={isPending} className={results()}>
              <Table variant="surface" layout="fixed">
                <TableHeader>
                  <TableRow>
                    <TableColumnHeaderCell overflow="truncate">
                      {t("grants.columns.principal")}
                    </TableColumnHeaderCell>
                    <TableColumnHeaderCell overflow="truncate">
                      {t("grants.columns.resource")}
                    </TableColumnHeaderCell>
                    <TableColumnHeaderCell overflow="truncate">
                      {t("grants.columns.entitlement")}
                    </TableColumnHeaderCell>
                    <TableColumnHeaderCell width="10rem" overflow="truncate">
                      {t("grants.columns.source")}
                    </TableColumnHeaderCell>
                    <TableColumnHeaderCell width="8rem">
                      {t("grants.columns.presence")}
                    </TableColumnHeaderCell>
                    <TableColumnHeaderCell width="10rem">
                      {t("grants.columns.lastSeen")}
                    </TableColumnHeaderCell>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {edges.map(({ node }) => (
                    <AccessInventoryGrantListItem
                      key={node.id}
                      grantKey={node}
                    />
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
      {showPager && (
        <div className={pager()}>
          <Pagination
            hasPrevious={pageInfo.hasPreviousPage}
            hasNext={pageInfo.hasNextPage}
            previousLabel={t("grants.actions.previous")}
            nextLabel={t("grants.actions.next")}
            showLabels
            disabled={isPending}
            onPrevious={goPrevious}
            onNext={goNext}
          />
        </div>
      )}
    </div>
  );
}
