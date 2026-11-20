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

import { WarningIcon } from "@phosphor-icons/react";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { DocumentPublishReviewCalloutQuery } from "#/__generated__/core/DocumentPublishReviewCalloutQuery.graphql";

export const documentPublishReviewCalloutQuery = graphql`
  query DocumentPublishReviewCalloutQuery($documentId: ID!) {
    node(id: $documentId) {
      __typename
      ... on Document {
        openThreads: commentThreads(first: 0, filter: { resolved: false }) {
          totalCount
        }
        openSuggestions: suggestions(first: 0, filter: { status: OPEN }) {
          totalCount
        }
      }
    }
  }
`;

interface DocumentPublishReviewCalloutProps {
  queryRef: PreloadedQuery<DocumentPublishReviewCalloutQuery> | null;
}

// Open comments and suggestions do not block publishing, so the dialog only
// warns about them. The warning never holds the dialog up: while it loads or
// after it fails, nothing shows.
export function DocumentPublishReviewCallout({ queryRef }: DocumentPublishReviewCalloutProps) {
  if (queryRef == null) {
    return null;
  }

  return (
    <ErrorBoundary fallback={null}>
      <Suspense fallback={null}>
        <DocumentPublishReviewCalloutLoaded queryRef={queryRef} />
      </Suspense>
    </ErrorBoundary>
  );
}

function DocumentPublishReviewCalloutLoaded({
  queryRef,
}: {
  queryRef: PreloadedQuery<DocumentPublishReviewCalloutQuery>;
}) {
  const { t } = useTranslation("organizations/documents");
  const data = usePreloadedQuery<DocumentPublishReviewCalloutQuery>(documentPublishReviewCalloutQuery, queryRef);

  if (data.node.__typename !== "Document") {
    return null;
  }

  const suggestions = data.node.openSuggestions.totalCount;
  // Each open suggestion owns exactly one open thread.
  const comments = Math.max(0, data.node.openThreads.totalCount - suggestions);

  if (comments === 0 && suggestions === 0) {
    return null;
  }

  return (
    <Callout color="amber" icon={<WarningIcon weight="fill" />}>
      {[
        comments > 0 ? t("details.actions.publish.openReview.comments", { count: comments }) : null,
        suggestions > 0 ? t("details.actions.publish.openReview.suggestions", { count: suggestions }) : null,
      ].filter(Boolean).join(" ")}
    </Callout>
  );
}
