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

import { useCallback } from "react";
import { useSearchParams } from "react-router";

export type DocumentTab = "all" | "public" | "private";

export const DOCUMENT_TABS: readonly DocumentTab[] = ["all", "public", "private"];

interface DocumentTabState {
  tab: DocumentTab;
  setTab: (value: DocumentTab) => void;
}

// Active documents tab (All / Public / Private), persisted in the URL so it is
// shareable and survives reloads. Pure URL state — no local state or effects —
// so the loader, page, and toolbar can all read it without racing.
export function useDocumentTab(): DocumentTabState {
  const [searchParams, setSearchParams] = useSearchParams();

  const raw = searchParams.get("tab");
  const tab: DocumentTab = raw === "public" || raw === "private" ? raw : "all";

  const setTab = useCallback((value: DocumentTab) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value === "all") {
        next.delete("tab");
      } else {
        next.set("tab", value);
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  return { tab, setTab };
}
