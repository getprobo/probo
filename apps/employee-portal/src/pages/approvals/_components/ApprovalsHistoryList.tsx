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

import { useTranslation } from "react-i18next";
import { graphql, usePaginationFragment } from "react-relay";
import { useParams } from "react-router";

import type { ApprovalsHistoryList_viewer$key } from "#/__generated__/core/ApprovalsHistoryList_viewer.graphql";
import type { ApprovalsHistoryListPaginationQuery } from "#/__generated__/core/ApprovalsHistoryListPaginationQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { DocumentListSection } from "#/pages/_components/DocumentListSection";
import { EmployeeDocumentListItem } from "#/pages/_components/EmployeeDocumentListItem";
import { DOCUMENT_LIST_PAGE_SIZE } from "#/pages/_lib/documentList";

const approvalsHistoryListFragment = graphql`
  fragment ApprovalsHistoryList_viewer on Viewer
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  )
  @refetchable(queryName: "ApprovalsHistoryListPaginationQuery")
  @throwOnFieldError {
    historyDocuments: approvableDocuments(
      organizationId: $organizationId
      first: $first
      after: $after
      filter: { approvalStates: [APPROVED] }
      orderBy: { field: UPDATED_AT, direction: DESC }
    ) @connection(key: "ApprovalsHistoryList_historyDocuments", filters: ["filter", "orderBy", "organizationId"]) {
      totalCount
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
  const { organizationId } = useParams();
  const { data, hasNext, loadNext, isLoadingNext } = usePaginationFragment<
    ApprovalsHistoryListPaginationQuery,
    ApprovalsHistoryList_viewer$key
  >(approvalsHistoryListFragment, viewerKey);

  if (organizationId == null) {
    throw new NotFoundError("organizationId is required");
  }

  const { historyDocuments } = data;
  const count = historyDocuments.totalCount;

  if (count === 0) {
    return null;
  }

  return (
    <DocumentListSection
      heading={t("history.heading", { count })}
      count={count}
      hasNext={hasNext}
      isLoadingNext={isLoadingNext}
      loadMoreLabel={t("loadMore")}
      onLoadMore={() => {
        loadNext(DOCUMENT_LIST_PAGE_SIZE);
      }}
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
