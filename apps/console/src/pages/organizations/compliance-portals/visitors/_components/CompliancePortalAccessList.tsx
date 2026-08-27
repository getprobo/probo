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
import { useEffect, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";
import { useParams } from "react-router";

import type { CompliancePortalAccessListFragment$key } from "#/__generated__/core/CompliancePortalAccessListFragment.graphql";
import type { CompliancePortalAccessListQuery } from "#/__generated__/core/CompliancePortalAccessListQuery.graphql";
import type { CompliancePortalAccessListRootQuery } from "#/__generated__/core/CompliancePortalAccessListRootQuery.graphql";

import { useAccessListFilters } from "../_lib/useAccessListFilters";
import { accessSection } from "../variants";

import { CompliancePortalAccessListEmpty } from "./CompliancePortalAccessListEmpty";
import { CompliancePortalAccessListItem } from "./CompliancePortalAccessListItem";

const accessListQuery = graphql`
  query CompliancePortalAccessListRootQuery(
    $compliancePortalId: ID!
    $order: CompliancePortalAccessOrder
    $filter: CompliancePortalAccessFilter
  ) {
    node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...CompliancePortalAccessListFragment @arguments(order: $order, filter: $filter)
      }
    }
  }
`;

const fragment = graphql`
  fragment CompliancePortalAccessListFragment on CompliancePortal
  @argumentDefinitions(
    first: { type: Int, defaultValue: 50 }
    after: { type: CursorKey, defaultValue: null }
    order: { type: CompliancePortalAccessOrder, defaultValue: { field: PENDING_REQUEST_COUNT, direction: DESC } }
    filter: { type: CompliancePortalAccessFilter, defaultValue: null }
  )
  @refetchable(queryName: "CompliancePortalAccessListQuery") {
    accesses(
      first: $first
      after: $after
      orderBy: $order
      filter: $filter
    ) @connection(key: "CompliancePortalAccessList_accesses" filters: ["orderBy", "filter"]) {
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
  const { t } = useTranslation("organizations/compliance-portals");
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const { order, query } = useAccessListFilters();
  const [queryVars] = useState({
    order,
    filter: { query },
  });
  const [isPending, startTransition] = useTransition();
  const { more, results } = accessSection({ pending: isPending });
  const skipFirstRefetch = useRef(true);
  const { node } = useLazyLoadQuery<CompliancePortalAccessListRootQuery>(
    accessListQuery,
    {
      compliancePortalId: compliancePortalId ?? "",
      order: queryVars.order,
      filter: queryVars.filter,
    },
  );
  const portalKey = node?.__typename === "CompliancePortal" ? node : null;
  const {
    data,
    hasNext,
    loadNext,
    isLoadingNext,
    refetch,
  } = usePaginationFragment<CompliancePortalAccessListQuery, CompliancePortalAccessListFragment$key>(
    fragment,
    portalKey,
  );

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    startTransition(() => {
      refetch({ order, filter: { query } }, { fetchPolicy: "store-or-network" });
    });
  }, [order, query, refetch]);

  if (compliancePortalId == null || data == null) {
    throw new Error("invalid type for node");
  }

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
