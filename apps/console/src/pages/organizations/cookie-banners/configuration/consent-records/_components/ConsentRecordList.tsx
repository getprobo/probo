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
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useMemo, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { ConsentRecordList_cookieBanner$key } from "#/__generated__/core/ConsentRecordList_cookieBanner.graphql";
import type {
  ConsentRecordListRefetchQuery,
  CookieConsentRecordFilter,
} from "#/__generated__/core/ConsentRecordListRefetchQuery.graphql";
import type { ConsentRecordVersionFilter_cookieBanner$key } from "#/__generated__/core/ConsentRecordVersionFilter_cookieBanner.graphql";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import { cookieBannerList } from "../../../variants";
import {
  CONSENT_RECORDS_DEFAULT_ORDER,
  CONSENT_RECORDS_PAGE_SIZE,
} from "../_lib/pageSize";

import { ConsentRecordListItem } from "./ConsentRecordListItem";
import { ConsentRecordVersionFilter } from "./ConsentRecordVersionFilter";

const cookieConsentActions = [
  "ACCEPT_ALL",
  "REJECT_ALL",
  "CUSTOMIZE",
  "GPC",
  "ACKNOWLEDGE",
] as const;

type CookieConsentAction = (typeof cookieConsentActions)[number];

const actionLabels = {
  ACCEPT_ALL: "acceptAll",
  REJECT_ALL: "rejectAll",
  CUSTOMIZE: "customize",
  GPC: "gpc",
  ACKNOWLEDGE: "acknowledge",
} as const;

export const consentRecordListFragment = graphql`
  fragment ConsentRecordList_cookieBanner on CookieBanner
  @refetchable(queryName: "ConsentRecordListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "CookieConsentRecordFilter", defaultValue: null }
    order: { type: "CookieConsentRecordOrder", defaultValue: { field: CREATED_AT, direction: DESC } }
  ) {
    capabilities {
      corsless
    }
    consentRecords(
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
          ...ConsentRecordListItem_cookieConsentRecord
        }
      }
    }
  }
`;

interface ConsentRecordListProps {
  cookieBannerKey: ConsentRecordList_cookieBanner$key;
  versionFilterKey: ConsentRecordVersionFilter_cookieBanner$key;
}

function isCookieConsentAction(value: string): value is CookieConsentAction {
  return (cookieConsentActions as readonly string[]).includes(value);
}

