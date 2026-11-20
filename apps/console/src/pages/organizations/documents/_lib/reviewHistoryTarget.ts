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

import { graphql, readInlineData } from "relay-runtime";

import type { reviewHistoryTarget_thread$key } from "#/__generated__/core/reviewHistoryTarget_thread.graphql";

import type { SuggestionAnchor, ThreadAnchorSource } from "./reviewAnchors";

// Where an entry of the history points: an open suggestion is found by its
// redlines; a comment thread or a decided suggestion by stored cursors.
export type ReviewHistoryTarget
  = | { kind: "suggestion"; id: string }
    | { kind: "thread"; anchor: ThreadAnchorSource };

const reviewHistoryTargetFragment = graphql`
  fragment reviewHistoryTarget_thread on DocumentCommentThread @inline {
    id
    anchorStart
    anchorEnd
    suggestion {
      id
      status
      anchorStart
      anchorEnd
    }
  }
`;

// The active key of a history entry, shared with the cards of the open view.
export function reviewHistoryKey(target: ReviewHistoryTarget): string {
  return target.kind === "suggestion" ? target.id : target.anchor.id;
}

// Null for an open suggestion whose redlines were all undone: nothing is left
// to review, so it stays out of the list until the server rejects it.
export function reviewHistoryTarget(
  threadKey: reviewHistoryTarget_thread$key,
  liveSuggestions: ReadonlyMap<string, SuggestionAnchor>,
): ReviewHistoryTarget | null {
  const thread = readInlineData(reviewHistoryTargetFragment, threadKey);
  const { suggestion } = thread;
  if (suggestion == null) {
    return {
      kind: "thread",
      anchor: { id: thread.id, anchorStart: thread.anchorStart, anchorEnd: thread.anchorEnd },
    };
  }
  if (suggestion.status === "OPEN") {
    return liveSuggestions.has(suggestion.id) ? { kind: "suggestion", id: suggestion.id } : null;
  }
  return {
    kind: "thread",
    anchor: {
      id: thread.id,
      anchorStart: suggestion.anchorStart ?? "",
      anchorEnd: suggestion.anchorEnd ?? "",
    },
  };
}
