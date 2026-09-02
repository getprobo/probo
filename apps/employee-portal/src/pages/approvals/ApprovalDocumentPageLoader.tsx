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

import { Suspense } from "react";

import type { ApprovalDocumentPageQuery } from "#/__generated__/core/ApprovalDocumentPageQuery.graphql";
import { useQueuedDocumentPageQuery } from "#/pages/_lib/useQueuedDocumentPageQuery";

import { ApprovalDocumentPage, approvalDocumentPageQuery } from "./ApprovalDocumentPage";
import { ApprovalDocumentPageSkeleton } from "./ApprovalDocumentPageSkeleton";

export default function ApprovalDocumentPageLoader() {
  const queryRef = useQueuedDocumentPageQuery<ApprovalDocumentPageQuery>(
    approvalDocumentPageQuery,
  );

  if (queryRef == null) {
    return <ApprovalDocumentPageSkeleton />;
  }

  return (
    <Suspense
      key={queryRef.variables.documentId}
      fallback={<ApprovalDocumentPageSkeleton />}
    >
      <ApprovalDocumentPage queryRef={queryRef} />
    </Suspense>
  );
}