export function ConsentRecordList({
  cookieBannerKey,
  versionFilterKey,
}: ConsentRecordListProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const [action, setAction] = useState<CookieConsentAction | null>(null);
  const [version, setVersion] = useState<number | null>(null);
  const [order, setOrder] = useState(CONSENT_RECORDS_DEFAULT_ORDER);
  const [isRefetchPending, startRefetchTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const [cookieBanner, refetch] = useRefetchableFragment<
    ConsentRecordListRefetchQuery,
    ConsentRecordList_cookieBanner$key
  >(consentRecordListFragment, cookieBannerKey);

  const graphqlFilter = useMemo((): CookieConsentRecordFilter | null => {
    if (action == null && version == null) {
      return null;
    }
    return {
      ...(action != null ? { action } : {}),
      ...(version != null ? { version } : {}),
    };
  }, [action, version]);

  const pageVariablesRef = useRef<CursorPaginationVariables>({
    first: CONSENT_RECORDS_PAGE_SIZE,
    after: null,
    last: null,
    before: null,
  });

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    pageVariablesRef.current = variables;
    refetch(
      { ...variables, filter: graphqlFilter, order },
      { fetchPolicy: "store-or-network" },
    );
  }, [graphqlFilter, order, refetch]);

  const { isPending: isPagePending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    cookieBanner.consentRecords.pageInfo,
    CONSENT_RECORDS_PAGE_SIZE,
  );

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    pageVariablesRef.current = {
      first: CONSENT_RECORDS_PAGE_SIZE,
      after: null,
      last: null,
      before: null,
    };
    startRefetchTransition(() => {
      refetch(
        { ...pageVariablesRef.current, filter: graphqlFilter, order },
        { fetchPolicy: "network-only" },
      );
    });
  }, [graphqlFilter, order, refetch]);

  const edges = cookieBanner.consentRecords.edges;
  const pageInfo = cookieBanner.consentRecords.pageInfo;
  const corsless = cookieBanner.capabilities.corsless;
  const isPending = isRefetchPending || isPagePending;
  const hasActiveFilters = action != null || version != null;
  const { root, results, pager, empty, filters, filter } = cookieBannerList({ pending: isPending });
  const allActionsLabel = t("consentRecordsPage.filters.allActions");

  return (
    <div className={root()}>
      <div className={filters()}>
        <div className={filter()}>
          <Select
            value={action}
            onValueChange={(value: string | null) => {
              if (value == null) {
                setAction(null);
                return;
              }
              if (isCookieConsentAction(value)) {
                setAction(value);
              }
            }}
          >
            <SelectTrigger
              size={2}
              placeholder={allActionsLabel}
              aria-label={t("consentRecordsPage.columns.action")}
            >
              {(value: CookieConsentAction | null) => (
                value != null
                  ? t(`consentRecordsPage.actions.${actionLabels[value]}`)
                  : allActionsLabel
              )}
            </SelectTrigger>
            <SelectPopup align="end">
              <SelectItem value={null}>{allActionsLabel}</SelectItem>
              {cookieConsentActions.map(consentAction => (
                <SelectItem key={consentAction} value={consentAction}>
                  {t(`consentRecordsPage.actions.${actionLabels[consentAction]}`)}
                </SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </div>
        <div className={filter()}>
          <ConsentRecordVersionFilter
            cookieBannerKey={versionFilterKey}
            version={version}
            onVersionChange={setVersion}
          />
        </div>
      </div>
      {edges.length === 0
        ? (
            <Card variant="soft" size={2}>
              <div className={empty()}>
                {hasActiveFilters
                  ? (
                      <Text size={2} color="faint">
                        {t("consentRecordsPage.emptyFiltered")}
                      </Text>
                    )
                  : (
                      <>
                        <Text size={3} weight="medium" highContrast>
                          {t("consentRecordsPage.empty.title")}
                        </Text>
                        <Text size={2} color="faint">
                          {t("consentRecordsPage.empty.description")}
                        </Text>
                      </>
                    )}
              </div>
            </Card>
          )
        : (
            <>
              <div aria-busy={isPending} className={results()}>
                <Table variant="surface" layout="fixed">
                  <TableHeader>
                    <TableRow>
                      <TableColumnHeaderCell width="8rem">
                        {t("consentRecordsPage.columns.action")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="7rem">
                        {t("consentRecordsPage.columns.regulation")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="8rem">
                        {t("consentRecordsPage.columns.location")}
                      </TableColumnHeaderCell>
                      <TableColumnHeaderCell width="6rem">
                        {t("consentRecordsPage.columns.bannerVersion")}
                      </TableColumnHeaderCell>
                      {corsless
                        ? (
                            <TableColumnHeaderCell overflow="truncate">
                              {t("consentRecordsPage.columns.origin")}
                            </TableColumnHeaderCell>
                          )
                        : null}
                      <TableColumnHeaderCell
                        width="10rem"
                        sort={order.direction === "ASC" ? "ascending" : "descending"}
                        onSort={() => {
                          setOrder(current => ({
                            field: "CREATED_AT",
                            direction: current.direction === "DESC" ? "ASC" : "DESC",
                          }));
                        }}
                        aria-label={t("consentRecordsPage.sort.date")}
                      >
                        {t("consentRecordsPage.columns.date")}
                      </TableColumnHeaderCell>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {edges.map(({ node }) => (
                      <ConsentRecordListItem
                        key={node.id}
                        cookieConsentRecordKey={node}
                        corsless={corsless}
                      />
                    ))}
                  </TableBody>
                </Table>
              </div>
              <div className={pager()}>
                <Pagination
                  hasPrevious={pageInfo.hasPreviousPage}
                  hasNext={pageInfo.hasNextPage}
                  previousLabel={t("consentRecordsPage.actions.previous")}
                  nextLabel={t("consentRecordsPage.actions.next")}
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
