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

import type { InternalControlsPageQuery } from "#/__generated__/core/InternalControlsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { useInternalControlsListFilters } from "./_lib/useInternalControlsListFilters";
import { InternalControlsPage, internalControlsPageQuery } from "./InternalControlsPage";
import { InternalControlsPageSkeleton } from "./InternalControlsPageSkeleton";

export default function InternalControlsPageLoader() {
  const organizationId = useOrganizationId();
  const filters = useInternalControlsListFilters();
  const filtersRef = useRef(filters);
  const [queryRef, loadQuery]
    = useQueryLoader<InternalControlsPageQuery>(internalControlsPageQuery);

  useEffect(() => {
    filtersRef.current = filters;
  }, [filters]);

  useEffect(() => {
    const current = filtersRef.current;
    loadQuery({
      organizationId,
      ...current.variables,
    }, { fetchPolicy: "store-and-network" });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <InternalControlsPageSkeleton />;
  }

  return (
    <Suspense fallback={<InternalControlsPageSkeleton />}>
      <InternalControlsPage queryRef={queryRef} />
    </Suspense>
  );
}
