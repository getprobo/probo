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
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { AuditLogEntryDialog_entry$key } from "#/__generated__/iam/AuditLogEntryDialog_entry.graphql";
import type { AuditLogList_organization$key } from "#/__generated__/iam/AuditLogList_organization.graphql";
import type { AuditLogListRefetchQuery } from "#/__generated__/iam/AuditLogListRefetchQuery.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import { useAuditLogListFilters } from "../_lib/useAuditLogListFilters";
import { auditLogList, auditLogPage } from "../variants";

import { AuditLogEntryDialog } from "./AuditLogEntryDialog";
import { AuditLogListFilters } from "./AuditLogListFilters";
import { AuditLogListItem } from "./AuditLogListItem";
import { ExportAuditLogDialog } from "./ExportAuditLogDialog";

const AUDIT_LOG_NS = "iam/organizations/audit-log";

export const AUDIT_LOG_PAGE_SIZE = 25;

export const auditLogListFragment = graphql`
  fragment AuditLogList_organization on Organization
  @refetchable(queryName: "AuditLogListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 25 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "AuditLogEntryFilter", defaultValue: null }
  ) {
    id
    canExportAuditLog: permission(action: "iam:audit-log:export")
    auditLogEntries(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: CREATED_AT, direction: DESC }
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
          ...AuditLogListItem_entry
          ...AuditLogEntryDialog_entry
        }
      }
    }
  }
`;

interface AuditLogListProps {
  organizationKey: AuditLogList_organization$key;
}

export function AuditLogList({ organizationKey }: AuditLogListProps) {
  const { t } = useTranslation(AUDIT_LOG_NS);
  const { graphqlFilter, hasActiveFilters } = useAuditLogListFilters();
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [selected, setSelected] = useState<AuditLogEntryDialog_entry$key | null>(null);
  const [entryOpen, setEntryOpen] = useState(false);
  const [organization, refetch] = useRefetchableFragment<
    AuditLogListRefetchQuery,
    AuditLogList_organization$key
  >(auditLogListFragment, organizationKey);

  const pageVariablesRef = useRef<CursorPaginationVariables>({
    first: AUDIT_LOG_PAGE_SIZE,
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
    organization.auditLogEntries.pageInfo,
    AUDIT_LOG_PAGE_SIZE,
  );

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    pageVariablesRef.current = {
      first: AUDIT_LOG_PAGE_SIZE,
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

  const edges = organization.auditLogEntries.edges;
  const pageInfo = organization.auditLogEntries.pageInfo;
  const isPending = isRefetchPending || isPagePending;
  const { header, intro } = auditLogPage();
  const { root, results, empty, pager } = auditLogList({ pending: isPending });

  return (
    <>
      <div className={header()}>
        <div className={intro()}>
          <Heading level={1} size={6} weight="medium" highContrast>
            {t("auditLogPage.title")}
          </Heading>
          <Text size={2}>
            {t("auditLogPage.description")}
          </Text>
        </div>
        {organization.canExportAuditLog && (
          <ExportAuditLogDialog organizationId={organization.id} />
        )}
      </div>
      <div className={root()}>
        <AuditLogListFilters />
        {edges.length === 0
          ? (
              <Card variant="soft" size={2}>
                <div className={empty()}>
                  <Text size={2} color="faint">
                    {t(hasActiveFilters ? "auditLogPage.emptyFiltered" : "auditLogPage.empty")}
                  </Text>
                </div>
              </Card>
            )
          : (
              <>
                <div aria-busy={isPending} className={results()}>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableColumnHeaderCell>
                          {t("auditLogPage.columns.when")}
                        </TableColumnHeaderCell>
                        <TableColumnHeaderCell>
                          {t("auditLogPage.columns.actor")}
                        </TableColumnHeaderCell>
                        <TableColumnHeaderCell>
                          {t("auditLogPage.columns.action")}
                        </TableColumnHeaderCell>
                        <TableColumnHeaderCell>
                          {t("auditLogPage.columns.target")}
                        </TableColumnHeaderCell>
                        <TableColumnHeaderCell>
                          {t("auditLogPage.columns.outcome")}
                        </TableColumnHeaderCell>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {edges.map(({ node }) => (
                        <AuditLogListItem
                          key={node.id}
                          entryKey={node}
                          onOpen={() => {
                            setSelected(node);
                            setEntryOpen(true);
                          }}
                        />
                      ))}
                    </TableBody>
                  </Table>
                </div>
                <div className={pager()}>
                  <Pagination
                    hasPrevious={pageInfo.hasPreviousPage}
                    hasNext={pageInfo.hasNextPage}
                    previousLabel={t("auditLogPage.actions.previous")}
                    nextLabel={t("auditLogPage.actions.next")}
                    showLabels
                    disabled={isPending}
                    onPrevious={goPrevious}
                    onNext={goNext}
                  />
                </div>
              </>
            )}
      </div>
      {selected != null && (
        <AuditLogEntryDialog
          entryKey={selected}
          open={entryOpen}
          onOpenChange={setEntryOpen}
        />
      )}
    </>
  );
}
