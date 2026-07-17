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

import { NoteIcon, PlusIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePaginationFragment, usePreloadedQuery } from "react-relay";
import { useSearchParams } from "react-router";

import { EmptyState } from "#/components/EmptyState/EmptyState";
import { ListErrorBoundary } from "#/components/errors/ListErrorBoundary";
import { PageHeader } from "#/components/PageHeader/PageHeader";
import { buildNewRequestContinueUrl, NEW_REQUEST_PARAM } from "#/lib/auth/continueUrl";
import { useSignInDialog } from "#/lib/auth/signInDialogContext";

import type { RequestsPage_query$key } from "./__generated__/RequestsPage_query.graphql";
import type { RequestsPageQuery } from "./__generated__/RequestsPageQuery.graphql";
import type { RequestsPageRefetchQuery } from "./__generated__/RequestsPageRefetchQuery.graphql";
import { NewRequestDialog } from "./_components/NewRequestDialog";
import { RightsRequestListItem } from "./_components/RightsRequestListItem";
import { rightsRequestList } from "./_components/variants";
import { requestsLayout } from "./variants";

export const requestsPageQuery = graphql`
  query RequestsPageQuery {
    viewer {
      email
      fullName
    }
    ...RequestsPage_query
  }
`;

const requestsPageFragment = graphql`
  fragment RequestsPage_query on Query
  @refetchable(queryName: "RequestsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    myRightsRequests(first: $first, after: $after)
    @connection(key: "RequestsPage_myRightsRequests") {
      __id
      edges {
        node {
          id
          ...RightsRequestListItem_rightsRequest
        }
      }
    }
  }
`;

interface RequestsPageProps {
  queryRef: PreloadedQuery<RequestsPageQuery>;
}

// Data Requests page: the viewer's own data subject requests, with a "New
// Request" flow gated behind sign-in for guests. Submission and the personal
// list are scoped to the verified viewer's email server-side.
export function RequestsPage({ queryRef }: RequestsPageProps) {
  const { t } = useTranslation("requests");
  const root = usePreloadedQuery<RequestsPageQuery>(requestsPageQuery, queryRef);
  const { data, loadNext, hasNext, isLoadingNext, refetch } = usePaginationFragment<
    RequestsPageRefetchQuery,
    RequestsPage_query$key
  >(requestsPageFragment, root);

  const { openSignIn } = useSignInDialog();
  const [searchParams, setSearchParams] = useSearchParams();
  const [dialogOpen, setDialogOpen] = useState(false);

  const viewer = root.viewer;
  const requests = data.myRightsRequests?.edges?.map(edge => edge.node) ?? [];
  const connectionId = data.myRightsRequests?.__id ?? "";

  // After a guest signs in to submit, they land back here with the new-request
  // marker; open the dialog once and drop the marker so a reload can't re-open.
  const resumed = useRef(false);
  useEffect(() => {
    if (resumed.current || viewer == null || searchParams.get(NEW_REQUEST_PARAM) == null) {
      return;
    }
    resumed.current = true;
    setDialogOpen(true);
    const next = new URLSearchParams(searchParams);
    next.delete(NEW_REQUEST_PARAM);
    setSearchParams(next, { replace: true });
  }, [viewer, searchParams, setSearchParams]);

  const onNewRequest = () => {
    if (viewer != null) {
      setDialogOpen(true);
      return;
    }
    openSignIn({ continueTo: buildNewRequestContinueUrl() });
  };

  const { page, results, loadMore } = requestsLayout();
  const { card } = rightsRequestList();

  const newRequestButton = (
    <Button
      variant="soft"
      color="neutral"
      highContrast
      iconStart={<PlusIcon />}
      onClick={onNewRequest}
    >
      {t("newRequest")}
    </Button>
  );

  return (
    <>
      <PageHeader title={t("title")} count={requests.length} actions={newRequestButton} />
      <div className={page()}>
        <div className={results()}>
          <ListErrorBoundary
            onRetry={done => refetch({}, { fetchPolicy: "network-only", onComplete: done })}
          >
            {requests.length === 0
              ? (
                  <EmptyState
                    icon={<NoteIcon />}
                    title={t("empty.title")}
                    description={t("empty.description")}
                    action={newRequestButton}
                  />
                )
              : (
                  <>
                    <div className={card()}>
                      {requests.map(request => (
                        <RightsRequestListItem key={request.id} rightsRequestKey={request} />
                      ))}
                    </div>
                    {hasNext && (
                      <div className={loadMore()}>
                        <Button
                          variant="soft"
                          color="neutral"
                          onClick={() => loadNext(50)}
                          loading={isLoadingNext}
                        >
                          {t("loadMore")}
                        </Button>
                      </div>
                    )}
                  </>
                )}
          </ListErrorBoundary>
        </div>
      </div>
      {viewer != null && (
        <NewRequestDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          connectionId={connectionId}
          viewerEmail={viewer.email}
          viewerName={viewer.fullName}
        />
      )}
    </>
  );
}
