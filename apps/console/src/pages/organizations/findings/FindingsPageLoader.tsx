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

import type { FindingsPageQuery } from "#/__generated__/core/FindingsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { useFindingsListFilters } from "./_lib/useFindingsListFilters";
import { FindingsPage, findingsPageQuery } from "./FindingsPage";
import { FindingsPageSkeleton } from "./FindingsPageSkeleton";

export default function FindingsPageLoader() {
  const organizationId = useOrganizationId();
  const filters = useFindingsListFilters();
  const filtersRef = useRef(filters);
  const [queryRef, loadQuery] = useQueryLoader<FindingsPageQuery>(findingsPageQuery);

  useEffect(() => {
    filtersRef.current = filters;
  }, [filters]);

  useEffect(() => {
    loadQuery({
      organizationId,
      ...filtersRef.current.variables,
    }, { fetchPolicy: "store-and-network" });
  }, [loadQuery, organizationId]);

  const currentQueryRef = queryRef != null && queryRef.variables.organizationId === organizationId
    ? queryRef
    : null;

  if (!currentQueryRef) {
    return <FindingsPageSkeleton />;
  }

  return (
    <Suspense fallback={<FindingsPageSkeleton />}>
      <FindingsPage key={organizationId} queryRef={currentQueryRef} />
    </Suspense>
  );
}
