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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { List } from "@probo/ui/src/v2/List/List";
import { ListSkeleton } from "@probo/ui/src/v2/List/ListSkeleton";
import { Suspense, useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";
import { useParams } from "react-router";

import type { CompliancePortalAccessList_compliancePortal$key } from "#/__generated__/core/CompliancePortalAccessList_compliancePortal.graphql";
import type { CompliancePortalAccessListQuery } from "#/__generated__/core/CompliancePortalAccessListQuery.graphql";
import type { CompliancePortalAccessListRefetchQuery } from "#/__generated__/core/CompliancePortalAccessListRefetchQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import {
  accessListGraphqlVariables,
  useAccessListFilters,
} from "../_lib/useAccessListFilters";
import { accessSection } from "../variants";

import { CompliancePortalAccessListEmpty } from "./CompliancePortalAccessListEmpty";
import { CompliancePortalAccessListItem } from "./CompliancePortalAccessListItem";

const accessListQuery = graphql`
  query CompliancePortalAccessListQuery(
    $compliancePortalId: ID!
    $order: CompliancePortalAccessOrder
    $filter: CompliancePortalAccessFilter
  ) {
    node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...CompliancePortalAccessList_compliancePortal @arguments(order: $order, filter: $filter)
      }
    }
  }
`;

const fragment = graphql`
  fragment CompliancePortalAccessList_compliancePortal on CompliancePortal
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    order: {
      type: "CompliancePortalAccessOrder"
      defaultValue: { field: PENDING_REQUEST_COUNT, direction: DESC }
    }
    filter: { type: "CompliancePortalAccessFilter", defaultValue: null }
  )
  @refetchable(queryName: "CompliancePortalAccessListRefetchQuery") {
    accesses(
      first: $first
      after: $after
      orderBy: $order
      filter: $filter
    ) @connection(key: "CompliancePortalAccessList_accesses", filters: ["orderBy", "filter"]) {
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      edges {
        node {
          id
          ...CompliancePortalAccessListItemFragment
        }
      }
    }
  }
`;

export function CompliancePortalAccessList() {
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const { order, query } = useAccessListFilters();
  const filterRef = useRef(accessListGraphqlVariables(order, query));
  const [queryRef, loadQuery] = useQueryLoader<CompliancePortalAccessListQuery>(
    accessListQuery,
  );

  useEffect(() => {
    filterRef.current = accessListGraphqlVariables(order, query);
  }, [order, query]);

  useEffect(() => {
    if (compliancePortalId == null) {
      return;
    }

    loadQuery({
      compliancePortalId,
      ...filterRef.current,
    });
  }, [loadQuery, compliancePortalId]);

  if (compliancePortalId == null) {
    throw new Error(":compliancePortalId missing in route params");
  }

  const currentQueryRef = queryRef != null
    && queryRef.variables.compliancePortalId === compliancePortalId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <ListSkeleton count={4} />;
  }

  return (
    <Suspense fallback={<ListSkeleton count={4} />}>
      <AccessListContent queryRef={currentQueryRef} />
    </Suspense>
  );
}

function AccessListContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<CompliancePortalAccessListQuery>;
}) {
  const queryData = usePreloadedQuery<CompliancePortalAccessListQuery>(
    accessListQuery,
    queryRef,
  );
  if (queryData.node?.__typename !== "CompliancePortal") {
    throw new NotFoundError("Compliance portal not found");
  }

  return <AccessListResults compliancePortalKey={queryData.node} />;
}

function AccessListResults({
  compliancePortalKey,
}: {
  compliancePortalKey: CompliancePortalAccessList_compliancePortal$key;
}) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { order, query } = useAccessListFilters();
  const [isPending, startTransition] = useTransition();
  const { more, results } = accessSection({ pending: isPending });
  const skipFirstRefetch = useRef(true);
  const {
    data,
    hasNext,
    loadNext,
    isLoadingNext,
    refetch,
  } = usePaginationFragment<
    CompliancePortalAccessListRefetchQuery,
    CompliancePortalAccessList_compliancePortal$key
  >(fragment, compliancePortalKey);

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    startTransition(() => {
      refetch(accessListGraphqlVariables(order, query), { fetchPolicy: "store-or-network" });
    });
  }, [order, query, refetch]);

  const { accesses } = data;

  return (
    <div
      aria-busy={isPending}
      className={results()}
    >
      {accesses.edges.length === 0
        ? (
            <CompliancePortalAccessListEmpty />
          )
        : (
            <>
              <List>
                {accesses.edges.map(({ node: access }) => (
                  <CompliancePortalAccessListItem
                    key={access.id}
                    accessKey={access}
                  />
                ))}
              </List>
              {hasNext && (
                <div className={more()}>
                  <Button
                    variant="ghost"
                    color="neutral"
                    loading={isLoadingNext}
                    onClick={() => loadNext(50)}
                  >
                    {t("accessList.actions.showMore")}
                  </Button>
                </div>
              )}
            </>
          )}
    </div>
  );
}
