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
import { automergeRemoteTransaction } from "@probo/ui/src/RichEditor/AutomergeSyncPlugin";
import type { Editor } from "@tiptap/core";
import type { Transaction } from "@tiptap/pm/state";
import { useEffect, useState } from "react";

import {
  type SuggestionAnchor,
  suggestionAnchors,
  threadAnchors,
  type ThreadAnchorSource,
} from "./reviewAnchors";

export type ReviewAnchors = {
  threads: ReadonlyMap<string, number>;
  suggestions: ReadonlyMap<string, SuggestionAnchor>;
};

const emptyAnchors: ReviewAnchors = {
  threads: new Map(),
  suggestions: new Map(),
};

// useReviewAnchors follows where each thread and suggestion sits in the
// document. Local edits map thread positions through the transaction, and a
// remote change resolves the cursors again. `threads` must be memoized.
export function useReviewAnchors(
  editor: Editor | null,
  handle: RepoCollaborationHandle,
  threads: ReadonlyArray<ThreadAnchorSource>,
): ReviewAnchors {
  const [anchors, setAnchors] = useState<ReviewAnchors>(emptyAnchors);

  useEffect(() => {
    if (!editor || editor.isDestroyed) {
      return;
    }
    let positions = threadAnchors(handle, editor.state.doc, threads);
    let frame = 0;

    const publish = () => {
      if (frame !== 0) return;
      frame = window.requestAnimationFrame(() => {
        frame = 0;
        if (editor.isDestroyed) return;
        setAnchors({
          threads: new Map(positions),
          suggestions: suggestionAnchors(editor.state.doc),
        });
      });
    };

    const onTransaction = ({ transaction }: { transaction: Transaction }) => {
      if (!transaction.docChanged) return;
      if (transaction.getMeta(automergeRemoteTransaction)) {
        positions = threadAnchors(handle, editor.state.doc, threads);
      } else {
        const mapped = new Map<string, number>();
        for (const [id, pos] of positions) {
          mapped.set(id, transaction.mapping.map(pos));
        }
        positions = mapped;
      }
      publish();
    };

    editor.on("transaction", onTransaction);
    publish();
    return () => {
      editor.off("transaction", onTransaction);
      if (frame !== 0) window.cancelAnimationFrame(frame);
    };
  }, [editor, handle, threads]);

  return anchors;
}
