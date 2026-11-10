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

import type { MentionSource } from "@probo/ui";
import type { ReactNode } from "react";

import type { DocumentReviewMarginSuggestion_suggestion$key } from "#/__generated__/core/DocumentReviewMarginSuggestion_suggestion.graphql";
import type { DocumentReviewMarginThread_thread$key } from "#/__generated__/core/DocumentReviewMarginThread_thread.graphql";

import type { SuggestionPreview } from "../../_lib/suggestionPreview";

import { DocumentReviewMarginSuggestion } from "./DocumentReviewMarginSuggestion";
import { DocumentReviewMarginThread } from "./DocumentReviewMarginThread";

type FloatingSuggestion = {
  id: string;
  suggestionKey?: DocumentReviewMarginSuggestion_suggestion$key | null;
  localSuggestion?: {
    id: string;
    author: { fullName: string | null | undefined };
  } | null;
  threadKey?: DocumentReviewMarginThread_thread$key | null;
  preview: SuggestionPreview;
};

type FloatingThread = {
  id: string;
  threadKey: DocumentReviewMarginThread_thread$key;
};

interface DocumentReviewMarginFloatingProps {
  documentId: string;
  suggestions: ReadonlyArray<FloatingSuggestion>;
  threads: ReadonlyArray<FloatingThread>;
  activeSuggestionId: string | null;
  activeThreadId: string | null;
  draft: boolean;
  canUpdate: boolean;
  canComment: boolean;
  profileId: string | null;
  mentionSource: MentionSource;
  onActiveSuggestion: (id: string | null) => void;
  onActiveThread: (id: string | null) => void;
  onSuggestionDecided: (id: string) => void;
  onOwnReview: () => void;
  onOwnReviewFailed: () => void;
  onLayout?: () => void;
  children?: ReactNode;
}

export function DocumentReviewMarginFloating({
  documentId,
  suggestions,
  threads,
  activeSuggestionId,
  activeThreadId,
  draft,
  canUpdate,
  canComment,
  profileId,
  mentionSource,
  onActiveSuggestion,
  onActiveThread,
  onSuggestionDecided,
  onOwnReview,
  onOwnReviewFailed,
  onLayout,
}: DocumentReviewMarginFloatingProps) {
  if (suggestions.length === 0 && threads.length === 0) return null;

  return (
    <div className="absolute top-14 right-3 z-20 flex w-60 flex-col gap-2">
      {suggestions.map(suggestion => (
        <div key={suggestion.id}>
          <DocumentReviewMarginSuggestion
            documentId={documentId}
            suggestionKey={suggestion.suggestionKey}
            localSuggestion={suggestion.localSuggestion}
            threadKey={suggestion.threadKey}
            preview={suggestion.preview}
            active={activeSuggestionId === suggestion.id && !draft}
            canUpdate={canUpdate}
            canComment={canComment}
            profileId={profileId}
            mentionSource={mentionSource}
            onActivate={() => {
              onActiveSuggestion(suggestion.id);
              onActiveThread(null);
            }}
            onDecided={onSuggestionDecided}
            onOwnReview={onOwnReview}
            onOwnReviewFailed={onOwnReviewFailed}
            onLayout={onLayout}
          />
        </div>
      ))}
      {threads.map(thread => (
        <div key={thread.id}>
          <DocumentReviewMarginThread
            threadKey={thread.threadKey}
            documentId={documentId}
            active={activeThreadId === thread.id && !draft}
            profileId={profileId}
            canComment={canComment}
            mentionSource={mentionSource}
            onActivate={() => {
              onActiveThread(thread.id);
              onActiveSuggestion(null);
            }}
            onResolved={() => {
              if (activeThreadId === thread.id) onActiveThread(null);
            }}
            onOwnReview={onOwnReview}
            onOwnReviewFailed={onOwnReviewFailed}
            onLayout={onLayout}
          />
        </div>
      ))}
    </div>
  );
}
