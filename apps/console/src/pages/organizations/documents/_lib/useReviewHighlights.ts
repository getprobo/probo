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

import { refreshCommentAnchors } from "@probo/ui/src/RichEditor/commentAnchors";
import { refreshSuggestionFocus } from "@probo/ui/src/RichEditor/SuggestionExtension";
import type { Editor } from "@tiptap/core";
import { useEffect } from "react";

import type { ThreadAnchorSource } from "./reviewAnchors";

// useReviewHighlights draws the open threads in the text and marks the active
// thread or suggestion, so the card and its text stay visibly linked.
export function useReviewHighlights(
  editor: Editor | null,
  threads: ReadonlyArray<ThreadAnchorSource>,
  activeThreadId: string | null,
  activeSuggestionId: string | null,
) {
  useEffect(() => {
    if (!editor) {
      return;
    }
    refreshCommentAnchors(editor, threads.map(thread => ({
      id: thread.id,
      anchorStart: thread.anchorStart,
      anchorEnd: thread.anchorEnd,
      active: thread.id === activeThreadId,
    })));
  }, [activeThreadId, editor, threads]);

  useEffect(() => {
    if (!editor) {
      return;
    }
    refreshSuggestionFocus(editor, activeSuggestionId);
  }, [activeSuggestionId, editor]);
}
