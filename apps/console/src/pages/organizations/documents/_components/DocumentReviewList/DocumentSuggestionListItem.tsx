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

import { graphql, useFragment } from "react-relay";

import type { DocumentSuggestionListItem_thread$key } from "#/__generated__/core/DocumentSuggestionListItem_thread.graphql";

import type { DocumentThreadConnections } from "../../_lib/documentThreadConnections";
import type { SuggestionAnchor } from "../../_lib/reviewAnchors";

import { DocumentReviewAuthor } from "./DocumentReviewAuthor";
import { DocumentReviewItem } from "./DocumentReviewItem";
import { DocumentReviewThread } from "./DocumentReviewThread";
import { DocumentSuggestionAnchorChanges } from "./DocumentSuggestionAnchorChanges";
import { DocumentSuggestionDecisions } from "./DocumentSuggestionDecisions";

const documentSuggestionListItemFragment = graphql`
  fragment DocumentSuggestionListItem_thread on DocumentCommentThread {
    suggestion {
      createdAt
      author {
        ...DocumentReviewAuthor_profile
      }
      canAccept: permission(action: "core:document:suggest-accept")
      canReject: permission(action: "core:document:suggest-reject")
      ...DocumentSuggestionDecisions_suggestion
    }
    ...DocumentReviewThread_thread
  }
`;

interface DocumentSuggestionListItemProps {
  threadKey: DocumentSuggestionListItem_thread$key;
  anchor: SuggestionAnchor;
  threadConnections: DocumentThreadConnections;
  active: boolean;
  onActivate: () => void;
}

// A suggestion's card: its changes read from the text, then its thread.
export function DocumentSuggestionListItem({
  threadKey,
  anchor,
  threadConnections,
  active,
  onActivate,
}: DocumentSuggestionListItemProps) {
  const thread = useFragment(documentSuggestionListItemFragment, threadKey);
  const suggestion = thread.suggestion;
  if (suggestion == null) {
    return null;
  }

  return (
    <DocumentReviewItem kind="suggestion" active={active} onActivate={onActivate}>
      <DocumentReviewAuthor
        profileKey={suggestion.author}
        createdAt={suggestion.createdAt}
        primary
        trailing={suggestion.canAccept || suggestion.canReject
          ? <DocumentSuggestionDecisions suggestionKey={suggestion} threadConnections={threadConnections} />
          : undefined}
      />
      <DocumentSuggestionAnchorChanges anchor={anchor} />
      <DocumentReviewThread
        threadKey={thread}
        threadConnections={threadConnections}
        expanded={active}
      />
    </DocumentReviewItem>
  );
}
