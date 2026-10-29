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

import type { AuditLogPageQuery } from "#/__generated__/iam/AuditLogPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import { AUDIT_LOG_PAGE_SIZE } from "./_components/AuditLogList";
import { useAuditLogListFilters } from "./_lib/useAuditLogListFilters";
import { AuditLogPage, auditLogPageQuery } from "./AuditLogPage";
import { AuditLogPageSkeleton } from "./AuditLogPageSkeleton";

function AuditLogPageQueryLoader() {
  const organizationId = useOrganizationId();
  const { graphqlFilter } = useAuditLogListFilters();
  const filterRef = useRef(graphqlFilter);
  const [queryRef, loadQuery] = useQueryLoader<AuditLogPageQuery>(auditLogPageQuery);

  useEffect(() => {
    filterRef.current = graphqlFilter;
  }, [graphqlFilter]);

  useEffect(() => {
    loadQuery({
      organizationId,
      first: AUDIT_LOG_PAGE_SIZE,
      filter: filterRef.current,
    }, { fetchPolicy: "network-only" });
  }, [loadQuery, organizationId]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <AuditLogPageSkeleton />;
  }

  return (
    <Suspense fallback={<AuditLogPageSkeleton />}>
      <AuditLogPage key={organizationId} queryRef={currentQueryRef} />
    </Suspense>
  );
}

export default function AuditLogPageLoader() {
  return (
    <IAMRelayProvider>
      <AuditLogPageQueryLoader />
    </IAMRelayProvider>
  );
}
