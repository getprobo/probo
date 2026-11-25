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

import { ScrollArea } from "@base-ui/react/scroll-area";
import { CollaborativeRichEditor } from "@probo/ui/src/RichEditor/CollaborativeRichEditor";
import type { CommentAnchorController } from "@probo/ui/src/RichEditor/commentAnchors";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { Editor } from "@tiptap/core";
import { type CSSProperties, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, useQueryLoader } from "react-relay";

import type { DocumentEditor_document$key } from "#/__generated__/core/DocumentEditor_document.graphql";
import type { DocumentReviewListQuery } from "#/__generated__/core/DocumentReviewListQuery.graphql";
import type { DocumentVersionContentQuery } from "#/__generated__/core/DocumentVersionContentQuery.graphql";
import type { DocumentVersionListQuery } from "#/__generated__/core/DocumentVersionListQuery.graphql";
import type { DocumentViewContentQuery } from "#/__generated__/core/DocumentViewContentQuery.graphql";
import { useUploadRichTextFile } from "#/pages/organizations/_lib/useUploadRichTextFile";

import type { DocumentCommentDraft } from "../../_lib/documentCommentDraft";
import {
  documentPageLayout,
  documentPageMargin,
  documentPagesHeight,
  documentPageSize,
} from "../../_lib/documentPageLayout";
import { type DocumentColumn, withDocumentColumn } from "../../_lib/documentSearch";
import { useDocumentCollaboration } from "../../_lib/useDocumentCollaboration";
import type { DocumentMode } from "../../_lib/useDocumentModes";
import { useDocumentSearch } from "../../_lib/useDocumentSearch";
import { useElementHeight } from "../../_lib/useElementHeight";
import { usePageScale } from "../../_lib/usePageScale";
import { usePreventSuggestionTextDrop } from "../../_lib/usePreventSuggestionTextDrop";
import { useSuggestionController } from "../../_lib/useSuggestionController";
import { documentEditor } from "../../variants";
import { DocumentReviewList } from "../DocumentReviewList/DocumentReviewList";
import type { DocumentReviewView } from "../DocumentReviewList/DocumentReviewViewSwitch";
import { DocumentVersionContent } from "../DocumentVersionList/DocumentVersionContent";
import { DocumentVersionList, documentVersionListQuery } from "../DocumentVersionList/DocumentVersionList";

import { DocumentCommentBubbleItem } from "./DocumentCommentBubbleItem";
import { DocumentEditorPages } from "./DocumentEditorPages";
import { DocumentEditorToolbar } from "./DocumentEditorToolbar";
import { DocumentViewContent, documentViewContentQuery } from "./DocumentViewContent";

const viewRefetchDelayMs = 1_000;

const documentEditorFragment = graphql`
  fragment DocumentEditor_document on Document {
    id
    status
    orientation
    canComment: permission(action: "core:document:comment")
  }
`;

interface DocumentEditorProps {
  documentKey: DocumentEditor_document$key;
  reviewQueryRef: PreloadedQuery<DocumentReviewListQuery> | null;
  // Set while a previous version replaces the live document, in viewing only.
  versionId: string | null;
  versionQueryRef: PreloadedQuery<DocumentVersionContentQuery> | null;
  onVersionChange: (versionId: string | null) => void;
  mode: DocumentMode;
  modes: readonly DocumentMode[];
  onModeChange: (mode: DocumentMode) => void;
}

