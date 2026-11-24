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
import { useParams, useSearchParams } from "react-router";

import type { DocumentDetailsPageQuery } from "#/__generated__/core/DocumentDetailsPageQuery.graphql";
import type { DocumentHistoryContentQuery } from "#/__generated__/core/DocumentHistoryContentQuery.graphql";
import type { DocumentReviewListQuery } from "#/__generated__/core/DocumentReviewListQuery.graphql";

import { documentHistoryContentQuery } from "./_components/DocumentHistory/DocumentHistoryContent";
import { documentReviewListQuery } from "./_components/DocumentReviewList/DocumentReviewList";
import { documentRevisionParam } from "./_lib/useDocumentRevisionParam";
import { useLegacyVersionRedirect } from "./_lib/useLegacyVersionRedirect";
import { DocumentDetailsPage, documentDetailsPageQuery } from "./DocumentDetailsPage";
import { DocumentDetailsPageSkeleton } from "./DocumentDetailsPageSkeleton";

export default function DocumentDetailsPageLoader() {
  const { documentId } = useParams<{ documentId: string }>();
  const [searchParams] = useSearchParams();
  const revisionId = searchParams.get(documentRevisionParam);
  useLegacyVersionRedirect(documentId);
  const [queryRef, loadQuery]
    = useQueryLoader<DocumentDetailsPageQuery>(documentDetailsPageQuery);
  const [reviewQueryRef, loadReviewQuery]
    = useQueryLoader<DocumentReviewListQuery>(documentReviewListQuery);
  const [revisionQueryRef, loadRevisionQuery, disposeRevisionQuery]
    = useQueryLoader<DocumentHistoryContentQuery>(documentHistoryContentQuery);

  useEffect(() => {
    if (documentId) {
      loadQuery({ documentId });
      loadReviewQuery({ documentId });
    }
  }, [loadQuery, loadReviewQuery, documentId]);

  useEffect(() => {
    if (documentId && revisionId) {
      loadRevisionQuery({ documentId, revisionId });
    } else {
      disposeRevisionQuery();
    }
  }, [disposeRevisionQuery, loadRevisionQuery, documentId, revisionId]);

  if (documentId == null) {
    throw new Error(":documentId missing in route params");
  }

  const currentQueryRef = queryRef != null
    && queryRef.variables.documentId === documentId
    ? queryRef
    : null;
  const currentReviewQueryRef = reviewQueryRef != null
    && reviewQueryRef.variables.documentId === documentId
    ? reviewQueryRef
    : null;
  const currentRevisionQueryRef = revisionId != null
    && revisionQueryRef != null
    && revisionQueryRef.variables.documentId === documentId
    && revisionQueryRef.variables.revisionId === revisionId
    ? revisionQueryRef
    : null;

  if (currentQueryRef == null) {
    return <DocumentDetailsPageSkeleton />;
  }

  return (
    <Suspense fallback={<DocumentDetailsPageSkeleton />}>
      <DocumentDetailsPage
        key={documentId}
        queryRef={currentQueryRef}
        reviewQueryRef={currentReviewQueryRef}
        revisionQueryRef={currentRevisionQueryRef}
      />
    </Suspense>
  );
}
