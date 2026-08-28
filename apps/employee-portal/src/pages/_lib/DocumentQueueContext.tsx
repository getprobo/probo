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

import { Toast } from "@base-ui/react/toast";
import { formatError, type GraphQLError } from "@probo/helpers";
import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useParams } from "react-router";

import {
  appendQueuePage,
  type DocumentQueueDirection,
  type DocumentQueueKind,
  type DocumentQueuePage,
  type DocumentQueueSnapshot,
  enterQueueSnapshot,
} from "./documentQueue";
import { fetchDocumentQueuePage } from "./fetchDocumentQueuePage";

type DocumentQueueContextValue = {
  snapshot: DocumentQueueSnapshot | null;
  advancing: boolean;
  enter: (kind: DocumentQueueKind, page: DocumentQueuePage, documentId: string) => void;
  leave: () => void;
  goTo: (documentId: string, direction: DocumentQueueDirection) => void;
  goForward: () => void;
  startQueue: (kind: DocumentQueueKind) => void;
  close: (kind?: DocumentQueueKind) => void;
};

const DocumentQueueContext = createContext<DocumentQueueContextValue | null>(null);

function setQueueDirection(direction: DocumentQueueDirection): void {
  document.documentElement.dataset.queueDirection = direction;
}

// Holds the frozen pending-document snapshot for the signing / approval flow.
export function DocumentQueueProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate();
  const { organizationId, documentId } = useParams();
  const toast = Toast.useToastManager();
  const { t } = useTranslation();
  const [snapshot, setSnapshot] = useState<DocumentQueueSnapshot | null>(null);
  const [advancing, setAdvancing] = useState(false);
  const fetchGenerationRef = useRef(0);

  useEffect(() => {
    try {
      // Drop a snapshot left by the previous sessionStorage-backed queue.
      sessionStorage.removeItem("employee-portal:document-queue");
    } catch {
      // Ignore: private mode / quota.
    }
  }, []);

  const enter = useCallback((
    kind: DocumentQueueKind,
    page: DocumentQueuePage,
    openedDocumentId: string,
  ) => {
    if (organizationId == null) {
      return;
    }
    fetchGenerationRef.current += 1;
    setAdvancing(false);
    setSnapshot(enterQueueSnapshot(
      kind,
      page,
      openedDocumentId,
      organizationId,
    ));
  }, [organizationId]);

  const leave = useCallback(() => {
    fetchGenerationRef.current += 1;
    setSnapshot(null);
    setAdvancing(false);
  }, []);

  const goTo = useCallback((targetId: string, direction: DocumentQueueDirection) => {
    if (
      organizationId == null
      || snapshot == null
      || snapshot.organizationId !== organizationId
    ) {
      return;
    }
    setQueueDirection(direction);
    void navigate(`/${organizationId}/${snapshot.kind}/${targetId}`);
  }, [navigate, organizationId, snapshot]);

  const goForward = useCallback(() => {
    if (
      organizationId == null
      || snapshot == null
      || snapshot.organizationId !== organizationId
      || documentId == null
      || advancing
    ) {
      return;
    }
    const index = snapshot.ids.indexOf(documentId);
    const nextId = index >= 0 && index < snapshot.ids.length - 1
      ? snapshot.ids[index + 1]
      : null;
    if (nextId != null) {
      goTo(nextId, "forward");
      return;
    }
    if (!snapshot.hasNextPage || snapshot.endCursor == null) {
      return;
    }

    const generation = fetchGenerationRef.current;
    setAdvancing(true);
    void fetchDocumentQueuePage({
      kind: snapshot.kind,
      organizationId,
      after: snapshot.endCursor,
    }).then((page) => {
      if (generation !== fetchGenerationRef.current) {
        return;
      }
      const next = appendQueuePage(snapshot, page);
      setSnapshot(next);
      const firstNew = next.ids.find(id => !snapshot.ids.includes(id));
      if (firstNew != null) {
        setQueueDirection("forward");
        void navigate(`/${organizationId}/${snapshot.kind}/${firstNew}`);
      }
    }).catch((error: unknown) => {
      if (generation !== fetchGenerationRef.current) {
        return;
      }
      toast.add({
        title: t("common.error"),
        description: formatError(t("common.error"), error as GraphQLError),
        type: "error",
      });
    }).finally(() => {
      if (generation !== fetchGenerationRef.current) {
        return;
      }
      setAdvancing(false);
    });
  }, [advancing, documentId, goTo, navigate, organizationId, snapshot, t, toast]);

  const startQueue = useCallback((kind: DocumentQueueKind) => {
    if (organizationId == null || advancing) {
      return;
    }

    const generation = fetchGenerationRef.current;
    setAdvancing(true);
    void fetchDocumentQueuePage({
      kind,
      organizationId,
    }).then((page) => {
      if (generation !== fetchGenerationRef.current) {
        return;
      }
      const firstId = page.ids[0];
      if (firstId == null) {
        toast.add({
          title: t("queue.empty"),
          type: "error",
        });
        return;
      }
      enter(kind, page, firstId);
      void navigate(`/${organizationId}/${kind}/${firstId}`);
    }).catch((error: unknown) => {
      if (generation !== fetchGenerationRef.current) {
        return;
      }
      toast.add({
        title: t("common.error"),
        description: formatError(t("common.error"), error as GraphQLError),
        type: "error",
      });
    }).finally(() => {
      if (generation !== fetchGenerationRef.current) {
        return;
      }
      setAdvancing(false);
    });
  }, [advancing, enter, navigate, organizationId, t, toast]);

  const close = useCallback((kind?: DocumentQueueKind) => {
    const dest = snapshot?.kind ?? kind;
    leave();
    if (organizationId == null || dest == null) {
      return;
    }
    void navigate(`/${organizationId}/${dest}`);
  }, [leave, navigate, organizationId, snapshot?.kind]);

  const scopedSnapshot = snapshot != null
    && organizationId != null
    && snapshot.organizationId === organizationId
    ? snapshot
    : null;

  const value = useMemo<DocumentQueueContextValue>(() => ({
    snapshot: scopedSnapshot,
    advancing,
    enter,
    leave,
    goTo,
    goForward,
    startQueue,
    close,
  }), [advancing, close, enter, goForward, goTo, leave, scopedSnapshot, startQueue]);

  return (
    <DocumentQueueContext.Provider value={value}>
      {children}
    </DocumentQueueContext.Provider>
  );
}

