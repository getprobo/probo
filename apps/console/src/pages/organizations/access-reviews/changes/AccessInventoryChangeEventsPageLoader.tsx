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

import type { AccessInventoryChangeEventsPageQuery } from "#/__generated__/core/AccessInventoryChangeEventsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { ACCESS_INVENTORY_CHANGE_EVENTS_PAGE_SIZE } from "./_lib/pageSize";
import { useAccessInventoryChangeEventsFilters } from "./_lib/useAccessInventoryChangeEventsFilters";
import {
  AccessInventoryChangeEventsPage,
  accessInventoryChangeEventsPageQuery,
} from "./AccessInventoryChangeEventsPage";
import { AccessInventoryChangeEventsPageError } from "./AccessInventoryChangeEventsPageError";
import { AccessInventoryChangeEventsPageSkeleton } from "./AccessInventoryChangeEventsPageSkeleton";

function AccessInventoryChangeEventsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const { graphqlFilter } = useAccessInventoryChangeEventsFilters();
  const filterRef = useRef(graphqlFilter);
  const [queryRef, loadQuery]
    = useQueryLoader<AccessInventoryChangeEventsPageQuery>(
      accessInventoryChangeEventsPageQuery,
    );

  useEffect(() => {
    filterRef.current = graphqlFilter;
  }, [graphqlFilter]);

  useEffect(() => {
    loadQuery({
      organizationId,
      first: ACCESS_INVENTORY_CHANGE_EVENTS_PAGE_SIZE,
      filter: filterRef.current,
    }, { fetchPolicy: "network-only" });
  }, [loadQuery, organizationId]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <AccessInventoryChangeEventsPageSkeleton />;
  }

  return (
    <Suspense fallback={<AccessInventoryChangeEventsPageSkeleton />}>
      <AccessInventoryChangeEventsPage
        key={organizationId}
        queryRef={currentQueryRef}
      />
    </Suspense>
  );
}

export default function AccessInventoryChangeEventsPageLoader() {
  return (
    <ErrorBoundary
      fallback={(_error, reset) => (
        <AccessInventoryChangeEventsPageError onRetry={reset} />
      )}
    >
      <AccessInventoryChangeEventsPageQueryLoader />
    </ErrorBoundary>
  );
}
