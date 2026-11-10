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
import { createPortal } from "react-dom";

import type { DocumentReviewMarginSuggestion_suggestion$key } from "#/__generated__/core/DocumentReviewMarginSuggestion_suggestion.graphql";
import type { DocumentReviewMarginThread_thread$key } from "#/__generated__/core/DocumentReviewMarginThread_thread.graphql";

import type { SuggestionPreview } from "../../_lib/suggestionPreview";
import type { CommentDraft } from "../../_lib/useReviewMarginLayout";

import { DocumentReviewMarginComposer } from "./DocumentReviewMarginComposer";
import { DocumentReviewMarginFloating } from "./DocumentReviewMarginFloating";
import { DocumentReviewMarginSuggestion } from "./DocumentReviewMarginSuggestion";
import { DocumentReviewMarginThread } from "./DocumentReviewMarginThread";

type PositionedThread = {
  key: string;
  anchor: number;
  threadKey: DocumentReviewMarginThread_thread$key;
};

type SuggestionCard = {
  key: string;
  anchor: number;
  suggestionKey?: DocumentReviewMarginSuggestion_suggestion$key | null;
  localSuggestion?: {
    id: string;
    author: { fullName: string | null | undefined };
  } | null;
  threadKey?: DocumentReviewMarginThread_thread$key | null;
  preview: SuggestionPreview;
};

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

interface DocumentReviewMarginListProps {
  documentId: string;
  narrow: boolean;
  marginHost: HTMLElement | null;
  tops: Record<string, number>;
  draft: CommentDraft | null;
  positioned: ReadonlyArray<PositionedThread>;
  suggestionCards: ReadonlyArray<SuggestionCard>;
  floatingThreads: ReadonlyArray<FloatingThread>;
  floatingSuggestions: ReadonlyArray<FloatingSuggestion>;
  activeThreadId: string | null;
  activeSuggestionId: string | null;
  canUpdate: boolean;
  canComment: boolean;
  profileId: string | null;
  mentionSource: MentionSource;
  bindCard: (node: HTMLDivElement) => (() => void) | undefined;
  onActiveThread: (id: string | null) => void;
  onActiveSuggestion: (id: string | null) => void;
  onDraftCancel: () => void;
  onDraftCreated: () => void;
  onSuggestionDecided: (id: string) => void;
  onOwnReview: () => void;
  onOwnReviewFailed: () => void;
  onLayout?: () => void;
}

