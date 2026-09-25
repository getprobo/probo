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

export const DOCUMENT_QUEUE_ID_PAGE_SIZE = 50;

export type DocumentQueueKind = "signatures" | "approvals";

export type DocumentQueueDirection = "forward" | "back";

export type DocumentQueuePage = {
  ids: string[];
  totalCount: number;
  endCursor: string | null;
  hasNextPage: boolean;
};

export type DocumentQueueSnapshot = {
  kind: DocumentQueueKind;
  organizationId: string;
  doneIds: string[];
} & DocumentQueuePage;

// Builds the snapshot ID list from the live pending connection, ensuring the
// document the employee opened is present even if it fell past the first page.
export function snapshotQueueIds(pendingIds: readonly string[], documentId: string): string[] {
  if (pendingIds.includes(documentId)) {
    return [...pendingIds];
  }
  return [documentId, ...pendingIds];
}

// First-page snapshot: freeze totalCount, keep the page cursor, prepend the
// opened document when it is not in the first page.
export function enterQueueSnapshot(
  kind: DocumentQueueKind,
  page: DocumentQueuePage,
  documentId: string,
  organizationId: string,
): DocumentQueueSnapshot {
  const ids = snapshotQueueIds(page.ids, documentId);
  return {
    kind,
    organizationId,
    ids,
    doneIds: [],
    totalCount: Math.max(page.totalCount, ids.length),
    endCursor: page.endCursor,
    hasNextPage: page.hasNextPage,
  };
}

// Appends a later page. Skips ids already in the snapshot. totalCount only
// rises (older doc assigned mid-queue); signing never lowers it.
export function appendQueuePage(
  snapshot: DocumentQueueSnapshot,
  page: DocumentQueuePage,
): DocumentQueueSnapshot {
  const seen = new Set(snapshot.ids);
  const ids = [...snapshot.ids];
  for (const id of page.ids) {
    if (!seen.has(id)) {
      seen.add(id);
      ids.push(id);
    }
  }
  return {
    kind: snapshot.kind,
    organizationId: snapshot.organizationId,
    ids,
    doneIds: snapshot.doneIds,
    totalCount: Math.max(snapshot.totalCount, ids.length),
    endCursor: page.endCursor,
    hasNextPage: page.hasNextPage,
  };
}

export function markQueueDone(
  snapshot: DocumentQueueSnapshot,
  documentId: string,
): DocumentQueueSnapshot {
  if (snapshot.doneIds.includes(documentId)) {
    return snapshot;
  }
  return {
    ...snapshot,
    doneIds: [...snapshot.doneIds, documentId],
  };
}

function firstUndone(
  ids: readonly string[],
  doneIds: readonly string[],
  start: number,
  end: number,
): string | null {
  const done = new Set(doneIds);
  for (let i = start; i < end; i++) {
    const id = ids[i];
    if (id != null && !done.has(id)) {
      return id;
    }
  }
  return null;
}

// Next leftover pending id after the current one. Wraps to an earlier leftover
// only once the snapshot has no further page to load.
export function nextForwardId(
  snapshot: DocumentQueueSnapshot,
  documentId: string,
): string | null {
  const index = snapshot.ids.indexOf(documentId);
  const after = firstUndone(
    snapshot.ids,
    snapshot.doneIds,
    index >= 0 ? index + 1 : 0,
    snapshot.ids.length,
  );
  if (after != null) {
    return after;
  }
  if (snapshot.hasNextPage) {
    return null;
  }
  return firstUndone(
    snapshot.ids,
    snapshot.doneIds,
    0,
    index >= 0 ? index : 0,
  );
}

export function canGoForward(
  snapshot: DocumentQueueSnapshot,
  documentId: string,
): boolean {
  return nextForwardId(snapshot, documentId) != null || snapshot.hasNextPage;
}
