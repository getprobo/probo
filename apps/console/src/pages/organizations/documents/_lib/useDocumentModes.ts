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

import { useState } from "react";
import { graphql, useFragment } from "react-relay";

import type { useDocumentModes_document$key } from "#/__generated__/core/useDocumentModes_document.graphql";

export type DocumentMode = "viewing" | "editing" | "suggesting";

const documentModesFragment = graphql`
  fragment useDocumentModes_document on Document {
    status
    writeMode
    canUpdate: permission(action: "core:document:update")
    canSuggest: permission(action: "core:document:suggest")
    pendingApprovalVersion {
      __typename
    }
  }
`;

export function useDocumentModes(documentKey: useDocumentModes_document$key) {
  const document = useFragment(documentModesFragment, documentKey);
  const locked = document.status === "ARCHIVED"
    || document.writeMode === "GENERATED"
    || document.pendingApprovalVersion != null;
  const modes: DocumentMode[] = ["viewing"];
  if (document.canUpdate && !locked) {
    modes.push("editing");
  }
  if (document.canSuggest && !locked) {
    modes.push("suggesting");
  }
  const [mode, setMode] = useState<DocumentMode>("viewing");

  // Permissions can change while a mode is selected, for example when an
  // approval starts. Drop back to viewing before the forbidden mode renders.
  if (!modes.includes(mode)) {
    setMode("viewing");
  }

  function selectMode(next: DocumentMode) {
    if (modes.includes(next)) {
      setMode(next);
    }
  }

  return { mode, selectMode, modes };
}
