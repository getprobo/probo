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

import { formatError } from "@probo/helpers";
import {
  discardSuggestionId,
  reassignSuggestionId,
  type SuggestingController,
} from "@probo/ui/src/RichEditor/SuggestionExtension";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import type { Editor } from "@tiptap/core";
import { type MutableRefObject, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { useDocumentReviewSuggestionSessionMutation } from "#/__generated__/core/useDocumentReviewSuggestionSessionMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { insertSuggestionEdge } from "./reviewMarginStore";

export type StartedSuggestion = {
  id: string;
  author: { fullName: string | null | undefined };
};

const mutationFeedback = { errorToast: false };

const startSuggestionMutation = graphql`
  mutation useDocumentReviewSuggestionSessionMutation($input: StartDocumentSuggestionInput!) {
    startDocumentSuggestion(input: $input) {
      documentSuggestionEdge {
        cursor
        node { id status createdAt author { id fullName } }
      }
    }
  }
`;

export function useDocumentReviewSuggestionSession({
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
}: {
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
}) {
  const { t } = useTranslation();
  const toast = useToast();
  const [startSuggestion] = useMutation<useDocumentReviewSuggestionSessionMutation>(
    startSuggestionMutation,
    mutationFeedback,
  );
  const [startedSuggestions, setStartedSuggestions] = useState<StartedSuggestion[]>([]);
  const ownedSuggestionIds = useRef(new Set<string>());
  const controllerRef = useRef(suggesting.current);
  const lastBlockedToast = useRef(0);

  useEffect(() => {
    controllerRef.current = suggesting.current;
  }, [suggesting]);

  useEffect(() => {
    if (profileId) {
      for (const suggestion of loadedSuggestions) {
        if (suggestion.author.id === profileId) {
          ownedSuggestionIds.current.add(suggestion.id);
        }
      }
    }
    for (const suggestion of startedSuggestions) {
      ownedSuggestionIds.current.add(suggestion.id);
    }
    controllerRef.current.ownsSuggestion = id => ownedSuggestionIds.current.has(id);
  }, [loadedSuggestions, profileId, startedSuggestions]);

  useEffect(() => {
    controllerRef.current.onNeedId = () => {
      const controller = controllerRef.current;
      if (controller.suggestionId) return Promise.resolve(controller.suggestionId);

      const localId = `local:${crypto.randomUUID()}`;
      controller.suggestionId = localId;
      ownedSuggestionIds.current.add(localId);
      const author = { fullName: profileFullName };
      setStartedSuggestions(current => current.some(suggestion => suggestion.id === localId)
        ? current
        : [...current, { id: localId, author }]);

      const request = new Promise<string | null>((resolve) => {
        onOwnReview();
        void startSuggestion({
          variables: { input: { documentId: documentId } },
          updater: store => insertSuggestionEdge(store, documentId),
          onCompleted: (response, errors) => {
            if (controller.request === request) controller.request = null;
            if (errors?.length) {
              onOwnReviewFailed();
              if (controller.suggestionId === localId) controller.suggestionId = null;
              ownedSuggestionIds.current.delete(localId);
              discardSuggestionId(editor, localId);
              setStartedSuggestions(current => current.filter(suggestion => suggestion.id !== localId));
              toast.add({ title: t("documentDescriptionPage.errors.title"), description: formatError(t("documentDescriptionPage.review.suggestionFailed"), errors), type: "error" });
              resolve(null);
              return;
            }
            const id = response.startDocumentSuggestion.documentSuggestionEdge.node.id;
            ownedSuggestionIds.current.add(id);
            reassignSuggestionId(editor, localId, id);
            if (controller.suggestionId === localId) controller.suggestionId = id;
            setStartedSuggestions(current => current.map(suggestion =>
              suggestion.id === localId ? { id, author } : suggestion,
            ));
            if (activeSuggestionId === localId) onActiveSuggestion(id);
            resolve(id);
          },
          onError: (error) => {
            onOwnReviewFailed();
            if (controller.request === request) controller.request = null;
            if (controller.suggestionId === localId) controller.suggestionId = null;
            ownedSuggestionIds.current.delete(localId);
            discardSuggestionId(editor, localId);
            setStartedSuggestions(current => current.filter(suggestion => suggestion.id !== localId));
            toast.add({ title: t("documentDescriptionPage.errors.title"), description: error.message, type: "error" });
            resolve(null);
          },
        });
      });
      controller.request = request;
      return request;
    };
    controllerRef.current.onBurstFinished = () => {
      controllerRef.current.suggestionId = null;
    };
    controllerRef.current.onBlocked = () => {
      // A resize drag is blocked once per frame; one toast is enough.
      const now = Date.now();
      if (now - lastBlockedToast.current < 3_000) return;
      lastBlockedToast.current = now;
      toast.add({
        title: t("documentDescriptionPage.errors.title"),
        description: t("documentDescriptionPage.review.suggestionUntracked"),
        type: "error",
      });
    };
  }, [
    activeSuggestionId,
    documentId,
    editor,
    onActiveSuggestion,
    onOwnReview,
    onOwnReviewFailed,
    profileFullName,
    startSuggestion,
    t,
    toast,
  ]);

  return {
    startedSuggestions,
    removeStartedSuggestion: (id: string) => {
      setStartedSuggestions(current => current.filter(item => item.id !== id));
    },
  };
}
