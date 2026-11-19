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
import { type CSSProperties, useCallback, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment } from "react-relay";

import type { DocumentEditor_document$key } from "#/__generated__/core/DocumentEditor_document.graphql";
import type { DocumentReviewListQuery } from "#/__generated__/core/DocumentReviewListQuery.graphql";
import { useUploadRichTextFile } from "#/pages/organizations/_lib/useUploadRichTextFile";

import {
  documentPageLayout,
  documentPageMargin,
  documentPagesHeight,
  documentPageSize,
} from "../../_lib/documentPageLayout";
import { useDocumentCollaboration } from "../../_lib/useDocumentCollaboration";
import type { DocumentMode } from "../../_lib/useDocumentModes";
import { usePageScale } from "../../_lib/usePageScale";
import { usePreventSuggestionTextDrop } from "../../_lib/usePreventSuggestionTextDrop";
import { useSuggestionController } from "../../_lib/useSuggestionController";
import { documentEditor } from "../../variants";
import { DocumentModeSelect } from "../DocumentModeSelect";
import { DocumentReviewList } from "../DocumentReviewList/DocumentReviewList";

import { DocumentEditorPages } from "./DocumentEditorPages";
import { DocumentPresenceAvatars } from "./DocumentPresenceAvatars";

const documentEditorFragment = graphql`
  fragment DocumentEditor_document on Document {
    id
    orientation
  }
`;

interface DocumentEditorProps {
  documentKey: DocumentEditor_document$key;
  reviewQueryRef: PreloadedQuery<DocumentReviewListQuery> | null;
  mode: DocumentMode;
  modes: readonly DocumentMode[];
  onModeChange: (mode: DocumentMode) => void;
}

export function DocumentEditor({
  documentKey,
  reviewQueryRef,
  mode,
  modes,
  onModeChange,
}: DocumentEditorProps) {
  const document = useFragment(documentEditorFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const [activeKey, setActiveKey] = useState<string | null>(null);
  const [editor, setEditorInstance] = useState<Editor | null>(null);
  const [pageCount, setPageCount] = useState(1);
  const [toolbarEl, setToolbarEl] = useState<HTMLDivElement | null>(null);
  const { handle, failed } = useDocumentCollaboration(document.id);
  const { controller, setEditor } = useSuggestionController(document.id, mode, setActiveKey);
  const commentAnchorController: CommentAnchorController = useRef({
    onOpenThread: (id: string) => setActiveKey(id),
  });
  const onUploadFile = useUploadRichTextFile();
  const paperRef = useRef<HTMLDivElement>(null);
  const pageSize = documentPageSize(document.orientation);
  const { ref: pagesRef, scale } = usePageScale(pageSize.width);
  const pagination = useMemo(
    () => ({
      layout: documentPageLayout(documentPageSize(document.orientation)),
      onPages: setPageCount,
    }),
    [document.orientation],
  );
  const collaborating = mode !== "viewing";
  const {
    toolbar, toolbarControls, toolbarEnd, toolbarMode,
    frame, viewport, content, pages, stack, paper, review, scrollbar, thumb,
  } = documentEditor({ review: collaborating, presence: collaborating });
  usePreventSuggestionTextDrop(paperRef, mode);

  const onEditor = useCallback((next: Editor | null) => {
    setEditor(next);
    setEditorInstance(next);
  }, [setEditor]);

  return (
    <>
      <div className={toolbar()}>
        <div ref={setToolbarEl} className={toolbarControls()} />
        <div className={toolbarEnd()}>
          {collaborating && handle != null && (
            <DocumentPresenceAvatars collaborationHandle={handle} />
          )}
          <div className={toolbarMode()}>
            <DocumentModeSelect mode={mode} modes={modes} onModeChange={onModeChange} />
          </div>
        </div>
      </div>
      <ScrollArea.Root className={frame()}>
        <ScrollArea.Viewport className={viewport()}>
          <ScrollArea.Content
            className={content()}
            style={{ "--page-width": `${pageSize.width}px` } as CSSProperties}
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
                  {failed
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
                            disabled={mode === "viewing"}
                            suggesting={controller}
                            commentAnchorController={commentAnchorController}
                            pagination={pagination}
                            toolbarContainer={toolbarEl}
                            onEditor={onEditor}
                            onUploadFile={onUploadFile}
                          />
                        )}
                </div>
              </div>
            </div>
            {collaborating && (
              <div className={review()}>
                {handle != null && (
                  <DocumentReviewList
                    queryRef={reviewQueryRef}
                    editor={editor}
                    collaborationHandle={handle}
                    activeKey={activeKey}
                    onActiveChange={setActiveKey}
                  />
                )}
              </div>
            )}
          </ScrollArea.Content>
        </ScrollArea.Viewport>
        <ScrollArea.Scrollbar className={scrollbar()} orientation="vertical">
          <ScrollArea.Thumb className={thumb()} />
        </ScrollArea.Scrollbar>
      </ScrollArea.Root>
    </>
  );
}
