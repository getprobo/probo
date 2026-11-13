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

import type { FindingList_organization$key } from "#/__generated__/core/FindingList_organization.graphql";
import type { FindingListRefetchQuery } from "#/__generated__/core/FindingListRefetchQuery.graphql";

import {
  type FindingsListVariables,
  findingsListVariablesKey,
  useFindingsListFilters,
} from "../../_lib/useFindingsListFilters";
import { findingColumnWidth, findingsPage } from "../variants";

import { FindingListItem } from "./FindingListItem";
import { FindingsEmpty } from "./FindingsEmpty";

export const findingListFragment = graphql`
  fragment FindingList_organization on Organization
  @refetchable(queryName: "FindingListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    kind: { type: "FindingKind", defaultValue: null }
    status: { type: "FindingStatus", defaultValue: null }
    priority: { type: "FindingPriority", defaultValue: null }
    ownerId: { type: "ID", defaultValue: null }
    auditId: { type: "ID", defaultValue: null }
  ) {
    canCreateFinding: permission(action: "core:finding:create")
    findings(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: CREATED_AT, direction: DESC }
      filter: {
        kind: $kind
        status: $status
        priority: $priority
        ownerId: $ownerId
        auditId: $auditId
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
          ...FindingListItem_finding
        }
      }
    }
  }
`;

interface FindingListProps {
  organizationKey: FindingList_organization$key;
  loadedVariables: Partial<FindingsListVariables>;
  create: ReactNode;
  onTotalCount: (count: number) => void;
}

export function FindingList(props: FindingListProps) {
  const { organizationKey, loadedVariables, create, onTotalCount } = props;
  const { t } = useTranslation("organizations/findings");
  const filters = useFindingsListFilters();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const fetchedKeyRef = useRef(findingsListVariablesKey(loadedVariables));
  const variablesRef = useRef(filters.variables);
  const variablesKey = findingsListVariablesKey(filters.variables);
  const listFiltersKey = JSON.stringify({
    ...filters.variables,
    first: null,
    after: null,
    last: null,
    before: null,
  });
  const listFiltersKeyRef = useRef(listFiltersKey);
  const [organization, refetch] = useRefetchableFragment<
    FindingListRefetchQuery,
    FindingList_organization$key
  >(findingListFragment, organizationKey);

  useEffect(() => {
    variablesRef.current = filters.variables;
  }, [filters.variables]);

  const refetchCurrent = useCallback(() => {
    startRefetchTransition(() => {
      refetch(variablesRef.current, { fetchPolicy: "network-only" });
    });
  }, [refetch]);

  useEffect(() => {
    onTotalCount(organization.findings.totalCount);
  }, [onTotalCount, organization.findings.totalCount]);

  useEffect(() => {
    const filtersChanged = listFiltersKeyRef.current !== listFiltersKey;
    listFiltersKeyRef.current = listFiltersKey;

    if (fetchedKeyRef.current === variablesKey) {
      return;
    }
    fetchedKeyRef.current = variablesKey;

    startRefetchTransition(() => {
      refetch(variablesRef.current, {
        fetchPolicy: filtersChanged ? "network-only" : "store-or-network",
      });
    });
  }, [variablesKey, listFiltersKey, refetch]);

  const edges = organization.findings.edges;
  const pageInfo = organization.findings.pageInfo;
  const { list, pager, nameCell } = findingsPage({ pending: isRefetchPending });
  const canCreate = organization.canCreateFinding ?? false;

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

  function handleDeleted() {
    if (edges.length === 1 && pageInfo.hasPreviousPage && pageInfo.startCursor != null) {
      filters.setPageCursor({
        before: pageInfo.startCursor,
        page: Math.max(1, filters.page - 1),
      });
      return;
    }
    refetchCurrent();
  }

  const createAction = canCreate
    ? create
    : null;

  return (
    <div aria-busy={isRefetchPending} className={list()}>
      <Table size={2} variant="surface" layout="fixed">
        <TableHeader>
          <TableRow align="center">
            <TableColumnHeaderCell width={findingColumnWidth.reference}>
              {t("list.columns.reference")}
            </TableColumnHeaderCell>
            <TableColumnHeaderCell className={nameCell()} minWidth={findingColumnWidth.title}>
              {t("list.columns.title")}
            </TableColumnHeaderCell>
            <TableColumnHeaderCell width={findingColumnWidth.kind}>
              {t("list.columns.kind")}
            </TableColumnHeaderCell>
            <TableColumnHeaderCell width={findingColumnWidth.status}>
              {t("list.columns.status")}
            </TableColumnHeaderCell>
            <TableColumnHeaderCell width={findingColumnWidth.priority}>
              {t("list.columns.priority")}
            </TableColumnHeaderCell>
            <TableColumnHeaderCell width={findingColumnWidth.owner}>
              {t("list.columns.owner")}
            </TableColumnHeaderCell>
            <TableColumnHeaderCell width={findingColumnWidth.due}>
              {t("list.columns.due")}
            </TableColumnHeaderCell>
            <TableColumnHeaderCell width={findingColumnWidth.actions} />
          </TableRow>
        </TableHeader>
        <TableBody>
          {edges.length === 0
            ? (
                <FindingsEmpty
                  filtered={filters.hasActiveFilters || organization.findings.totalCount > 0}
                  canCreate={canCreate}
                  create={createAction ?? (
                    <Button variant="solid" iconStart={<PlusIcon />} disabled>
                      {t("list.actions.new")}
                    </Button>
                  )}
                  onClear={filters.clear}
                />
              )
            : edges.map(({ node }) => (
                <FindingListItem
                  key={node.id}
                  findingKey={node}
                  onDeleted={handleDeleted}
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
          disabled={isRefetchPending}
          onPrevious={goPrevious}
          onNext={goNext}
        />
      </div>
    </div>
  );
}
