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

import type { SharedDocumentPageQuery } from "#/__generated__/core/SharedDocumentPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { SharedDocumentPage, sharedDocumentPageQuery } from "./SharedDocumentPage";
import { SharedDocumentPageSkeleton } from "./SharedDocumentPageSkeleton";

export default function SharedDocumentPageLoader() {
  const { documentId } = useParams();
  const [queryRef, loadQuery] = useQueryLoader<SharedDocumentPageQuery>(
    sharedDocumentPageQuery,
  );

  useEffect(() => {
    if (documentId == null) {
      return;
    }
    loadQuery({ documentId }, { fetchPolicy: "network-only" });
  }, [documentId, loadQuery]);

  if (documentId == null) {
    throw new NotFoundError("documentId is required");
  }

  const currentQueryRef = queryRef != null
    && queryRef.variables.documentId === documentId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <SharedDocumentPageSkeleton />;
  }

  return (
    <Suspense key={documentId} fallback={<SharedDocumentPageSkeleton />}>
      <SharedDocumentPage queryRef={currentQueryRef} />
    </Suspense>
  );
}
