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

import { Pagination } from "@probo/ui/src/v2/Pagination/Pagination";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery, useRefetchableFragment } from "react-relay";

import { MailingListUpdateListItem } from "#/components/MailingListUpdateListItem/MailingListUpdateListItem";
import { PageHeader } from "#/components/PageHeader/PageHeader";
import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";
import { useCursorPagination } from "#/lib/relay/useCursorPagination";

import type { UpdatesPage_query$key } from "./__generated__/UpdatesPage_query.graphql";
import type { UpdatesPageQuery } from "./__generated__/UpdatesPageQuery.graphql";
import type { UpdatesPageRefetchQuery } from "./__generated__/UpdatesPageRefetchQuery.graphql";
import { UpdatesEmpty } from "./_components/UpdatesEmpty";
import { UpdatesList } from "./_components/UpdatesList";
import { UpdatesSubscribeButton } from "./_components/UpdatesSubscribeButton";
import { UPDATES_PAGE_SIZE } from "./_lib/constants";

export const updatesPageQuery = graphql`
  query UpdatesPageQuery($first: Int, $after: CursorKey, $last: Int, $before: CursorKey) {
    ...UpdatesPage_query @arguments(first: $first, after: $after, last: $last, before: $before)
  }
`;

const updatesPageFragment = graphql`
  fragment UpdatesPage_query on Query
  @refetchable(queryName: "UpdatesPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int" }
    after: { type: "CursorKey" }
    last: { type: "Int" }
    before: { type: "CursorKey" }
  ) {
    currentTrustCenter @required(action: THROW) {
      updates(first: $first, after: $after, last: $last, before: $before) {
        pageInfo {
          hasNextPage
          hasPreviousPage
          startCursor
          endCursor
        }
        edges {
          node {
            id
            ...MailingListUpdateListItem_update
          }
        }
      }
    }
  }
`;

interface UpdatesPageProps {
  queryRef: PreloadedQuery<UpdatesPageQuery>;
}

export function UpdatesPage({ queryRef }: UpdatesPageProps) {
  const { t } = useTranslation("updates");
  const { t: tCommon } = useTranslation();
  const root = usePreloadedQuery<UpdatesPageQuery>(updatesPageQuery, queryRef);
  const [data, refetch] = useRefetchableFragment<UpdatesPageRefetchQuery, UpdatesPage_query$key>(
    updatesPageFragment,
    root,
  );

  const refetchUpdates = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const { updates } = data.currentTrustCenter;
  const { pageInfo } = updates;
  const { isPending, goPrevious, goNext } = useCursorPagination(refetchUpdates, pageInfo, UPDATES_PAGE_SIZE);

  const nodes = updates.edges.map(edge => edge.node);
  const isEmpty = nodes.length === 0;

  return (
    <>
      <PageHeader title={t("title")} actions={isEmpty ? undefined : <UpdatesSubscribeButton />} />
      <div className="flex w-full flex-col items-center px-8 py-8">
        <div className="w-full max-w-5xl">
          {isEmpty
            ? <UpdatesEmpty />
            : (
                <div className="flex flex-col gap-8">
                  <UpdatesList busy={isPending}>
                    {nodes.map(node => (
                      <MailingListUpdateListItem key={node.id} updateKey={node} />
                    ))}
                  </UpdatesList>
                  <Pagination
                    hasPrevious={pageInfo.hasPreviousPage}
                    hasNext={pageInfo.hasNextPage}
                    previousLabel={tCommon("pagination.previous")}
                    nextLabel={tCommon("pagination.next")}
                    onPrevious={goPrevious}
                    onNext={goNext}
                  />
                </div>
              )}
        </div>
      </div>
    </>
  );
}