export function DocumentEditor({
  documentKey,
  reviewQueryRef,
  versionId,
  versionQueryRef,
  onVersionChange,
  mode,
  modes,
  onModeChange,
}: DocumentEditorProps) {
  const document = useFragment(documentEditorFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const [activeKey, setActiveKey] = useState<string | null>(null);
  const [commentDraft, setCommentDraft] = useState<DocumentCommentDraft | null>(null);
  const [draftMode, setDraftMode] = useState(mode);
  if (draftMode !== mode) {
    setDraftMode(mode);
    setCommentDraft(null);
  }
  // Opening another card drops the comment being written.
  const activate = useCallback((key: string | null) => {
    setActiveKey(key);
    if (key != null) {
      setCommentDraft(null);
    }
  }, []);
  const { column, update } = useDocumentSearch();
  const selectColumn = useCallback(
    (next: DocumentColumn) => update(params => withDocumentColumn(params, next)),
    [update],
  );
  const reviewView: DocumentReviewView = column === "all" ? "all" : "open";
  // Comment and suggestion cards being written only show in the open view.
  const showOpenReview = useCallback(
    () => update(params => withDocumentColumn(params, "open"), { replace: true }),
    [update],
  );
  const startCommentDraft = useCallback((draft: DocumentCommentDraft) => {
    setActiveKey(null);
    setCommentDraft(draft);
    selectColumn("open");
  }, [selectColumn]);
  const [editor, setEditorInstance] = useState<Editor | null>(null);
  const [pageCount, setPageCount] = useState(1);
  const [toolbarEl, setToolbarEl] = useState<HTMLDivElement | null>(null);
  const { handle, failed } = useDocumentCollaboration(
    document.id,
    mode,
    document.status !== "ARCHIVED",
  );
  // A card opened under its local id stays open once the real id replaces it.
  const settleActiveSuggestion = useCallback((localId: string, id: string) => {
    setActiveKey(current => (current === localId ? id : current));
  }, []);
  const { controller, setEditor } = useSuggestionController(
    document.id,
    mode,
    activate,
    showOpenReview,
    settleActiveSuggestion,
  );
  const commentAnchorController: CommentAnchorController = useRef({
    onOpenThread: (id: string) => activate(id),
  });
  const onUploadFile = useUploadRichTextFile();
  const paperRef = useRef<HTMLDivElement>(null);
  const frameRef = useRef<HTMLDivElement>(null);
  const [reviewElement, setReviewElement] = useState<HTMLDivElement | null>(null);
  const [switchContainer, setSwitchContainer] = useState<HTMLDivElement | null>(null);
  const [switchBox, setSwitchBox] = useState<{ left: number; width: number } | null>(null);
  const [versionListQueryRef, loadVersionList]
    = useQueryLoader<DocumentVersionListQuery>(documentVersionListQuery);
  const [versionOrientation, setVersionOrientation] = useState<string | null>(null);
  const viewing = mode === "viewing";
  const viewingVersion = viewing && versionId != null;
  const showingVersion = viewingVersion && versionQueryRef != null;
  const showVersions = column === "versions";
  const versionListLoaded = versionListQueryRef != null;
  useEffect(() => {
    if (showVersions && !versionListLoaded) {
      loadVersionList({ documentId: document.id }, { fetchPolicy: "store-and-network" });
    }
  }, [document.id, loadVersionList, showVersions, versionListLoaded]);
  const [viewQueryRef, loadView, disposeView]
    = useQueryLoader<DocumentViewContentQuery>(documentViewContentQuery);
  const viewingLive = viewing && versionId == null;
  // The server persists a change after peers receive it, so the refetch
  // trails the last change instead of following each one.
  useEffect(() => {
    if (!viewingLive) {
      disposeView();
      return;
    }
    const load = () => loadView({ documentId: document.id }, { fetchPolicy: "store-and-network" });
    load();
    if (handle == null) {
      return;
    }
    let timer: ReturnType<typeof setTimeout> | undefined;
    const onChange = () => {
      clearTimeout(timer);
      timer = setTimeout(load, viewRefetchDelayMs);
    };
    handle.on("change", onChange);
    return () => {
      clearTimeout(timer);
      handle.off("change", onChange);
    };
  }, [disposeView, document.id, handle, loadView, viewingLive]);
  const pageSize = documentPageSize(
    showingVersion && versionOrientation != null ? versionOrientation : document.orientation,
  );
  const { ref: pagesRef, scale } = usePageScale(pageSize.width);
  // Caps the review history, which scrolls on its own within the frame.
  const { ref: viewportRef, height: viewportHeight } = useElementHeight();
  const pagination = useMemo(
    () => ({
      layout: documentPageLayout(documentPageSize(document.orientation)),
      onPages: setPageCount,
    }),
    [document.orientation],
  );
  const collaborating = !viewing;
  const showReview = collaborating && !showVersions;
  const {
    frame, viewport, content, pages, stack, paper, review,
    reviewSwitch, reviewSwitchTarget, scrollbar, thumb,
  } = documentEditor({ review: collaborating || showVersions });

  // The switch sits on the frame, so it follows the review column's box
  // instead of scrolling inside the faded viewport.
  useEffect(() => {
    const frameElement = frameRef.current;
    if (reviewElement == null || frameElement == null) {
      return;
    }
    const observer = new ResizeObserver(() => {
      const frameRect = frameElement.getBoundingClientRect();
      const rect = reviewElement.getBoundingClientRect();
      const style = getComputedStyle(reviewElement);
      const paddingLeft = Number.parseFloat(style.paddingLeft) || 0;
      const paddingRight = Number.parseFloat(style.paddingRight) || 0;
      const left = rect.left - frameRect.left + paddingLeft;
      const width = Math.max(0, rect.width - paddingLeft - paddingRight);
      setSwitchBox(current =>
        current != null && current.left === left && current.width === width
          ? current
          : { left, width },
      );
    });
    observer.observe(reviewElement);
    observer.observe(frameElement);
    return () => observer.disconnect();
  }, [reviewElement]);
  usePreventSuggestionTextDrop(paperRef, mode);

  const onEditor = useCallback((next: Editor | null) => {
    setEditor(next);
    setEditorInstance(next);
  }, [setEditor]);

  const canComment = document.canComment;
  const bubbleMenu = useCallback(
    (current: Editor) => handle != null && canComment && (
      <DocumentCommentBubbleItem
        editor={current}
        collaborationHandle={handle}
        onComment={startCommentDraft}
      />
    ),
    [canComment, handle, startCommentDraft],
  );

  const onVersionMissing = useCallback(() => onVersionChange(null), [onVersionChange]);

  const showLatest = useCallback(() => onVersionChange(null), [onVersionChange]);

  return (
    <>
      <DocumentEditorToolbar
        mode={mode}
        modes={modes}
        onModeChange={onModeChange}
        collaborationHandle={handle}
        onControlsElement={setToolbarEl}
        versionsOpen={showVersions}
        onVersionsOpenChange={open => selectColumn(open ? "versions" : "open")}
        versionQueryRef={viewing ? versionQueryRef : null}
        onShowLatest={showLatest}
      />
      <ScrollArea.Root ref={frameRef} className={frame()}>
        <ScrollArea.Viewport ref={viewportRef} className={viewport()}>
          <ScrollArea.Content
            className={content()}
            style={{
              "--page-width": `${pageSize.width}px`,
              "--review-viewport-height": viewportHeight > 0 ? `${viewportHeight}px` : undefined,
            } as CSSProperties}
          >
            <div ref={pagesRef} className={pages()}>
              <div
                className={stack()}
                style={{ "width": pageSize.width, "zoom": scale, "--page-scale": scale } as CSSProperties}
              >
                <DocumentEditorPages count={pageCount} size={pageSize} />
                <div
                  ref={paperRef}
                  className={paper()}
                  data-theme="document"
                  style={{
                    padding: documentPageMargin,
                    minHeight: documentPagesHeight(pageSize, pageCount),
                  }}
                >
                  {showingVersion
                    ? (
                        <DocumentVersionContent
                          queryRef={versionQueryRef}
                          documentId={document.id}
                          onPages={setPageCount}
                          onOrientation={setVersionOrientation}
                          onMissing={onVersionMissing}
                        />
                      )
                    : viewingVersion
                      ? <Text size={2} color="faint">{t("details.editor.loading")}</Text>
                      : viewing
                        ? viewQueryRef != null && (
                          <DocumentViewContent queryRef={viewQueryRef} onPages={setPageCount} />
                        )
                        : failed
                          ? <Text size={2} color="faint">{t("details.editor.openError")}</Text>
                          : handle == null
                            ? <Text size={2} color="faint">{t("details.editor.loading")}</Text>
                            : (
                                <CollaborativeRichEditor
                                  key={document.id}
                                  className="w-full"
                                  collaborationHandle={handle}
                                  content=""
                                  page={false}
                                  suggesting={controller}
                                  commentAnchorController={commentAnchorController}
                                  pagination={pagination}
                                  toolbarContainer={toolbarEl}
                                  bubbleMenu={bubbleMenu}
                                  onEditor={onEditor}
                                  onUploadFile={onUploadFile}
                                />
                              )}
                </div>
              </div>
            </div>
            {showReview && (
              <div ref={setReviewElement} className={review()}>
                {handle != null && (
                  <DocumentReviewList
                    queryRef={reviewQueryRef}
                    editor={editor}
                    collaborationHandle={handle}
                    switchContainer={switchContainer}
                    view={reviewView}
                    onViewChange={selectColumn}
                    activeKey={activeKey}
                    onActiveChange={activate}
                    commentDraft={commentDraft}
                    onCommentDraftChange={setCommentDraft}
                  />
                )}
              </div>
            )}
            {showVersions && (
              <div className={review()}>
                <DocumentVersionList
                  queryRef={versionListQueryRef}
                  activeVersionId={viewing ? versionId : null}
                  onSelect={onVersionChange}
                  onClose={() => selectColumn("open")}
                />
              </div>
            )}
          </ScrollArea.Content>
        </ScrollArea.Viewport>
        {showReview && switchBox != null && switchBox.width > 0 && (
          <div
            className={reviewSwitch()}
            style={{ left: switchBox.left, width: switchBox.width }}
          >
            <div ref={setSwitchContainer} className={reviewSwitchTarget()} />
          </div>
        )}
        <ScrollArea.Scrollbar className={scrollbar()} orientation="vertical">
          <ScrollArea.Thumb className={thumb()} />
        </ScrollArea.Scrollbar>
      </ScrollArea.Root>
    </>
  );
}
