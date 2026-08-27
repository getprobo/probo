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

import type { SignaturesHistoryList_viewer$key } from "#/__generated__/core/SignaturesHistoryList_viewer.graphql";
import type { SignaturesHistoryListPaginationQuery } from "#/__generated__/core/SignaturesHistoryListPaginationQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { DocumentListSection } from "#/pages/_components/DocumentListSection";
import { EmployeeDocumentListItem } from "#/pages/_components/EmployeeDocumentListItem";
import { DOCUMENT_LIST_PAGE_SIZE } from "#/pages/_lib/documentList";

const signaturesHistoryListFragment = graphql`
  fragment SignaturesHistoryList_viewer on Viewer
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  )
  @refetchable(queryName: "SignaturesHistoryListPaginationQuery")
  @throwOnFieldError {
    historyDocuments: signableDocuments(
      organizationId: $organizationId
      first: $first
      after: $after
      filter: { signed: true }
      orderBy: { field: UPDATED_AT, direction: DESC }
    ) @connection(key: "SignaturesHistoryList_historyDocuments", filters: ["filter", "orderBy", "organizationId"]) {
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

export interface SignaturesHistoryListProps {
  viewerKey: SignaturesHistoryList_viewer$key;
}

export function SignaturesHistoryList({ viewerKey }: SignaturesHistoryListProps) {
  const { t } = useTranslation("signatures");
  const { organizationId } = useParams();
  const { data, hasNext, loadNext, isLoadingNext } = usePaginationFragment<
    SignaturesHistoryListPaginationQuery,
    SignaturesHistoryList_viewer$key
  >(signaturesHistoryListFragment, viewerKey);

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
          to={`/${organizationId}/signatures/${node.id}`}
          trailing="chevron"
        />
      ))}
    </DocumentListSection>
  );
}
