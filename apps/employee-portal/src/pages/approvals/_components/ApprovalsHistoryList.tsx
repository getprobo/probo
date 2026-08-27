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

import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";
import { useParams } from "react-router";

import type { ApprovalsHistoryList_viewer$key } from "#/__generated__/core/ApprovalsHistoryList_viewer.graphql";
import type { ApprovalsHistoryListRefetchQuery } from "#/__generated__/core/ApprovalsHistoryListRefetchQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";
import { DocumentListSection } from "#/pages/_components/DocumentListSection";
import { EmployeeDocumentListItem } from "#/pages/_components/EmployeeDocumentListItem";
import { DOCUMENT_LIST_PAGE_SIZE } from "#/pages/_lib/documentList";

const approvalsHistoryListFragment = graphql`
  fragment ApprovalsHistoryList_viewer on Viewer
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    first: { type: "Int" }
    after: { type: "CursorKey" }
    last: { type: "Int" }
    before: { type: "CursorKey" }
  )
  @refetchable(queryName: "ApprovalsHistoryListRefetchQuery")
  @throwOnFieldError {
    historyDocuments: approvableDocuments(
      organizationId: $organizationId
      first: $first
      after: $after
      last: $last
      before: $before
      filter: { approvalStates: [APPROVED] }
      orderBy: { field: UPDATED_AT, direction: DESC }
    ) {
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
          ...EmployeeDocumentListItem_document
        }
      }
    }
  }
`;

export interface ApprovalsHistoryListProps {
  viewerKey: ApprovalsHistoryList_viewer$key;
}

export function ApprovalsHistoryList({ viewerKey }: ApprovalsHistoryListProps) {
  const { t } = useTranslation("approvals");
  const { t: tApp } = useTranslation();
  const { organizationId } = useParams();
  const [data, refetch] = useRefetchableFragment<
    ApprovalsHistoryListRefetchQuery,
    ApprovalsHistoryList_viewer$key
  >(approvalsHistoryListFragment, viewerKey);

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const { historyDocuments } = data;
  const { isPending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    historyDocuments.pageInfo,
    DOCUMENT_LIST_PAGE_SIZE,
  );

  if (organizationId == null) {
    throw new NotFoundError("organizationId is required");
  }

  const count = historyDocuments.totalCount;

  if (count === 0) {
    return null;
  }

  return (
    <DocumentListSection
      heading={t("history.heading", { count })}
      count={count}
      hasPrevious={historyDocuments.pageInfo.hasPreviousPage}
      hasNext={historyDocuments.pageInfo.hasNextPage}
      busy={isPending}
      previousLabel={tApp("pagination.previous")}
      nextLabel={tApp("pagination.next")}
      onPrevious={goPrevious}
      onNext={goNext}
    >
      {historyDocuments.edges.map(({ node }) => (
        <EmployeeDocumentListItem
          key={node.id}
          documentKey={node}
          to={`/${organizationId}/approvals/${node.id}`}
          trailing="chevron"
        />
      ))}
    </DocumentListSection>
  );
}
