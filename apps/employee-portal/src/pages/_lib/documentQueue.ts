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

export const DOCUMENT_QUEUE_STORAGE_KEY = "employee-portal:document-queue";

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
} & DocumentQueuePage;

function isQueueKind(value: unknown): value is DocumentQueueKind {
  return value === "signatures" || value === "approvals";
}

function isStringArray(value: unknown): value is string[] {
  return Array.isArray(value) && value.every(id => typeof id === "string");
}

// Accepts today's snapshot and older { kind, ids } payloads from sessionStorage.
function parseSnapshot(value: unknown): DocumentQueueSnapshot | null {
  if (value == null || typeof value !== "object") {
    return null;
  }
  const candidate = value as {
    kind?: unknown;
    ids?: unknown;
    totalCount?: unknown;
    endCursor?: unknown;
    hasNextPage?: unknown;
  };
  if (!isQueueKind(candidate.kind) || !isStringArray(candidate.ids)) {
    return null;
  }
  const ids = candidate.ids;
  const totalCount = typeof candidate.totalCount === "number"
    ? candidate.totalCount
    : ids.length;
  const endCursor = typeof candidate.endCursor === "string" ? candidate.endCursor : null;
  const hasNextPage = candidate.hasNextPage === true;

  return {
    kind: candidate.kind,
    ids,
    totalCount,
    endCursor,
    hasNextPage,
  };
}

// Reads the frozen queue snapshot written when the employee entered a pending
// flow. Returns null when storage is empty or the payload is malformed.
export function readDocumentQueueSnapshot(): DocumentQueueSnapshot | null {
  try {
    const raw = sessionStorage.getItem(DOCUMENT_QUEUE_STORAGE_KEY);
    if (raw == null) {
      return null;
    }
    return parseSnapshot(JSON.parse(raw));
  } catch {
    return null;
  }
}

// Persists the enter-time pending ID list so the counter stays frozen after
// sign / approve / reject.
export function writeDocumentQueueSnapshot(snapshot: DocumentQueueSnapshot): void {
  sessionStorage.setItem(DOCUMENT_QUEUE_STORAGE_KEY, JSON.stringify(snapshot));
}

// Drops the snapshot when the employee leaves the flow.
export function clearDocumentQueueSnapshot(): void {
  sessionStorage.removeItem(DOCUMENT_QUEUE_STORAGE_KEY);
}

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
): DocumentQueueSnapshot {
  const ids = snapshotQueueIds(page.ids, documentId);
  return {
    kind,
    ids,
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
    ids,
    totalCount: Math.max(snapshot.totalCount, ids.length),
    endCursor: page.endCursor,
    hasNextPage: page.hasNextPage,
  };
}
