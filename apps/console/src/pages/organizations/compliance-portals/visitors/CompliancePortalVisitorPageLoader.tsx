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

import type { CompliancePortalVisitorPageQuery } from "#/__generated__/core/CompliancePortalVisitorPageQuery.graphql";

import {
  documentAccessListGraphqlFilter,
  useDocumentAccessListFilters,
} from "./_lib/useDocumentAccessListFilters";
import {
  CompliancePortalVisitorPage,
  compliancePortalVisitorPageQuery,
} from "./CompliancePortalVisitorPage";
import { CompliancePortalVisitorPageSkeleton } from "./CompliancePortalVisitorPageSkeleton";

export default function CompliancePortalVisitorPageLoader() {
  const { accessId } = useParams<{
    accessId: string;
  }>();
  if (accessId == null) {
    throw new Error(":accessId missing in route params");
  }
  const { status } = useDocumentAccessListFilters();
  const filterRef = useRef(documentAccessListGraphqlFilter(status));
  const [queryRef, loadQuery] = useQueryLoader<CompliancePortalVisitorPageQuery>(
    compliancePortalVisitorPageQuery,
  );

  useEffect(() => {
    filterRef.current = documentAccessListGraphqlFilter(status);
  }, [status]);

  useEffect(() => {
    loadQuery({ accessId, filter: filterRef.current });
  }, [accessId, loadQuery]);

  if (queryRef == null) {
    return <CompliancePortalVisitorPageSkeleton />;
  }

  return (
    <Suspense fallback={<CompliancePortalVisitorPageSkeleton />}>
      <CompliancePortalVisitorPage queryRef={queryRef} />
    </Suspense>
  );
}
