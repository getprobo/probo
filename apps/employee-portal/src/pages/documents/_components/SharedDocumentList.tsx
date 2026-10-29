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
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";

import type { SharedDocumentList_employeePortal$key } from "#/__generated__/core/SharedDocumentList_employeePortal.graphql";
import type { SharedDocumentListRefetchQuery } from "#/__generated__/core/SharedDocumentListRefetchQuery.graphql";
import {
  type CursorPaginationVariables,
  useCursorPagination,
} from "#/lib/relay/useCursorPagination";
import { DOCUMENT_LIST_PAGE_SIZE } from "#/pages/_lib/documentList";

import { sharedDocumentList } from "../variants";

import { SharedDocumentListItem } from "./SharedDocumentListItem";

const NS = "documents";

const employeePortalFragment = graphql`
  fragment SharedDocumentList_employeePortal on EmployeePortal
  @argumentDefinitions(
    first: { type: "Int" }
    after: { type: "CursorKey" }
    last: { type: "Int" }
    before: { type: "CursorKey" }
  )
  @refetchable(queryName: "SharedDocumentListRefetchQuery")
  @throwOnFieldError {
    documents(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { field: CREATED_AT, direction: DESC }
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
          ...SharedDocumentListItem_document
        }
      }
    }
  }
`;

export interface SharedDocumentListProps {
  employeePortalKey: SharedDocumentList_employeePortal$key;
}

export function SharedDocumentList({ employeePortalKey }: SharedDocumentListProps) {
  const { t } = useTranslation(NS);
  const { t: tApp } = useTranslation();
  const [data, refetch] = useRefetchableFragment<
    SharedDocumentListRefetchQuery,
    SharedDocumentList_employeePortal$key
  >(employeePortalFragment, employeePortalKey);

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const pageInfo = data.documents.pageInfo;
  const { isPending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    pageInfo,
    DOCUMENT_LIST_PAGE_SIZE,
  );
  const { root, results, pager } = sharedDocumentList({ pending: isPending });
  const documents = data.documents.edges.map(({ node }) => node);
  const isEmpty = documents.length === 0
    && !pageInfo.hasPreviousPage
    && !pageInfo.hasNextPage;

  return (
    <section className={root()}>
      <Heading level={2} size={2} weight="medium" highContrast>
        {t("shared.heading")}
      </Heading>
      {isEmpty
        ? (
            <Text size={2} color="faint">
              {t("shared.empty")}
            </Text>
          )
        : (
            <>
              <div aria-busy={isPending || undefined} className={results()}>
                <List>
                  {documents.map(document => (
                    <SharedDocumentListItem
                      key={document.id}
                      documentKey={document}
                    />
                  ))}
                </List>
              </div>
              <Pagination
                className={pager()}
                hasPrevious={pageInfo.hasPreviousPage}
                hasNext={pageInfo.hasNextPage}
                previousLabel={tApp("pagination.previous")}
                nextLabel={tApp("pagination.next")}
                showLabels
                disabled={isPending}
                onPrevious={goPrevious}
                onNext={goNext}
              />
            </>
          )}
    </section>
  );
}