export function useOptionalDocumentQueue(): DocumentQueueContextValue | null {
  return useContext(DocumentQueueContext);
}

// Reads the queue snapshot and navigation helpers. Must be used under
// DocumentQueueProvider.
export function useDocumentQueue(): DocumentQueueContextValue {
  const value = useOptionalDocumentQueue();
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
  pendingPage: DocumentQueuePage;
};

// Enters queue mode when the opened document is pending (or already in the
// snapshot) and leaves it when opening a history document.
export function useSyncDocumentQueue({
  kind,
  documentId,
  isPending,
  pendingPage,
}: SyncDocumentQueueOptions): void {
  const { snapshot, enter, leave } = useDocumentQueue();

  useLayoutEffect(() => {
    if (snapshot != null && snapshot.kind === kind && snapshot.ids.includes(documentId)) {
      return;
    }
    if (isPending) {
      enter(kind, pendingPage, documentId);
      return;
    }
    if (snapshot != null) {
      leave();
    }
  }, [documentId, enter, isPending, kind, leave, pendingPage, snapshot]);
}

// List pages drop the snapshot so a later visit captures a fresh pending count.
export function useClearDocumentQueueOnMount(): void {
  const { leave } = useDocumentQueue();

  useEffect(() => {
    leave();
  }, [leave]);
}
