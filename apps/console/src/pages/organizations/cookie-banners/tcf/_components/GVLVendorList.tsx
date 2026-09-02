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
import { useCallback, useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { GVLVendorList_cookieBanner$key } from "#/__generated__/core/GVLVendorList_cookieBanner.graphql";
import type { GVLVendorList_query$key } from "#/__generated__/core/GVLVendorList_query.graphql";
import type { GVLVendorListAddMutation } from "#/__generated__/core/GVLVendorListAddMutation.graphql";
import type { GVLVendorListRefetchQuery } from "#/__generated__/core/GVLVendorListRefetchQuery.graphql";
import type { GVLVendorListRemoveMutation } from "#/__generated__/core/GVLVendorListRemoveMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import type { CursorPaginationVariables } from "../_lib/useCursorPagination";
import { useCursorPagination } from "../_lib/useCursorPagination";
import { gvlVendorGraphqlFilter, useGVLVendorFilters } from "../_lib/useGVLVendorFilters";
import { tcfSection } from "../variants";

import { GVLVendorListEmpty } from "./GVLVendorListEmpty";
import { GVLVendorListItem } from "./GVLVendorListItem";

const PAGE_SIZE = 15;

const cookieBannerFragment = graphql`
  fragment GVLVendorList_cookieBanner on CookieBanner {
    id
    canUpdate: permission(action: "core:cookie-banner:update")
    gvlVendors(first: 500) {
      edges {
        node {
          iabVendorId
        }
      }
    }
  }
`;

const catalogFragment = graphql`
  fragment GVLVendorList_query on Query
  @refetchable(queryName: "GVLVendorListRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 15 }
    after: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    filter: { type: "CommonGVLVendorFilter", defaultValue: null }
  ) {
    commonGVLVendors(
      first: $first
      after: $after
      last: $last
      before: $before
      filter: $filter
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
          iabVendorId
          ...GVLVendorListItem_commonGVLVendor
        }
      }
    }
  }
`;

const addVendorMutation = graphql`
  mutation GVLVendorListAddMutation($input: AddCookieBannerGVLVendorInput!) {
    addCookieBannerGVLVendor(input: $input) {
      cookieBanner {
        gvlVendors(first: 500) {
          totalCount
          edges {
            node {
              iabVendorId
            }
          }
        }
      }
    }
  }
`;

const removeVendorMutation = graphql`
  mutation GVLVendorListRemoveMutation($input: RemoveCookieBannerGVLVendorInput!) {
    removeCookieBannerGVLVendor(input: $input) {
      cookieBanner {
        gvlVendors(first: 500) {
          totalCount
          edges {
            node {
              iabVendorId
            }
          }
        }
      }
    }
  }
`;

interface GVLVendorListProps {
  queryKey: GVLVendorList_query$key;
  cookieBannerKey: GVLVendorList_cookieBanner$key;
}

export function GVLVendorList({ queryKey, cookieBannerKey }: GVLVendorListProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const { query, membership } = useGVLVendorFilters();
  const [isSearchPending, startSearchTransition] = useTransition();
  const skipFirstRefetch = useRef(true);
  const banner = useFragment<GVLVendorList_cookieBanner$key>(
    cookieBannerFragment,
    cookieBannerKey,
  );
  const [catalog, refetch] = useRefetchableFragment<
    GVLVendorListRefetchQuery,
    GVLVendorList_query$key
  >(catalogFragment, queryKey);

  const refetchCatalog = useCallback((variables: CursorPaginationVariables) => {
    refetch(variables, { fetchPolicy: "store-or-network" });
  }, [refetch]);

  const refetchCatalogFromStart = useCallback((
    fetchPolicy: "store-or-network" | "network-only",
  ) => {
    startSearchTransition(() => {
      refetch(
        {
          first: PAGE_SIZE,
          after: null,
          last: null,
          before: null,
          filter: gvlVendorGraphqlFilter(query, membership, banner.id),
        },
        { fetchPolicy },
      );
    });
  }, [banner.id, membership, query, refetch]);

  const { isPending: isPagePending, goPrevious, goNext } = useCursorPagination(
    refetchCatalog,
    catalog.commonGVLVendors.pageInfo,
    PAGE_SIZE,
  );

  useEffect(() => {
    if (skipFirstRefetch.current) {
      skipFirstRefetch.current = false;
      return;
    }

    refetchCatalogFromStart("store-or-network");
  }, [refetchCatalogFromStart]);

  const [addVendor] = useMutation<GVLVendorListAddMutation>(addVendorMutation, {
    errorToast: t("tcfPage.errors.add"),
  });
  const [removeVendor] = useMutation<GVLVendorListRemoveMutation>(removeVendorMutation, {
    errorToast: t("tcfPage.errors.remove"),
  });

  const selectedIDs = new Set(
    (banner.gvlVendors?.edges ?? []).map(edge => edge.node.iabVendorId),
  );
  const catalogEdges = catalog.commonGVLVendors.edges;
  const isPending = isSearchPending || isPagePending;
  const { results, pager } = tcfSection({ pending: isPending });

  return (
    <div
      aria-busy={isPending}
      className={results()}
    >
      {catalogEdges.length === 0
        ? <GVLVendorListEmpty />
        : (
            <>
              <List>
                {catalogEdges.map(edge => (
                  <GVLVendorListItem
                    key={edge.node.id}
                    commonGVLVendorKey={edge.node}
                    selected={selectedIDs.has(edge.node.iabVendorId)}
                    onPress={banner.canUpdate
                      ? () => {
                          const input = {
                            cookieBannerId: banner.id,
                            iabVendorId: edge.node.iabVendorId,
                          };
                          const commit = selectedIDs.has(edge.node.iabVendorId)
                            ? removeVendor
                            : addVendor;
                          void commit({ variables: { input } })
                            .then(() => {
                              if (membership !== "all") {
                                refetchCatalogFromStart("network-only");
                              }
                            })
                            .catch(() => undefined);
                        }
                      : undefined}
                  />
                ))}
              </List>
              <div className={pager()}>
                <Pagination
                  hasPrevious={catalog.commonGVLVendors.pageInfo.hasPreviousPage}
                  hasNext={catalog.commonGVLVendors.pageInfo.hasNextPage}
                  previousLabel={t("tcfPage.previous")}
                  nextLabel={t("tcfPage.next")}
                  showLabels
                  variant="surface"
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
