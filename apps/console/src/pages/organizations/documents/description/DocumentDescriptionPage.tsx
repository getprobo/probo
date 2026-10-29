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

import { CaretDownIcon, PencilSimpleIcon } from "@phosphor-icons/react";
import {
  connectRepoDocument,
  type RepoCollaborationHandle,
} from "@probo/collaboration";
import { formatError } from "@probo/helpers";
import { Button, IconClock } from "@probo/ui";
import {
  CollaborativeRichEditor,
  supportsRichEditorCollaboration,
} from "@probo/ui/src/RichEditor/CollaborativeRichEditor";
import { type CommentAnchor } from "@probo/ui/src/RichEditor/commentAnchors";
import { type SuggestingController } from "@probo/ui/src/RichEditor/SuggestionExtension";
import { Button as ButtonV2 } from "@probo/ui/src/v2/Button/Button";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownRadioGroup } from "@probo/ui/src/v2/Dropdown/DropdownRadioGroup";
import { DropdownRadioItem } from "@probo/ui/src/v2/Dropdown/DropdownRadioItem";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import type { Editor } from "@tiptap/core";
import { Suspense, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { useOutletContext } from "react-router";
import { graphql } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type { DocumentDescriptionPage_updateContentMutation } from "#/__generated__/core/DocumentDescriptionPage_updateContentMutation.graphql";
import type { DocumentDescriptionPageQuery } from "#/__generated__/core/DocumentDescriptionPageQuery.graphql";
import type { DocumentHistoryPanelQuery } from "#/__generated__/core/DocumentHistoryPanelQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";
import { RichEditor } from "#/pages/organizations/_components/RichEditor";

import { documentHistoryPanelQuery, DocumentHistoryView } from "./_components/DocumentHistoryPanel";
import { DocumentPresence } from "./_components/DocumentPresence";
import { DocumentReviewMargin } from "./_components/DocumentReviewMargin/DocumentReviewMargin";

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
        canComment: permission(action: "core:document:comment")
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

  const [updateContent] = useMutation<DocumentDescriptionPage_updateContentMutation>(
    updateContentMutation,
    { errorToast: false },
  );

  const documentId = document.id;

  const handleUpdate = useDebounceCallback(
    useCallback((content: string) => {
      void updateContent({
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
  const canReview = isEditable && (document.canUpdate || document.canComment);
  const content = contentSource;
  const collaborationSupported = canReview
    && supportsRichEditorCollaboration(content);
  const [mode, setMode] = useState<"viewing" | "editing" | "suggesting">("viewing");
  const [editor, setEditor] = useState<Editor | null>(null);
  const reviewToolbarRef = useRef<HTMLDivElement>(null);
  const [reviewMarginHost, setReviewMarginHost] = useState<HTMLElement | null>(null);
  const [activeThreadId, setActiveThreadId] = useState<string | null>(null);
  const [activeSuggestionId, setActiveSuggestionId] = useState<string | null>(null);
  const [reviewThreads, setReviewThreads] = useState<Array<{
    id: string;
    anchorStart: string;
    anchorEnd: string;
    suggestionId: string | null;
  }>>([]);
  const [reviewVersion, setReviewVersion] = useState(0);
  const ignoreOwnReview = useRef(false);
  const suggestingRef = useRef<SuggestingController>({
    enabled: false,
    suggestionId: null,
    pending: null,
    request: null,
    onNeedId: () => Promise.resolve(null),
    onBurstFinished: () => {},
    onOpenSuggestion: () => {},
  });
  const anchorController = useRef({ onOpenThread: (_id: string) => {} });
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

  useEffect(() => {
    suggestingRef.current.enabled = mode === "suggesting";
    suggestingRef.current.onOpenSuggestion = (id: string) => {
      setActiveSuggestionId(id);
      setActiveThreadId(null);
    };
    anchorController.current.onOpenThread = (id: string) => {
      if (id === "draft") return;
      setActiveThreadId(id);
      setActiveSuggestionId(null);
    };
  }, [mode]);

  useEffect(() => {
    let timer = 0;
    const unsubscribe = collaborationHandle?.onReview?.(() => {
      window.clearTimeout(timer);
      timer = window.setTimeout(() => {
        if (ignoreOwnReview.current) {
          ignoreOwnReview.current = false;
          return;
        }
        setReviewVersion(version => version + 1);
      }, 50);
    });
    return () => {
      window.clearTimeout(timer);
      unsubscribe?.();
    };
  }, [collaborationHandle]);

  const commentAnchors = useMemo<CommentAnchor[]>(() => reviewThreads
    .filter(thread => thread.suggestionId == null)
    .map(thread => ({
      id: thread.id,
      anchorStart: thread.anchorStart,
      anchorEnd: thread.anchorEnd,
      active: thread.id === activeThreadId || thread.id === "draft",
    })), [activeThreadId, reviewThreads]);

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
  const [historyQueryRef, loadHistoryQuery]
    = useQueryLoader<DocumentHistoryPanelQuery>(documentHistoryPanelQuery);

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
        {historyQueryRef
          ? (
              <Suspense
                fallback={(
                  <div className="flex-1 px-8 py-6 text-sm text-txt-secondary">
                    {t("documentDescriptionPage.history.loading")}
                  </div>
                )}
              >
                <DocumentHistoryView
                  queryRef={historyQueryRef}
                  documentId={documentId}
                  canRestore={canEdit}
                  onClose={() => setHistoryOpen(false)}
                  onRestored={() => setHistoryOpen(false)}
                />
              </Suspense>
            )
          : (
              <div className="flex-1 px-8 py-6 text-sm text-txt-secondary">
                {t("documentDescriptionPage.history.loading")}
              </div>
            )}
      </div>
    );
  }

  return (
    <div className="flex flex-1 min-h-0 flex-col">
      {(historyAvailable || collaborationHandle) && (
        <div className="flex items-center justify-end gap-3 px-4 pb-2">
          {collaborationHandle && document.canComment && (
            <Dropdown>
              <DropdownTrigger
                render={(
                  <ButtonV2
                    variant="ghost"
                    color={mode === "suggesting" ? "green" : "neutral"}
                    size={2}
                    iconStart={<PencilSimpleIcon />}
                    iconEnd={<CaretDownIcon />}
                  >
                    {mode === "viewing"
                      ? t("documentDescriptionPage.review.viewing")
                      : mode === "suggesting"
                        ? t("documentDescriptionPage.review.suggesting")
                        : t("documentDescriptionPage.review.editing")}
                  </ButtonV2>
                )}
              />
              <DropdownPopup align="end">
                <DropdownRadioGroup
                  value={mode}
                  onValueChange={(value: string) => {
                    if (value === "viewing" || value === "editing" || value === "suggesting") setMode(value);
                  }}
                >
                  <DropdownRadioItem value="viewing">
                    {t("documentDescriptionPage.review.viewing")}
                  </DropdownRadioItem>
                  {canEdit && (
                    <DropdownRadioItem value="editing">
                      {t("documentDescriptionPage.review.editing")}
                    </DropdownRadioItem>
                  )}
                  <DropdownRadioItem value="suggesting">
                    {t("documentDescriptionPage.review.suggesting")}
                  </DropdownRadioItem>
                </DropdownRadioGroup>
              </DropdownPopup>
            </Dropdown>
          )}
          <div ref={reviewToolbarRef} className="flex items-center" />
          {collaborationHandle && (
            <DocumentPresence collaborationHandle={collaborationHandle} />
          )}
          {historyAvailable && (
            <Button
              icon={IconClock}
              variant="secondary"
              onClick={() => {
                loadHistoryQuery({ documentId }, { fetchPolicy: "network-only" });
                setHistoryOpen(true);
              }}
            >
              {t("documentDescriptionPage.history.open")}
            </Button>
          )}
        </div>
      )}
      <div className="flex min-h-0 flex-1" data-review-layout="">
        {collaborationHandle
          ? (
              <>
                <CollaborativeRichEditor
                  key={editorKey}
                  className="flex-1"
                  collaborationHandle={collaborationHandle}
                  margin={editor
                    ? (
                        <Suspense fallback={null}>
                          <DocumentReviewMargin
                            activeSuggestionId={activeSuggestionId}
                            activeThreadId={activeThreadId}
                            canComment={document.canComment}
                            canUpdate={document.canUpdate}
                            collaborationHandle={collaborationHandle}
                            documentId={documentId}
                            editor={editor}
                            mode={mode}
                            onActiveSuggestion={setActiveSuggestionId}
                            onActiveThread={setActiveThreadId}
                            onThreads={setReviewThreads}
                            onOwnReview={() => {
                              ignoreOwnReview.current = true;
                            }}
                            onOwnReviewFailed={() => {
                              ignoreOwnReview.current = false;
                            }}
                            reviewVersion={reviewVersion}
                            suggesting={suggestingRef}
                            toolbarHost={reviewToolbarRef}
                            marginHost={reviewMarginHost}
                          />
                        </Suspense>
                      )
                    : null}
                  commentAnchorController={anchorController}
                  commentAnchors={commentAnchors}
                  content={content}
                  data-theme="document"
                  disabled={collaborationConnecting || mode === "viewing" || (!canEdit && mode !== "suggesting")}
                  formatting={canEdit && mode === "editing"}
                  onChangeContent={handleContentChange}
                  onEditor={setEditor}
                  onUploadingChange={(uploading) => {
                    uploadingRef.current = uploading;
                    if (uploading) {
                      handleUpdate.cancel();
                    }
                  }}
                  suggesting={suggestingRef}
                />
                <aside
                  ref={setReviewMarginHost}
                  className="relative ml-4 hidden w-68 shrink-0 bg-sand-2 lg:block"
                  data-review-margin=""
                />
              </>
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
    </div>
  );
}
