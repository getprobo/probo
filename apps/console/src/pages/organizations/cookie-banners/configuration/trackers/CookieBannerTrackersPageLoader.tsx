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

import { Suspense, useEffect, useRef } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { CookieBannerTrackersPageQuery } from "#/__generated__/core/CookieBannerTrackersPageQuery.graphql";

import { useTrackersListFilters } from "./_lib/useTrackersListFilters";
import { CookieBannerTrackersPage, cookieBannerTrackersPageQuery } from "./CookieBannerTrackersPage";
import { CookieBannerTrackersPageSkeleton } from "./CookieBannerTrackersPageSkeleton";

export default function CookieBannerTrackersPageLoader() {
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  if (typeof cookieBannerId !== "string") {
    throw new Error("Missing cookieBannerId parameter");
  }

  const { graphqlFilter, graphqlOrder, graphqlPagination } = useTrackersListFilters();
  const filterRef = useRef(graphqlFilter);
  const orderRef = useRef(graphqlOrder);
  const paginationRef = useRef(graphqlPagination);
  const [queryRef, loadQuery] = useQueryLoader<CookieBannerTrackersPageQuery>(
    cookieBannerTrackersPageQuery,
  );

  useEffect(() => {
    filterRef.current = graphqlFilter;
  }, [graphqlFilter]);

  useEffect(() => {
    orderRef.current = graphqlOrder;
  }, [graphqlOrder]);

  useEffect(() => {
    paginationRef.current = graphqlPagination;
  }, [graphqlPagination]);

  useEffect(() => {
    loadQuery({
      cookieBannerId,
      ...paginationRef.current,
      filter: filterRef.current,
      order: orderRef.current,
    });
  }, [loadQuery, cookieBannerId]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.cookieBannerId === cookieBannerId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <CookieBannerTrackersPageSkeleton />;
  }

  return (
    <Suspense fallback={<CookieBannerTrackersPageSkeleton />}>
      <CookieBannerTrackersPage key={cookieBannerId} queryRef={currentQueryRef} />
    </Suspense>
  );
}
