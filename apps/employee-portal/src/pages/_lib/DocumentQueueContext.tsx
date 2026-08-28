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

import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useMemo,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router";

import {
  clearDocumentQueueSnapshot,
  type DocumentQueueDirection,
  type DocumentQueueKind,
  type DocumentQueueSnapshot,
  readDocumentQueueSnapshot,
  snapshotQueueIds,
  writeDocumentQueueSnapshot,
} from "./documentQueue";

type DocumentQueueContextValue = {
  snapshot: DocumentQueueSnapshot | null;
  enter: (kind: DocumentQueueKind, ids: readonly string[]) => void;
  leave: () => void;
  goTo: (documentId: string, direction: DocumentQueueDirection) => void;
  close: () => void;
};

const DocumentQueueContext = createContext<DocumentQueueContextValue | null>(null);

function setQueueDirection(direction: DocumentQueueDirection): void {
  document.documentElement.dataset.queueDirection = direction;
}

// Holds the frozen pending-document snapshot for the signing / approval flow
// so MainLayout can swap chrome without a first-paint flash on refresh.
export function DocumentQueueProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate();
  const { organizationId } = useParams();
  const [snapshot, setSnapshot] = useState<DocumentQueueSnapshot | null>(
    readDocumentQueueSnapshot,
  );

  const enter = useCallback((kind: DocumentQueueKind, ids: readonly string[]) => {
    const next: DocumentQueueSnapshot = { kind, ids: [...ids] };
    writeDocumentQueueSnapshot(next);
    setSnapshot(next);
  }, []);

  const leave = useCallback(() => {
    clearDocumentQueueSnapshot();
    setSnapshot(null);
  }, []);

  const goTo = useCallback((documentId: string, direction: DocumentQueueDirection) => {
    if (organizationId == null || snapshot == null) {
      return;
    }
    setQueueDirection(direction);
    void navigate(`/${organizationId}/${snapshot.kind}/${documentId}`, {
      viewTransition: true,
    });
  }, [navigate, organizationId, snapshot]);

  const close = useCallback(() => {
    const kind = snapshot?.kind;
    leave();
    if (organizationId == null || kind == null) {
      return;
    }
    void navigate(`/${organizationId}/${kind}`);
  }, [leave, navigate, organizationId, snapshot?.kind]);

  const value = useMemo<DocumentQueueContextValue>(() => ({
    snapshot,
    enter,
    leave,
    goTo,
    close,
  }), [close, enter, goTo, leave, snapshot]);

  return (
    <DocumentQueueContext.Provider value={value}>
      {children}
    </DocumentQueueContext.Provider>
  );
}

// Reads the queue snapshot and navigation helpers. Must be used under
// DocumentQueueProvider.
export function useDocumentQueue(): DocumentQueueContextValue {
  const value = useContext(DocumentQueueContext);
  if (value == null) {
    throw new Error("useDocumentQueue must be used within DocumentQueueProvider");
  }
  return value;
}

// True when the current document belongs to the frozen snapshot (queue chrome).
export function useDocumentQueueActive(): boolean {
  const { snapshot } = useDocumentQueue();
  const { documentId } = useParams();
  return snapshot != null
    && documentId != null
    && snapshot.ids.includes(documentId);
}

type SyncDocumentQueueOptions = {
  kind: DocumentQueueKind;
  documentId: string;
  isPending: boolean;
  pendingIds: readonly string[];
};

// Enters queue mode when the opened document is pending (or already in the
// snapshot) and leaves it when opening a history document.
export function useSyncDocumentQueue({
  kind,
  documentId,
  isPending,
  pendingIds,
}: SyncDocumentQueueOptions): void {
  const { snapshot, enter, leave } = useDocumentQueue();

  useLayoutEffect(() => {
    if (snapshot != null && snapshot.kind === kind && snapshot.ids.includes(documentId)) {
      return;
    }
    if (isPending) {
      enter(kind, snapshotQueueIds(pendingIds, documentId));
      return;
    }
    if (snapshot != null) {
      leave();
    }
  }, [documentId, enter, isPending, kind, leave, pendingIds, snapshot]);
}

// List pages drop the snapshot so a later visit captures a fresh pending count.
export function useClearDocumentQueueOnMount(): void {
  const { leave } = useDocumentQueue();

  useEffect(() => {
    leave();
  }, [leave]);
}
