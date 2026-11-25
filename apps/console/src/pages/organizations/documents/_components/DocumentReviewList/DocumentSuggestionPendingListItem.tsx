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

import { useContext, useState } from "react";

import { CurrentUser } from "#/providers/CurrentUser";

import type { SuggestionAnchor } from "../../_lib/reviewAnchors";

import { DocumentReviewAuthorView } from "./DocumentReviewAuthor";
import { DocumentReviewItem } from "./DocumentReviewItem";
import { DocumentSuggestionChangeSummary } from "./DocumentSuggestionChangeSummary";

interface DocumentSuggestionPendingListItemProps {
  anchor: SuggestionAnchor;
  active: boolean;
  onActivate: () => void;
}

// The card of a suggestion this user just started, shown while the server
// assigns its id. The real card takes its place once the thread arrives.
export function DocumentSuggestionPendingListItem({
  anchor,
  active,
  onActivate,
}: DocumentSuggestionPendingListItemProps) {
  const profile = useContext(CurrentUser).profile;
  const [createdAt] = useState(() => new Date().toISOString());

  return (
    <DocumentReviewItem kind="suggestion" active={active} onActivate={onActivate}>
      {profile != null && (
        <DocumentReviewAuthorView
          fullName={profile.fullName}
          emailAddress={profile.emailAddress}
          avatarUrl={profile.avatarUrl}
          createdAt={createdAt}
          primary
        />
      )}
      <DocumentSuggestionChangeSummary changes={anchor.changes} changeCount={anchor.changeCount} />
    </DocumentReviewItem>
  );
}
