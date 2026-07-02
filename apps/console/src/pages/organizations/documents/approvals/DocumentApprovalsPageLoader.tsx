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

import { Suspense, useCallback, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useOutletContext, useParams } from "react-router";

import type { DocumentApprovalsPageQuery } from "#/__generated__/core/DocumentApprovalsPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";

import { DocumentApprovalsPage, documentApprovalsPageQuery } from "./DocumentApprovalsPage";

function DocumentApprovalsPageQueryLoader() {
  const { documentId, versionId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }

  const { approvalRequestedAt }
    = useOutletContext<{
      approvalRequestedAt?: number;
    }>();

  const [queryRef, loadQuery] = useQueryLoader<DocumentApprovalsPageQuery>(documentApprovalsPageQuery);

  const loadQueryParams = useCallback(() => {
    loadQuery(
      {
        documentId: documentId,
        versionId: versionId ?? "",
        versionSpecified: !!versionId,
      },
      { fetchPolicy: "network-only" },
    );
  }, [documentId, versionId, loadQuery]);

  useEffect(() => {
    if (!queryRef) {
      loadQueryParams();
    }
  }, [queryRef, loadQueryParams]);

  // Reload approvals data whenever a new approval round is requested from the layout
  useEffect(() => {
    if (approvalRequestedAt) {
      loadQueryParams();
    }
  }, [approvalRequestedAt, loadQueryParams]);

  if (!queryRef) {
    return <LinkCardSkeleton />;
  }

  return <DocumentApprovalsPage queryRef={queryRef} />;
}

export default function DocumentApprovalsPageLoader() {
  return (
    <Suspense fallback={<LinkCardSkeleton />}>
      <DocumentApprovalsPageQueryLoader />
    </Suspense>
  );
}
