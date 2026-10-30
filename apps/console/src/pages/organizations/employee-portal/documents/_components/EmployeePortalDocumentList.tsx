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

import { List } from "@probo/ui/src/v2/List/List";
import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EmployeePortalDocumentList_employeePortal$key } from "#/__generated__/core/EmployeePortalDocumentList_employeePortal.graphql";
import type { EmployeePortalDocumentList_organization$key } from "#/__generated__/core/EmployeePortalDocumentList_organization.graphql";
import type { EmployeePortalDocumentListPaginationQuery } from "#/__generated__/core/EmployeePortalDocumentListPaginationQuery.graphql";
import { SearchField } from "#/components/SearchField/SearchField";
import {
  type CursorPaginationVariables,
  useCursorPagination,
} from "#/lib/relay/useCursorPagination";

import { useEmployeePortalDocumentFilters } from "../_lib/useEmployeePortalDocumentFilters";
import { useEmployeePortalDocumentSearch } from "../_lib/useEmployeePortalDocumentSearch";
import { employeePortalDocumentsPage } from "../variants";

import { EmployeePortalDocumentListItem } from "./EmployeePortalDocumentListItem";

const NS = "organizations/employee-portal";
const PAGE_SIZE = 50;

const employeePortalFragment = graphql`
  fragment EmployeePortalDocumentList_employeePortal on EmployeePortal {
    id
    ...EmployeePortalDocumentListItem_employeePortal
  }
`;

const organizationFragment = graphql`
  fragment EmployeePortalDocumentList_organization on Organization
  @refetchable(queryName: "EmployeePortalDocumentListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    query: { type: "String", defaultValue: null }
  ) {
    documents(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: TITLE, direction: ASC }
      filter: { status: [ACTIVE], published: true, query: $query }
    ) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...EmployeePortalDocumentListItem_document
        }
      }
    }
  }
`;

export function EmployeePortalDocumentList(props: {
  organizationKey: EmployeePortalDocumentList_organization$key;
  employeePortalKey: EmployeePortalDocumentList_employeePortal$key;
}) {
  const { t } = useTranslation(NS);
  const { query } = useEmployeePortalDocumentFilters();
  const [queryInput, setQueryInput] = useEmployeePortalDocumentSearch();
  const queryFilter = query || null;
  const employeePortal = useFragment(employeePortalFragment, props.employeePortalKey);
  const [data, refetch] = useRefetchableFragment<
    EmployeePortalDocumentListPaginationQuery,
    EmployeePortalDocumentList_organization$key
  >(organizationFragment, props.organizationKey);
  const [isSearching, startSearchTransition] = useTransition();
  const previousQuery = useRef<string | null | undefined>(undefined);

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch({ ...variables, query: queryFilter }, { fetchPolicy: "store-or-network" });
  }, [refetch, queryFilter]);

  useEffect(() => {
    if (previousQuery.current === queryFilter) {
      return;
    }

    const isInitial = previousQuery.current === undefined;
    previousQuery.current = queryFilter;
    if (isInitial && queryFilter == null) {
      return;
    }

    startSearchTransition(() => {
      refetch({
        first: PAGE_SIZE,
        after: null,
        last: null,
        before: null,
        query: queryFilter,
      }, { fetchPolicy: "network-only" });
    });
  }, [queryFilter, refetch]);

  const pageInfo = data.documents.pageInfo;
  const { isPending: isPaging, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    pageInfo,
    PAGE_SIZE,
  );
  const isPending = isPaging || isSearching;
  const { list, search, results, pager } = employeePortalDocumentsPage({ pending: isPending });
  const documents = data.documents.edges.map(({ node }) => node);
  const isEmpty = documents.length === 0
    && !pageInfo.hasPreviousPage
    && !pageInfo.hasNextPage;

  if (isEmpty && queryFilter == null) {
    return (
      <Text size={2} color="faint">
        {t("documentsPage.empty")}
      </Text>
    );
  }

  return (
    <div className={list()}>
      <div className={search()}>
        <SearchField
          size={2}
          value={queryInput}
          onValueChange={setQueryInput}
          placeholder={t("documentsPage.search.placeholder")}
          aria-label={t("documentsPage.search.label")}
        />
      </div>
      <div aria-busy={isPending || undefined} className={results()}>
        {isEmpty
          ? (
              <Text size={2} color="faint">
                {t("documentsPage.search.empty")}
              </Text>
            )
          : (
              <List>
                {documents.map(document => (
                  <EmployeePortalDocumentListItem
                    key={document.id}
                    documentKey={document}
                    employeePortalKey={employeePortal}
                  />
                ))}
              </List>
            )}
      </div>
      {!isEmpty && (
        <div className={pager()}>
          <Pagination
            hasPrevious={pageInfo.hasPreviousPage}
            hasNext={pageInfo.hasNextPage}
            previousLabel={t("documentsPage.actions.previous")}
            nextLabel={t("documentsPage.actions.next")}
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
