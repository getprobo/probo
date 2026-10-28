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

import { PlusIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { type ReactNode, useCallback, useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { InternalControlList_organization$key } from "#/__generated__/core/InternalControlList_organization.graphql";
import type { InternalControlListRefetchQuery } from "#/__generated__/core/InternalControlListRefetchQuery.graphql";
import InternalControlFormDialog from "#/pages/organizations/internal-controls/dialog/InternalControlFormDialog";

import { useInternalControlsListFilters } from "../../_lib/useInternalControlsListFilters";
import { internalControlsPage } from "../variants";

import { InternalControlListItem } from "./InternalControlListItem";
import { InternalControlsEmpty } from "./InternalControlsEmpty";

export const internalControlListFragment = graphql`
  fragment InternalControlList_organization on Organization
  @refetchable(queryName: "InternalControlListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    order: { type: "InternalControlOrder", defaultValue: { direction: ASC, field: NAME } }
    query: { type: "String", defaultValue: null }
    implementationStatuses: { type: "[InternalControlImplementationStatus!]", defaultValue: null }
    category: { type: "String", defaultValue: null }
    ownerId: { type: "ID", defaultValue: null }
    needsAttention: { type: "Boolean", defaultValue: null }
  ) {
    canCreateInternalControl: permission(action: "core:internal-control:create")
    internalControls(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: {
        query: $query
        implementationStatuses: $implementationStatuses
        category: $category
        ownerId: $ownerId
        needsAttention: $needsAttention
      }
    ) @required(action: THROW) {
      totalCount
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...InternalControlListItem_internalControl
        }
      }
    }
  }
`;

interface InternalControlListProps {
  organizationKey: InternalControlList_organization$key;
  importAction?: ReactNode;
  onRefetch: (refetch: () => void) => void;
  onTotalCount: (count: number) => void;
}

export function InternalControlList(props: InternalControlListProps) {
  const { organizationKey, importAction, onRefetch, onTotalCount } = props;
  const { t } = useTranslation("organizations/internal-controls");
  const filters = useInternalControlsListFilters();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const variablesRef = useRef(filters.variables);
  const variablesKey = JSON.stringify(filters.variables);
  const [organization, refetch] = useRefetchableFragment<
    InternalControlListRefetchQuery,
    InternalControlList_organization$key
  >(internalControlListFragment, organizationKey);

  useEffect(() => {
    variablesRef.current = filters.variables;
  }, [filters.variables]);

  const refetchCurrent = useCallback(() => {
    startRefetchTransition(() => {
      refetch(variablesRef.current, { fetchPolicy: "store-and-network" });
    });
  }, [refetch]);

  useEffect(() => {
    onRefetch(refetchCurrent);
  }, [onRefetch, refetchCurrent]);

  useEffect(() => {
    onTotalCount(organization.internalControls.totalCount);
  }, [onTotalCount, organization.internalControls.totalCount]);

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    startRefetchTransition(() => {
      refetch(variablesRef.current, { fetchPolicy: "store-and-network" });
    });
  }, [variablesKey, refetch]);

  const edges = organization.internalControls.edges;
  const pageInfo = organization.internalControls.pageInfo;
  const isPending = isRefetchPending;
  const { list, pager } = internalControlsPage({ pending: isPending });
  const canCreate = organization.canCreateInternalControl ?? false;

  function goNext() {
    if (!pageInfo.hasNextPage || pageInfo.endCursor == null) {
      return;
    }
    filters.setPageCursor({ after: pageInfo.endCursor, page: filters.page + 1 });
  }

  function goPrevious() {
    if (!pageInfo.hasPreviousPage || pageInfo.startCursor == null) {
      return;
    }
    filters.setPageCursor({
      before: pageInfo.startCursor,
      page: Math.max(1, filters.page - 1),
    });
  }

  function handleChanged() {
    if (edges.length === 1 && pageInfo.hasPreviousPage && pageInfo.startCursor != null) {
      filters.setPageCursor({
        before: pageInfo.startCursor,
        page: Math.max(1, filters.page - 1),
      });
      return;
    }
    refetchCurrent();
  }

  function sortState(field: "name" | "nextDue") {
    if (filters.sort !== field) {
      return "none" as const;
    }
    return filters.direction === "ASC" ? "ascending" as const : "descending" as const;
  }

  const create = canCreate
    ? (
        <InternalControlFormDialog onSaved={refetchCurrent}>
          <Button variant="solid" iconStart={<PlusIcon />}>
            {t("list.actions.new")}
          </Button>
        </InternalControlFormDialog>
      )
    : null;

  return (
    <div aria-busy={isPending} className={list()}>
      <Table size={2} variant="surface">
        <TableHeader>
          <TableRow align="center">
            <TableColumnHeaderCell
              sort={sortState("name")}
              onSort={() => filters.setSort("name")}
            >
              {t("list.columns.control")}
            </TableColumnHeaderCell>
            <TableColumnHeaderCell>{t("list.columns.status")}</TableColumnHeaderCell>
            <TableColumnHeaderCell>{t("list.columns.owner")}</TableColumnHeaderCell>
            <TableColumnHeaderCell
              sort={sortState("nextDue")}
              onSort={() => filters.setSort("nextDue")}
            >
              {t("list.columns.nextDue")}
            </TableColumnHeaderCell>
            <TableColumnHeaderCell width="3rem" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {edges.length === 0
            ? (
                <InternalControlsEmpty
                  filtered={filters.hasActiveFilters || organization.internalControls.totalCount > 0}
                  canCreate={canCreate}
                  create={(
                    <>
                      {importAction}
                      {create}
                    </>
                  )}
                  onClear={filters.clear}
                />
              )
            : edges.map(({ node }) => (
                <InternalControlListItem
                  key={node.id}
                  internalControlKey={node}
                  onChanged={handleChanged}
                />
              ))}
        </TableBody>
      </Table>
      <div className={pager()}>
        <Pagination
          hasPrevious={pageInfo.hasPreviousPage}
          hasNext={pageInfo.hasNextPage}
          label={t("list.page", { page: filters.page })}
          previousLabel={t("list.actions.previous")}
          nextLabel={t("list.actions.next")}
          showLabels
          disabled={isPending}
          onPrevious={goPrevious}
          onNext={goNext}
        />
      </div>
    </div>
  );
}
