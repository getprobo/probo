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

import type { CookieBannerResourcesPageQuery } from "#/__generated__/core/CookieBannerResourcesPageQuery.graphql";

import { RESOURCES_PAGE_SIZE } from "./_components/TrackerResourceList";
import { useResourcesListFilters } from "./_lib/useResourcesListFilters";
import { CookieBannerResourcesPage, cookieBannerResourcesPageQuery } from "./CookieBannerResourcesPage";
import { CookieBannerResourcesPageSkeleton } from "./CookieBannerResourcesPageSkeleton";

export default function CookieBannerResourcesPageLoader() {
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  if (typeof cookieBannerId !== "string") {
    throw new Error("Missing cookieBannerId parameter");
  }

  const { graphqlFilter, graphqlOrder } = useResourcesListFilters();
  const filterRef = useRef(graphqlFilter);
  const orderRef = useRef(graphqlOrder);
  const [queryRef, loadQuery] = useQueryLoader<CookieBannerResourcesPageQuery>(
    cookieBannerResourcesPageQuery,
  );

  useEffect(() => {
    filterRef.current = graphqlFilter;
  }, [graphqlFilter]);

  useEffect(() => {
    orderRef.current = graphqlOrder;
  }, [graphqlOrder]);

  useEffect(() => {
    loadQuery({
      cookieBannerId,
      first: RESOURCES_PAGE_SIZE,
      filter: filterRef.current,
      order: orderRef.current,
    });
  }, [loadQuery, cookieBannerId]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.cookieBannerId === cookieBannerId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <CookieBannerResourcesPageSkeleton />;
  }

  return (
    <Suspense fallback={<CookieBannerResourcesPageSkeleton />}>
      <CookieBannerResourcesPage key={cookieBannerId} queryRef={currentQueryRef} />
    </Suspense>
  );
}
