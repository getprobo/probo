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

import type { AccessInventoryChangeEventsList_organization$key } from "#/__generated__/core/AccessInventoryChangeEventsList_organization.graphql";
import type { AccessInventoryChangeEventsListRefetchQuery } from "#/__generated__/core/AccessInventoryChangeEventsListRefetchQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import { ACCESS_INVENTORY_CHANGE_EVENTS_PAGE_SIZE } from "../_lib/pageSize";
import { useAccessInventoryChangeEventsFilters } from "../_lib/useAccessInventoryChangeEventsFilters";

import { AccessInventoryChangeEventListItem } from "./AccessInventoryChangeEventListItem";
import { AccessInventoryChangeEventsFilters } from "./AccessInventoryChangeEventsFilters";
import { accessInventoryChangeEventsList } from "./variants";

export const accessInventoryChangeEventsListFragment = graphql`
  fragment AccessInventoryChangeEventsList_organization on Organization
  @refetchable(queryName: "AccessInventoryChangeEventsListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 25 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "AccessInventoryChangeEventFilter", defaultValue: null }
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
    accessInventoryChangeEvents(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { direction: DESC, field: EFFECTIVE_AT }
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
          ...AccessInventoryChangeEventListItem_event
        }
      }
    }
  }
`;

interface AccessInventoryChangeEventsListProps {
  organizationKey: AccessInventoryChangeEventsList_organization$key;
}

export function AccessInventoryChangeEventsList({
  organizationKey,
}: AccessInventoryChangeEventsListProps) {
  const { t } = useTranslation("organizations/access-reviews");
  const { graphqlFilter, hasActiveFilters }
    = useAccessInventoryChangeEventsFilters();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [organization, refetch] = useRefetchableFragment<
    AccessInventoryChangeEventsListRefetchQuery,
    AccessInventoryChangeEventsList_organization$key
  >(accessInventoryChangeEventsListFragment, organizationKey);

  const pageVariablesRef = useRef<CursorPaginationVariables>({
    first: ACCESS_INVENTORY_CHANGE_EVENTS_PAGE_SIZE,
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
    organization.accessInventoryChangeEvents.pageInfo,
    ACCESS_INVENTORY_CHANGE_EVENTS_PAGE_SIZE,
  );

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    pageVariablesRef.current = {
      first: ACCESS_INVENTORY_CHANGE_EVENTS_PAGE_SIZE,
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

  const sources = useMemo(
    () => organization.accessReviewSources.edges.map(({ node }) => ({
      id: node.id,
      name: node.name,
      accountName: node.connectorAccount?.name,
    })),
    [organization.accessReviewSources.edges],
  );

  const edges = organization.accessInventoryChangeEvents.edges;
  const pageInfo = organization.accessInventoryChangeEvents.pageInfo;
  const isPending = isRefetchPending || isPagePending;
  const { root, results, empty, pager } = accessInventoryChangeEventsList({
    pending: isPending,
  });

  return (
    <div className={root()}>
      <AccessInventoryChangeEventsFilters sources={sources} />
      {edges.length === 0
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                <Text size={2} color="faint">
                  {t(hasActiveFilters ? "changes.emptyFiltered" : "changes.empty")}
                </Text>
              </div>
            </Card>
          )
        : (
            <>
              <div aria-busy={isPending} className={results()}>
                <Table variant="surface" layout="fixed">
                  <TableHeader>
                    <TableRow>
                      <TableColumnHeaderCell width="9rem">
                        {t("changes.columns.changeType")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell overflow="truncate">
                        {t("changes.columns.principal")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell overflow="truncate">
                        {t("changes.columns.resource")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell overflow="truncate">
                        {t("changes.columns.entitlement")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="10rem" overflow="truncate">
                        {t("changes.columns.source")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="10rem">
                        {t("changes.columns.effectiveAt")}
                      </TableColumnHeaderCell>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {edges.map(({ node }) => (
                      <AccessInventoryChangeEventListItem
                        key={node.id}
                        eventKey={node}
                      />
                    ))}
                  </TableBody>
                </Table>
              </div>
              <div className={pager()}>
                <Pagination
                  hasPrevious={pageInfo.hasPreviousPage}
                  hasNext={pageInfo.hasNextPage}
                  previousLabel={t("changes.actions.previous")}
                  nextLabel={t("changes.actions.next")}
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
