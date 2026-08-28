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

import { useEffect } from "react";
import { type PreloadedQuery, useQueryLoader } from "react-relay";
import { useParams } from "react-router";
import type { GraphQLTaggedNode, OperationType } from "relay-runtime";

import { NotFoundError } from "#/lib/relay/errors";

import { DOCUMENT_QUEUE_ID_PAGE_SIZE } from "./documentQueue";
import { DOCUMENT_VERSION_PAGE_SIZE } from "./documentVersion";
import { useQueuedDocumentQuery } from "./useQueuedDocumentQuery";

type QueuedDocumentPageVariables = {
  organizationId: string;
  documentId: string;
  first: number;
  versionsFirst: number;
};

function matchesRoute<TQuery extends OperationType>(
  queryRef: PreloadedQuery<TQuery> | null | undefined,
  organizationId: string | undefined,
  documentId: string | undefined,
): queryRef is PreloadedQuery<TQuery> {
  return queryRef != null
    && organizationId != null
    && documentId != null
    && queryRef.variables.organizationId === organizationId
    && queryRef.variables.documentId === documentId;
}

// Loads a queue document page query for the current route and returns the ref
// only once it can paint. Returns null while the route has no paint-ready ref,
// so the caller renders its skeleton. Shared by the signature and approval
// loaders so route and query-ref handling cannot diverge.
export function useQueuedDocumentPageQuery<TQuery extends OperationType>(
  query: GraphQLTaggedNode,
): PreloadedQuery<TQuery> | null {
  const { organizationId, documentId } = useParams();
  const [queryRef, loadQuery] = useQueryLoader<TQuery>(query);

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
      } satisfies QueuedDocumentPageVariables,
      { fetchPolicy: "network-only" },
    );
  }, [organizationId, documentId, loadQuery]);

  // Drop a ref left over from the previous document so the queue never paints
  // the outgoing page under the incoming route.
  const currentQueryRef = matchesRoute(queryRef, organizationId, documentId)
    ? queryRef
    : null;
  const visibleQueryRef = useQueuedDocumentQuery<TQuery>(query, currentQueryRef);

  if (organizationId == null || documentId == null) {
    throw new NotFoundError("organizationId and documentId are required");
  }

  return matchesRoute(visibleQueryRef, organizationId, documentId)
    ? visibleQueryRef
    : null;
}
