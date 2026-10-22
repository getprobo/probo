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

import {
  connectRepoDocument,
  type RepoCollaborationHandle,
} from "@probo/collaboration";
import { formatError } from "@probo/helpers";
import { Button, CollaborativeRichEditor, IconClock, supportsRichEditorCollaboration } from "@probo/ui";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { useOutletContext } from "react-router";
import { graphql } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type { DocumentDescriptionPage_updateContentMutation } from "#/__generated__/core/DocumentDescriptionPage_updateContentMutation.graphql";
import type { DocumentDescriptionPageQuery } from "#/__generated__/core/DocumentDescriptionPageQuery.graphql";
import { RichEditor } from "#/pages/organizations/_components/RichEditor";

import { DocumentHistoryView } from "./_components/DocumentHistoryPanel";
import { DocumentPresence } from "./_components/DocumentPresence";

const autoSaveIntervalMs = 1000;

type CollaborationState = {
  versionID: string;
  attempt: number;
  handle?: RepoCollaborationHandle;
  failed?: boolean;
};

export const documentDescriptionPageQuery = graphql`
  query DocumentDescriptionPageQuery($documentId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    # We use this on /documents/:documentId/versions/:versionId/description
    version: node(id: $versionId) @include(if: $versionSpecified) {
      __typename
      ... on DocumentVersion {
        id
        content {
          prosemirror
        }
        status
      }
    }
    document: node(id: $documentId) {
      __typename
      ... on Document {
        id
        status
        writeMode
        content {
          prosemirror
        }
        canUpdate: permission(action: "core:document:update")
        canViewHistory: permission(action: "core:document-revision:list")
      }
    }
  }
`;

const updateContentMutation = graphql`
  mutation DocumentDescriptionPage_updateContentMutation($input: UpdateDocumentInput!) {
    updateDocument(input: $input) {
      document {
        id
      }
    }
  }
`;

