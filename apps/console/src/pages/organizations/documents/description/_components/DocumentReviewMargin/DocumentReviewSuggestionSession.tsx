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

import type { SuggestingController } from "@probo/ui/src/RichEditor/SuggestionExtension";
import type { Editor } from "@tiptap/core";
import { type MutableRefObject, useEffect } from "react";

import {
  type StartedSuggestion,
  useDocumentReviewSuggestionSession,
} from "../../_lib/useDocumentReviewSuggestionSession";

interface DocumentReviewSuggestionSessionProps {
  documentId: string;
  editor: Editor;
  suggesting: MutableRefObject<SuggestingController>;
  loadedSuggestions: ReadonlyArray<{ id: string; author: { id: string } }>;
  profileId: string | undefined;
  profileFullName: string | null | undefined;
  activeSuggestionId: string | null;
  onActiveSuggestion: (id: string | null) => void;
  onOwnReview: () => void;
  onOwnReviewFailed: () => void;
  onStartedSuggestions: (suggestions: StartedSuggestion[]) => void;
  removeStartedSuggestionRef: MutableRefObject<((id: string) => void) | null>;
}

export function DocumentReviewSuggestionSession({
  documentId,
  editor,
  suggesting,
  loadedSuggestions,
  profileId,
  profileFullName,
  activeSuggestionId,
  onActiveSuggestion,
  onOwnReview,
  onOwnReviewFailed,
  onStartedSuggestions,
  removeStartedSuggestionRef,
}: DocumentReviewSuggestionSessionProps) {
  const { startedSuggestions, removeStartedSuggestion } = useDocumentReviewSuggestionSession({
    documentId,
    editor,
    suggesting,
    loadedSuggestions,
    profileId,
    profileFullName,
    activeSuggestionId,
    onActiveSuggestion,
    onOwnReview,
    onOwnReviewFailed,
  });

  useEffect(() => {
    onStartedSuggestions(startedSuggestions);
  }, [onStartedSuggestions, startedSuggestions]);

  useEffect(() => {
    removeStartedSuggestionRef.current = removeStartedSuggestion;
    return () => {
      removeStartedSuggestionRef.current = null;
    };
  }, [removeStartedSuggestion, removeStartedSuggestionRef]);

  return null;
}
