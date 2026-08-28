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

import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { SignatureDocumentPageQuery } from "#/__generated__/core/SignatureDocumentPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { DOCUMENT_QUEUE_ID_PAGE_SIZE } from "#/pages/_lib/documentQueue";
import { DOCUMENT_VERSION_PAGE_SIZE } from "#/pages/_lib/documentVersion";
import { useQueuedDocumentQuery } from "#/pages/_lib/useQueuedDocumentQuery";

import { SignatureDocumentPage, signatureDocumentPageQuery } from "./SignatureDocumentPage";
import { SignatureDocumentPageSkeleton } from "./SignatureDocumentPageSkeleton";

export default function SignatureDocumentPageLoader() {
  const { organizationId, documentId } = useParams();
  const [queryRef, loadQuery] = useQueryLoader<SignatureDocumentPageQuery>(
    signatureDocumentPageQuery,
  );

  useEffect(() => {
    if (organizationId == null || documentId == null) {
      return;
    }
    loadQuery(
      {
        organizationId,
        documentId,
        first: DOCUMENT_QUEUE_ID_PAGE_SIZE,
        versionsFirst: DOCUMENT_VERSION_PAGE_SIZE,
      },
      { fetchPolicy: "network-only" },
    );
  }, [organizationId, documentId, loadQuery]);

  const currentQueryRef = queryRef != null
    && organizationId != null
    && documentId != null
    && queryRef.variables.organizationId === organizationId
    && queryRef.variables.documentId === documentId
    ? queryRef
    : null;
  const visibleQueryRef = useQueuedDocumentQuery<SignatureDocumentPageQuery>(
    signatureDocumentPageQuery,
    currentQueryRef,
  );

  if (organizationId == null || documentId == null) {
    throw new NotFoundError("organizationId and documentId are required");
  }

  if (
    visibleQueryRef == null
    || visibleQueryRef.variables.organizationId !== organizationId
    || visibleQueryRef.variables.documentId !== documentId
  ) {
    return <SignatureDocumentPageSkeleton />;
  }

  return (
    <Suspense
      key={visibleQueryRef.variables.documentId}
      fallback={<SignatureDocumentPageSkeleton />}
    >
      <SignatureDocumentPage queryRef={visibleQueryRef} />
    </Suspense>
  );
}