export function DocumentDescriptionPage(props: {
  queryRef: PreloadedQuery<DocumentDescriptionPageQuery>;
  versionChangedAt: number;
}) {
  const { queryRef, versionChangedAt } = props;

  const { t } = useTranslation();
  const toast = useToast();
  const { isEditable } = useOutletContext<{
    isEditable: boolean;
  }>();

  const { document, version } = usePreloadedQuery<DocumentDescriptionPageQuery>(
    documentDescriptionPageQuery,
    queryRef,
  );
  if (document.__typename !== "Document" || (version && version.__typename !== "DocumentVersion")) {
    throw new Error("invalid type for node");
  }

  const snapshot = version ?? null;
  const contentSource = snapshot?.content.prosemirror ?? document.content.prosemirror;
  const editorIdentity = snapshot?.id ?? document.id;

  const [updateContent] = useMutation<DocumentDescriptionPage_updateContentMutation>(updateContentMutation);

  const documentId = document.id;

  const handleUpdate = useDebounceCallback(
    useCallback((content: string) => {
      updateContent({
        variables: {
          input: {
            id: documentId,
            content,
          },
        },
        onCompleted: (_data, errors) => {
          if (errors?.length) {
            toast.add({
              title: t("documentDescriptionPage.errors.title"),
              description: formatError(t("documentDescriptionPage.errors.save"), errors),
              type: "error",
            });
            return;
          }

          toast.add({
            title: t("documentDescriptionPage.messages.successTitle"),
            description: t("documentDescriptionPage.messages.saved"),
            type: "success",
          });
        },
        onError: (error) => {
          toast.add({
            title: t("documentDescriptionPage.errors.title"),
            description: error.message ?? t("documentDescriptionPage.errors.save"),
            type: "error",
          });
        },
      });
    }, [documentId, updateContent, toast, t]),
    autoSaveIntervalMs,
  );
  const uploadingRef = useRef(false);
  const handleContentChange = useCallback((content: string) => {
    if (uploadingRef.current) {
      handleUpdate.cancel();
      return;
    }

    handleUpdate(content);
  }, [handleUpdate]);

  const canEdit = isEditable
    && document.canUpdate
    && document.status !== "ARCHIVED"
    && document.writeMode !== "GENERATED";
  const content = contentSource;
  const collaborationSupported = canEdit
    && supportsRichEditorCollaboration(content);
  const [collaboration, setCollaboration] = useState<CollaborationState>();
  const connectionAttempt = useRef(0);

  useEffect(() => {
    if (!collaborationSupported) return;

    const attempt = ++connectionAttempt.current;
    let activeHandle: RepoCollaborationHandle | undefined;

    connectRepoDocument(document.id)
      .then((handle) => {
        if (attempt !== connectionAttempt.current) {
          handle.close();
          return;
        }

        activeHandle = handle;
        setCollaboration({
          versionID: document.id,
          attempt,
          handle,
        });
      })
      .catch((error) => {
        if (attempt !== connectionAttempt.current) return;

        setCollaboration({
          versionID: document.id,
          attempt,
          failed: true,
        });
        toast.add({
          title: t("documentDescriptionPage.errors.title"),
          description: error instanceof Error
            ? error.message
            : t("documentDescriptionPage.errors.save"),
          type: "error",
        });
      });

    return () => {
      connectionAttempt.current += 1;
      activeHandle?.close();
      setCollaboration(current =>
        current?.attempt === attempt ? undefined : current,
      );
    };
  }, [collaborationSupported, document.id, t, toast]);

  const collaborationConnecting = collaborationSupported
    && (collaboration?.versionID !== document.id
      || (!collaboration?.handle && !collaboration?.failed));
  const collaborationHandle = collaborationSupported
    && collaboration?.versionID === document.id
    && !collaboration.failed
    ? collaboration.handle
    : undefined;

  // The editor key must change on explicit actions (delete draft, edit
  // title/type) but NOT on auto-save side effects (cursor preservation).
  // We track a "data generation" that only increments when an explicit
  // action (versionChangedAt change) is followed by fresh data arriving
  // (currentVersion.id change). This uses React's "adjust state during
  // render" pattern so we avoid refs-during-render and setState-in-effects.
  const [prevVCA, setPrevVCA] = useState(versionChangedAt);
  const [prevVersionId, setPrevVersionId] = useState(editorIdentity);
  const [dataGeneration, setDataGeneration] = useState(0);
  const [pendingExplicit, setPendingExplicit] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);

  if (versionChangedAt !== prevVCA) {
    setPrevVCA(versionChangedAt);
    if (editorIdentity !== prevVersionId) {
      // Both changed at once — data was already available.
      setPrevVersionId(editorIdentity);
      setDataGeneration(g => g + 1);
      setPendingExplicit(false);
    } else {
      // Explicit action fired but data hasn't arrived yet.
      setPendingExplicit(true);
    }
  } else if (editorIdentity !== prevVersionId) {
    setPrevVersionId(editorIdentity);
    if (pendingExplicit) {
      // Fresh data arrived for a pending explicit action — remount.
      setDataGeneration(g => g + 1);
      setPendingExplicit(false);
    }
    // Otherwise auto-save changed the version — don't bump generation.
  }

  const editorKey = `${editorIdentity}-${dataGeneration}`;
  const historyAvailable = document.canViewHistory
    && supportsRichEditorCollaboration(content);

  if (historyOpen) {
    return (
      <div className="flex flex-1 min-h-0 min-w-0">
        <Suspense
          fallback={(
            <div className="flex-1 px-8 py-6 text-sm text-txt-secondary">
              {t("documentDescriptionPage.history.loading")}
            </div>
          )}
        >
          <DocumentHistoryView
            documentId={documentId}
            canRestore={canEdit}
            onClose={() => setHistoryOpen(false)}
            onRestored={() => setHistoryOpen(false)}
          />
        </Suspense>
      </div>
    );
  }

  return (
    <div className="flex flex-1 min-h-0 flex-col">
      {(historyAvailable || collaborationHandle) && (
        <div className="flex items-center justify-end gap-3 px-4 pb-2">
          {collaborationHandle && (
            <DocumentPresence collaborationHandle={collaborationHandle} />
          )}
          {historyAvailable && (
            <Button
              icon={IconClock}
              variant="secondary"
              onClick={() => setHistoryOpen(true)}
            >
              {t("documentDescriptionPage.history.open")}
            </Button>
          )}
        </div>
      )}
      {collaborationHandle
        ? (
            <CollaborativeRichEditor
              key={editorKey}
              className="flex-1"
              collaborationHandle={collaborationHandle}
              content={content}
              data-theme="document"
              disabled={!canEdit || collaborationConnecting}
              onChangeContent={handleContentChange}
              onUploadingChange={(uploading) => {
                uploadingRef.current = uploading;
                if (uploading) {
                  handleUpdate.cancel();
                }
              }}
            />
          )
        : (
            <RichEditor
              key={editorKey}
              className="flex-1"
              content={content}
              data-theme="document"
              disabled={!canEdit || collaborationConnecting}
              onChangeContent={handleContentChange}
              onUploadingChange={(uploading) => {
                uploadingRef.current = uploading;
                if (uploading) {
                  handleUpdate.cancel();
                }
              }}
            />
          )}
    </div>
  );
}
