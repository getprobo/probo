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
import { useCallback, useContext, useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useRelayEnvironment } from "react-relay";
import { ConnectionHandler, fetchQuery, graphql } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type { useSuggestionControllerMutation } from "#/__generated__/core/useSuggestionControllerMutation.graphql";
import type { useSuggestionControllerQuery } from "#/__generated__/core/useSuggestionControllerQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { CurrentUser } from "#/providers/CurrentUser";

import type { DocumentMode } from "./useDocumentModes";

const blockedToastIntervalMs = 3_000;
const blockedToastOptions = { leading: true, trailing: false };

// The new edge carries every field the review column reads from a
// suggestion, so its card appears without a refetch.
const startSuggestionMutation = graphql`
  mutation useSuggestionControllerMutation(
    $input: StartDocumentSuggestionInput!
    $connections: [ID!]!
  ) {
    startDocumentSuggestion(input: $input) {
      documentSuggestionEdge @appendEdge(connections: $connections) {
        cursor
        node {
          id
          status
          ...DocumentSuggestionListItem_suggestion
        }
      }
    }
  }
`;

const suggestionAuthorQuery = graphql`
  query useSuggestionControllerQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on DocumentSuggestion {
        author { id }
      }
    }
  }
`;

// Must match the arguments of the review column's suggestions connection.
function openSuggestionsConnectionId(documentId: string) {
  return ConnectionHandler.getConnectionID(documentId, "DocumentReviewPanel_suggestions", {
    filter: { status: "OPEN" },
    orderBy: { field: "CREATED_AT", direction: "ASC" },
  });
}

export function useSuggestionController(
  documentId: string,
  mode: DocumentMode,
  onOpenSuggestion: (id: string) => void,
) {
  const { t } = useTranslation("organizations/documents");
  const toast = useToast();
  const environment = useRelayEnvironment();
  const profileId = useContext(CurrentUser).profile?.id;
  const [startSuggestion] = useMutation<useSuggestionControllerMutation>(
    startSuggestionMutation,
    { errorToast: false },
  );
  const controller = useRef<SuggestingController>({
    enabled: false,
    suggestionId: null,
    pending: null,
    request: null,
    onStartSuggesting: () => Promise.resolve(null),
    onSuggestionDone: () => {},
    onOpenSuggestion: () => {},
  });
  const editorRef = useRef<Editor | null>(null);
  const ownedIds = useRef(new Set<string>());
  const deniedIds = useRef(new Set<string>());
  const claims = useRef(new Map<string, Promise<boolean>>());
  const showBlockedToast = useCallback(() => {
    toast.add({
      title: t("details.editor.suggestionFailed"),
      description: t("details.editor.suggestionUntracked"),
      type: "error",
    });
  }, [t, toast]);
  const notifyBlocked = useDebounceCallback(
    showBlockedToast,
    blockedToastIntervalMs,
    blockedToastOptions,
  );

  useEffect(() => {
    controller.current.enabled = mode === "suggesting";
    if (mode !== "suggesting") {
      controller.current.suggestionId = null;
    }
  }, [mode]);

  useEffect(() => {
    const current = controller.current;
    current.ownsSuggestion = id => ownedIds.current.has(id);
    current.claimSuggestion = (id) => {
      if (ownedIds.current.has(id)) return Promise.resolve(true);
      if (deniedIds.current.has(id)) return Promise.resolve(false);
      const existing = claims.current.get(id);
      if (existing) return existing;

      const request = fetchQuery<useSuggestionControllerQuery>(
        environment,
        suggestionAuthorQuery,
        { id },
        { fetchPolicy: "network-only" },
      ).toPromise().then((data) => {
        if (!profileId) return false;
        const authorId = data?.node?.__typename === "DocumentSuggestion"
          ? data.node.author.id
          : null;
        if (authorId === profileId) {
          ownedIds.current.add(id);
          return true;
        }
        deniedIds.current.add(id);
        return false;
      }).catch(() => false).finally(() => {
        claims.current.delete(id);
      });
      claims.current.set(id, request);
      return request;
    };
    current.onSuggestionDone = () => {
      current.suggestionId = null;
    };
    current.onOpenSuggestion = onOpenSuggestion;
    current.onBlocked = notifyBlocked;
    current.onStartSuggesting = () => {
      if (current.suggestionId) {
        return Promise.resolve(current.suggestionId);
      }

      const localId = `local:${crypto.randomUUID()}`;
      current.suggestionId = localId;
      ownedIds.current.add(localId);

      const request = new Promise<string | null>((resolve) => {
        void startSuggestion({
          variables: {
            input: { documentId },
            connections: [openSuggestionsConnectionId(documentId)],
          },
          onCompleted: (response, errors) => {
            if (current.request === request) {
              current.request = null;
            }
            const editor = editorRef.current;
            if (errors?.length) {
              if (current.suggestionId === localId) {
                current.suggestionId = null;
              }
              ownedIds.current.delete(localId);
              if (editor) {
                discardSuggestionId(editor, localId);
              }
              toast.add({
                title: t("details.editor.suggestionFailed"),
                description: formatError(t("details.editor.suggestionFailed"), errors),
                type: "error",
              });
              resolve(null);
              return;
            }
            const id = response.startDocumentSuggestion.documentSuggestionEdge.node.id;
            ownedIds.current.add(id);
            if (editor) {
              reassignSuggestionId(editor, localId, id);
            }
            if (current.suggestionId === localId) {
              current.suggestionId = id;
            }
            resolve(id);
          },
          onError: (error) => {
            if (current.request === request) {
              current.request = null;
            }
            if (current.suggestionId === localId) {
              current.suggestionId = null;
            }
            ownedIds.current.delete(localId);
            const editor = editorRef.current;
            if (editor) {
              discardSuggestionId(editor, localId);
            }
            toast.add({
              title: t("details.editor.suggestionFailed"),
              description: error.message,
              type: "error",
            });
            resolve(null);
          },
        });
      });
      current.request = request;
      return request;
    };
  }, [documentId, environment, notifyBlocked, onOpenSuggestion, profileId, startSuggestion, t, toast]);

  const setEditor = useCallback((editor: Editor | null) => {
    editorRef.current = editor;
  }, []);

  return { controller, setEditor };
}
