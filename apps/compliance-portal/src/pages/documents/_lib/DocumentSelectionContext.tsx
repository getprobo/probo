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

import type { ReactNode } from "react";
import { createContext, useCallback, useContext, useMemo, useState } from "react";

// The three selectable resource kinds on the documents page. They map onto the
// requestAccesses mutation's id lists (Document → documentIds, AuditReport →
// reportIds using the report file id, CompliancePortalFile → compliancePortalFileIds).
export type DocumentKind = "Document" | "AuditReport" | "CompliancePortalFile";

// A selectable row, resolved at page level so the toolbar can compute counts
// (e.g. how many selected rows are still locked) without touching each fragment.
export interface DocumentSelectionEntry {
  id: string;
  kind: DocumentKind;
  locked: boolean;
}

interface DocumentSelectionContextValue {
  selectedIds: ReadonlySet<string>;
  isSelected: (id: string) => boolean;
  toggle: (id: string) => void;
  selectAll: (ids: string[]) => void;
  clear: () => void;
}

const DocumentSelectionContext = createContext<DocumentSelectionContextValue | null>(null);

interface DocumentSelectionProviderProps {
  // Selection is cleared whenever this value changes (e.g. the active tab), so
  // switching between slices never carries a stale selection across.
  resetKey?: string;
  children: ReactNode;
}

export function DocumentSelectionProvider({ resetKey, children }: DocumentSelectionProviderProps) {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(() => new Set());

  // Reset the selection during render when the key changes (e.g. the active
  // tab). This is React's "adjust state on prop change" pattern, avoiding an
  // effect and the extra commit it would cost.
  const [prevResetKey, setPrevResetKey] = useState(resetKey);
  if (resetKey !== prevResetKey) {
    setPrevResetKey(resetKey);
    setSelectedIds(new Set());
  }

  const toggle = useCallback((id: string) => {
    setSelectedIds((current) => {
      const next = new Set(current);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const selectAll = useCallback((ids: string[]) => {
    setSelectedIds(new Set(ids));
  }, []);

  const clear = useCallback(() => {
    setSelectedIds(new Set());
  }, []);

  const isSelected = useCallback((id: string) => selectedIds.has(id), [selectedIds]);

  const value = useMemo<DocumentSelectionContextValue>(
    () => ({ selectedIds, isSelected, toggle, selectAll, clear }),
    [selectedIds, isSelected, toggle, selectAll, clear],
  );

  return (
    <DocumentSelectionContext.Provider value={value}>
      {children}
    </DocumentSelectionContext.Provider>
  );
}

export function useDocumentSelection() {
  const context = useContext(DocumentSelectionContext);
  if (context == null) {
    throw new Error("useDocumentSelection must be used within a DocumentSelectionProvider");
  }
  return context;
}
