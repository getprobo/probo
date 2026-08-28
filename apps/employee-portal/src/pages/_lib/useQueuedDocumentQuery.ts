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

import { useLayoutEffect, useState } from "react";
import { flushSync } from "react-dom";
import type { PreloadedQuery } from "react-relay";
import {
  createOperationDescriptor,
  getRequest,
  type GraphQLTaggedNode,
  type IEnvironment,
  type OperationType,
  type Subscribable,
} from "relay-runtime";

type QueryRefSnapshot = {
  environment: IEnvironment;
  source?: Subscribable<unknown> | null;
  variables: Record<string, unknown>;
};

function clearQueueDirection(): void {
  delete document.documentElement.dataset.queueDirection;
}

let queueTransitionGeneration = 0;

function isQueryPaintReady(
  query: GraphQLTaggedNode,
  queryRef: QueryRefSnapshot,
): boolean {
  const request = getRequest(query);
  const operation = createOperationDescriptor(request, queryRef.variables);
  const { status } = queryRef.environment.check(operation);
  return status === "available" || status === "stale";
}

function whenQueryPaintReady(
  query: GraphQLTaggedNode,
  queryRef: QueryRefSnapshot,
  onReady: () => void,
): () => void {
  let settled = false;
  const ready = () => {
    if (settled) {
      return;
    }
    settled = true;
    onReady();
  };

  if (isQueryPaintReady(query, queryRef)) {
    ready();
    return () => {
      settled = true;
    };
  }

  const request = getRequest(query);
  const operation = createOperationDescriptor(request, queryRef.variables);
  const snapshot = queryRef.environment.lookup(operation.fragment);
  const storeSubscription = queryRef.environment.subscribe(snapshot, () => {
    if (isQueryPaintReady(query, queryRef)) {
      ready();
    }
  });

  const sourceSubscription = queryRef.source?.subscribe({
    next: () => {
      if (isQueryPaintReady(query, queryRef)) {
        ready();
      }
    },
    complete: ready,
    error: ready,
  });

  return () => {
    settled = true;
    storeSubscription.dispose();
    sourceSubscription?.unsubscribe();
  };
}

type DocumentWithViewTransition = Document & {
  startViewTransition?: (update: () => void) => { finished: Promise<unknown> };
};

function revealQueryRef(apply: () => void, animate: boolean): void {
  if (!animate) {
    apply();
    return;
  }

  const commit = () => {
    flushSync(apply);
  };

  const viewDocument = document as DocumentWithViewTransition;
  if (typeof viewDocument.startViewTransition !== "function") {
    commit();
    clearQueueDirection();
    return;
  }

  const generation = ++queueTransitionGeneration;
  const transition = viewDocument.startViewTransition(commit);
  void transition.finished.finally(() => {
    if (generation === queueTransitionGeneration) {
      clearQueueDirection();
    }
  });
}

// Holds the last ready document query so queue navigation does not flash a
// skeleton. When a new query can paint, swaps it inside a view transition.
export function useQueuedDocumentQuery<TQuery extends OperationType>(
  query: GraphQLTaggedNode,
  currentQueryRef: PreloadedQuery<TQuery> | null,
): PreloadedQuery<TQuery> | null {
  const [visibleQueryRef, setVisibleQueryRef] = useState<
    PreloadedQuery<TQuery> | null
  >(null);

  useLayoutEffect(() => {
    if (currentQueryRef == null || currentQueryRef === visibleQueryRef) {
      return;
    }

    return whenQueryPaintReady(query, currentQueryRef, () => {
      revealQueryRef(() => {
        setVisibleQueryRef(currentQueryRef);
      }, visibleQueryRef != null);
    });
  }, [currentQueryRef, query, visibleQueryRef]);

  return visibleQueryRef;
}
