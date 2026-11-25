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

import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Suspense, useEffect, useRef } from "react";
import { useQueryLoader } from "react-relay";

import type { AccessInventoryGrantsPageQuery } from "#/__generated__/core/AccessInventoryGrantsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { ACCESS_INVENTORY_GRANTS_PAGE_SIZE } from "./_lib/pageSize";
import { useAccessInventoryGrantsFilters } from "./_lib/useAccessInventoryGrantsFilters";
import {
  AccessInventoryGrantsPage,
  accessInventoryGrantsPageQuery,
} from "./AccessInventoryGrantsPage";
import { AccessInventoryGrantsPageError } from "./AccessInventoryGrantsPageError";
import { AccessInventoryGrantsPageSkeleton } from "./AccessInventoryGrantsPageSkeleton";

function AccessInventoryGrantsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const { graphqlFilter } = useAccessInventoryGrantsFilters();
  const filterRef = useRef(graphqlFilter);
  const [queryRef, loadQuery]
    = useQueryLoader<AccessInventoryGrantsPageQuery>(
      accessInventoryGrantsPageQuery,
    );

  useEffect(() => {
    filterRef.current = graphqlFilter;
  }, [graphqlFilter]);

  useEffect(() => {
    loadQuery({
      organizationId,
      first: ACCESS_INVENTORY_GRANTS_PAGE_SIZE,
      filter: filterRef.current,
    }, { fetchPolicy: "network-only" });
  }, [loadQuery, organizationId]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <AccessInventoryGrantsPageSkeleton />;
  }

  return (
    <Suspense fallback={<AccessInventoryGrantsPageSkeleton />}>
      <AccessInventoryGrantsPage
        key={organizationId}
        queryRef={currentQueryRef}
      />
    </Suspense>
  );
}

export default function AccessInventoryGrantsPageLoader() {
  return (
    <ErrorBoundary
      fallback={(_error, reset) => (
        <AccessInventoryGrantsPageError onRetry={reset} />
      )}
    >
      <AccessInventoryGrantsPageQueryLoader />
    </ErrorBoundary>
  );
}
