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

import { useEffect, useMemo } from "react";
import { graphql, useFragment } from "react-relay";

import type { useDocumentModes_document$key } from "#/__generated__/core/useDocumentModes_document.graphql";

import { normalizeDocumentSearch, withDocumentMode } from "./documentSearch";
import { useDocumentSearch } from "./useDocumentSearch";

export type DocumentMode = "viewing" | "editing" | "suggesting";

const documentModesFragment = graphql`
  fragment useDocumentModes_document on Document {
    status
    writeMode
    canUpdate: permission(action: "core:document:update")
    canSuggest: permission(action: "core:document:suggest")
    pendingApprovalCount: versions(first: 0, filter: { statuses: [PENDING_APPROVAL] }) {
      totalCount
    }
  }
`;

export function useDocumentModes(documentKey: useDocumentModes_document$key) {
  const document = useFragment(documentModesFragment, documentKey);
  const { searchParams, mode: requested, update } = useDocumentSearch();
  const locked = document.status === "ARCHIVED"
    || document.writeMode === "GENERATED"
    || document.pendingApprovalCount.totalCount > 0;
  const canEdit = document.canUpdate && !locked;
  const canSuggest = document.canSuggest && !locked;
  const modes = useMemo(() => {
    const allowed: DocumentMode[] = ["viewing"];
    if (canEdit) {
      allowed.push("editing");
    }
    if (canSuggest) {
      allowed.push("suggesting");
    }
    return allowed;
  }, [canEdit, canSuggest]);
  const mode = modes.includes(requested) ? requested : "viewing";

  // A link can ask for a mode the viewer may not use, and permissions can
  // change while a mode is selected, for example when an approval starts.
  // The page falls back to viewing, and the URL follows.
  useEffect(() => {
    update(params => normalizeDocumentSearch(params, modes), { replace: true });
  }, [modes, searchParams, update]);

  function selectMode(next: DocumentMode) {
    if (modes.includes(next)) {
      update(params => withDocumentMode(params, next));
    }
  }

  return { mode, selectMode, modes };
}
