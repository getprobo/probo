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

import { StampIcon } from "@phosphor-icons/react";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useRefetchableFragment } from "react-relay";
import { useParams } from "react-router";

import type { ApprovalsPendingList_viewer$key } from "#/__generated__/core/ApprovalsPendingList_viewer.graphql";
import type { ApprovalsPendingListRefetchQuery } from "#/__generated__/core/ApprovalsPendingListRefetchQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { type CursorPaginationVariables, useCursorPagination } from "#/lib/relay/useCursorPagination";
import { DocumentEmpty } from "#/pages/_components/DocumentEmpty";
import { DocumentListSection } from "#/pages/_components/DocumentListSection";
import { DocumentQueueSummary } from "#/pages/_components/DocumentQueueSummary";
import { EmployeeDocumentListItem } from "#/pages/_components/EmployeeDocumentListItem";
import { DOCUMENT_LIST_PAGE_SIZE } from "#/pages/_lib/documentList";

const approvalsPendingListFragment = graphql`
  fragment ApprovalsPendingList_viewer on Viewer
  @argumentDefinitions(
    organizationId: { type: "ID!" }
    first: { type: "Int" }
    after: { type: "CursorKey" }
    last: { type: "Int" }
    before: { type: "CursorKey" }
  )
  @refetchable(queryName: "ApprovalsPendingListRefetchQuery")
  @throwOnFieldError {
    pendingDocuments: approvableDocuments(
      organizationId: $organizationId
      first: $first
      after: $after
      last: $last
      before: $before
      filter: { approvalStates: [PENDING] }
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
    historyCount: approvableDocuments(
      organizationId: $organizationId
      filter: { approvalStates: [APPROVED, REJECTED] }
    ) {
      totalCount
    }
  }
`;

export interface ApprovalsPendingListProps {
  viewerKey: ApprovalsPendingList_viewer$key;
}

export function ApprovalsPendingList({ viewerKey }: ApprovalsPendingListProps) {
  const { t } = useTranslation("approvals");
  const { t: tApp } = useTranslation();
  const { organizationId } = useParams();
  const [data, refetch] = useRefetchableFragment<
    ApprovalsPendingListRefetchQuery,
    ApprovalsPendingList_viewer$key
  >(approvalsPendingListFragment, viewerKey);

  const refetchPage = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const { pendingDocuments, historyCount } = data;
  const { isPending, goPrevious, goNext } = useCursorPagination(
    refetchPage,
    pendingDocuments.pageInfo,
    DOCUMENT_LIST_PAGE_SIZE,
  );

  if (organizationId == null) {
    throw new NotFoundError("organizationId is required");
  }

  const count = pendingDocuments.totalCount;
  const firstPendingId = pendingDocuments.edges[0]?.node.id ?? null;
  const emptyKey = historyCount.totalCount === 0 ? "none" : "allDone";

  return (
    <DocumentListSection
      heading={t("pending.heading", { count })}
      count={count}
      empty={(
        <DocumentEmpty
          title={t(`empty.${emptyKey}.title`)}
          description={t(`empty.${emptyKey}.description`)}
        />
      )}
      summary={firstPendingId == null
        ? undefined
        : (
            <DocumentQueueSummary
              icon={<StampIcon />}
              title={t("pending.summaryTitle", { count })}
              description={t("pending.summaryDescription")}
              actionLabel={t("pending.action")}
              actionTo={`/${organizationId}/approvals/${firstPendingId}`}
            />
          )}
      hasPrevious={pendingDocuments.pageInfo.hasPreviousPage}
      hasNext={pendingDocuments.pageInfo.hasNextPage}
      busy={isPending}
      previousLabel={tApp("pagination.previous")}
      nextLabel={tApp("pagination.next")}
      onPrevious={goPrevious}
      onNext={goNext}
    >
      {pendingDocuments.edges.map(({ node }) => (
        <EmployeeDocumentListItem
          key={node.id}
          documentKey={node}
          to={`/${organizationId}/approvals/${node.id}`}
          trailing="action"
          actionLabel={t("pending.itemAction")}
        />
      ))}
    </DocumentListSection>
  );
}
