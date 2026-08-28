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

export const DOCUMENT_QUEUE_ID_PAGE_SIZE = 200;

export const DOCUMENT_QUEUE_STORAGE_KEY = "employee-portal:document-queue";

export type DocumentQueueKind = "signatures" | "approvals";

export type DocumentQueueDirection = "forward" | "back";

export type DocumentQueueSnapshot = {
  kind: DocumentQueueKind;
  ids: string[];
};

function isQueueKind(value: unknown): value is DocumentQueueKind {
  return value === "signatures" || value === "approvals";
}

function isSnapshot(value: unknown): value is DocumentQueueSnapshot {
  if (value == null || typeof value !== "object") {
    return false;
  }
  const candidate = value as { kind?: unknown; ids?: unknown };
  return isQueueKind(candidate.kind)
    && Array.isArray(candidate.ids)
    && candidate.ids.every(id => typeof id === "string");
}

// Reads the frozen queue snapshot written when the employee entered a pending
// flow. Returns null when storage is empty or the payload is malformed.
export function readDocumentQueueSnapshot(): DocumentQueueSnapshot | null {
  try {
    const raw = sessionStorage.getItem(DOCUMENT_QUEUE_STORAGE_KEY);
    if (raw == null) {
      return null;
    }
    const parsed: unknown = JSON.parse(raw);
    return isSnapshot(parsed) ? parsed : null;
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