export function DocumentReviewMarginList({
  documentId,
  narrow,
  marginHost,
  tops,
  draft,
  positioned,
  suggestionCards,
  floatingThreads,
  floatingSuggestions,
  activeThreadId,
  activeSuggestionId,
  canUpdate,
  canComment,
  profileId,
  mentionSource,
  bindCard,
  onActiveThread,
  onActiveSuggestion,
  onDraftCancel,
  onDraftCreated,
  onSuggestionDecided,
  onOwnReview,
  onOwnReviewFailed,
  onLayout,
}: DocumentReviewMarginListProps) {
  const composer = draft
    ? (
        <DocumentReviewMarginComposer
          documentId={documentId}
          draft={draft}
          mentionSource={mentionSource}
          onOwnReview={onOwnReview}
          onOwnReviewFailed={onOwnReviewFailed}
          onCancel={onDraftCancel}
          onCreated={onDraftCreated}
        />
      )
    : null;

  const floating = (
    <DocumentReviewMarginFloating
      documentId={documentId}
      suggestions={floatingSuggestions}
      threads={floatingThreads}
      activeSuggestionId={activeSuggestionId}
      activeThreadId={activeThreadId}
      draft={draft != null}
      canUpdate={canUpdate}
      canComment={canComment}
      profileId={profileId}
      mentionSource={mentionSource}
      onActiveSuggestion={onActiveSuggestion}
      onActiveThread={onActiveThread}
      onSuggestionDecided={onSuggestionDecided}
      onOwnReview={onOwnReview}
      onOwnReviewFailed={onOwnReviewFailed}
      onLayout={onLayout}
    />
  );

  if (narrow) {
    const dockedSuggestion = suggestionCards.find(item => item.key === activeSuggestionId);
    const dockedThread = positioned.find(item => item.key === activeThreadId);
    const docked = draft
      ? composer
      : dockedSuggestion
        ? (
            <DocumentReviewMarginSuggestion
              documentId={documentId}
              suggestionKey={dockedSuggestion.suggestionKey}
              localSuggestion={dockedSuggestion.localSuggestion}
              threadKey={dockedSuggestion.threadKey}
              preview={dockedSuggestion.preview}
              active
              canUpdate={canUpdate}
              canComment={canComment}
              profileId={profileId}
              mentionSource={mentionSource}
              onActivate={() => {
                onActiveSuggestion(dockedSuggestion.key);
                onActiveThread(null);
              }}
              onDecided={onSuggestionDecided}
              onOwnReview={onOwnReview}
              onOwnReviewFailed={onOwnReviewFailed}
              onLayout={onLayout}
            />
          )
        : dockedThread
          ? (
              <DocumentReviewMarginThread
                threadKey={dockedThread.threadKey}
                documentId={documentId}
                active
                profileId={profileId}
                canComment={canComment}
                mentionSource={mentionSource}
                onActivate={() => {
                  onActiveThread(dockedThread.key);
                  onActiveSuggestion(null);
                }}
                onResolved={() => {
                  if (activeThreadId === dockedThread.key) onActiveThread(null);
                }}
                onOwnReview={onOwnReview}
                onOwnReviewFailed={onOwnReviewFailed}
                onLayout={onLayout}
              />
            )
          : null;
    return docked
      ? <div className="sticky top-2 z-20 mx-3 mb-3">{docked}</div>
      : floating;
  }

  const desktopCards = (
    <>
      {floating}
      {positioned.map(item => (
        <div
          key={item.key}
          ref={bindCard}
          data-review-card={item.key}
          className="absolute right-3 z-10 w-60"
          style={{ top: tops[item.key] ?? item.anchor }}
        >
          <DocumentReviewMarginThread
            threadKey={item.threadKey}
            documentId={documentId}
            active={activeThreadId === item.key && !draft}
            profileId={profileId}
            canComment={canComment}
            mentionSource={mentionSource}
            onActivate={() => {
              onActiveThread(item.key);
              onActiveSuggestion(null);
            }}
            onResolved={() => {
              if (activeThreadId === item.key) onActiveThread(null);
            }}
            onOwnReview={onOwnReview}
            onOwnReviewFailed={onOwnReviewFailed}
            onLayout={onLayout}
          />
        </div>
      ))}
      {suggestionCards.map(item => (
        <div
          key={item.key}
          ref={bindCard}
          data-review-card={item.key}
          className="absolute right-3 z-10 w-60"
          style={{ top: tops[item.key] ?? item.anchor }}
        >
          <DocumentReviewMarginSuggestion
            documentId={documentId}
            suggestionKey={item.suggestionKey}
            localSuggestion={item.localSuggestion}
            threadKey={item.threadKey}
            preview={item.preview}
            active={activeSuggestionId === item.key && !draft}
            canUpdate={canUpdate}
            canComment={canComment}
            profileId={profileId}
            mentionSource={mentionSource}
            onActivate={() => {
              onActiveSuggestion(item.key);
              onActiveThread(null);
            }}
            onDecided={onSuggestionDecided}
            onOwnReview={onOwnReview}
            onOwnReviewFailed={onOwnReviewFailed}
            onLayout={onLayout}
          />
        </div>
      ))}
      {draft && tops.draft != null && (
        <div
          ref={bindCard}
          data-review-card="draft"
          className="absolute right-3 z-10 w-60"
          style={{ top: tops.draft }}
        >
          {composer}
        </div>
      )}
    </>
  );

  return marginHost
    ? createPortal(desktopCards, marginHost)
    : desktopCards;
}
