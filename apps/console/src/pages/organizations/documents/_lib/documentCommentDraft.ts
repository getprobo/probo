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

import type { RepoCollaborationHandle } from "@probo/collaboration";
import { encodeDocumentAnchor } from "@probo/ui/src/RichEditor/commentAnchors";
import { presenceFromPmSelection, richEditorPresenceAdapter } from "@probo/ui/src/RichEditor/repoPresence";
import type { Editor } from "@tiptap/core";

// The review column key of the comment being written, before it has a thread.
export const documentCommentDraftKey = "draft";

export type DocumentCommentDraft = {
  quote: string;
  anchorStart: string;
  anchorEnd: string;
  pos: number;
};

// Anchors carry Automerge cursors, so the comment follows its text through
// later edits by anyone.
export function commentDraftFromSelection(
  editor: Editor,
  handle: RepoCollaborationHandle,
): DocumentCommentDraft | null {
  const { from, to, empty } = editor.state.selection;
  if (empty) return null;
  const document = handle.doc();
  if (!document) return null;
  const selection = presenceFromPmSelection(richEditorPresenceAdapter(), document, from, to);
  if (!selection) return null;
  return {
    quote: editor.state.doc.textBetween(from, to, "\n"),
    anchorStart: encodeDocumentAnchor(String(selection.anchor)),
    anchorEnd: encodeDocumentAnchor(String(selection.head)),
    pos: Math.min(from, to),
  };
}
