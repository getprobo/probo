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
import { useCallback, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment } from "react-relay";

import type { DocumentEditor_document$key } from "#/__generated__/core/DocumentEditor_document.graphql";
import type { DocumentReviewListQuery } from "#/__generated__/core/DocumentReviewListQuery.graphql";
import { useUploadRichTextFile } from "#/pages/organizations/_lib/useUploadRichTextFile";

import { useDocumentCollaboration } from "../_lib/useDocumentCollaboration";
import type { DocumentMode } from "../_lib/useDocumentModes";
import { usePreventSuggestionTextDrop } from "../_lib/usePreventSuggestionTextDrop";
import { useSuggestionController } from "../_lib/useSuggestionController";
import { documentEditor } from "../variants";

import { DocumentReviewList } from "./DocumentReviewList/DocumentReviewList";

const documentEditorFragment = graphql`
  fragment DocumentEditor_document on Document {
    id
  }
`;

interface DocumentEditorProps {
  documentKey: DocumentEditor_document$key;
  reviewQueryRef: PreloadedQuery<DocumentReviewListQuery> | null;
  mode: DocumentMode;
}

export function DocumentEditor({ documentKey, reviewQueryRef, mode }: DocumentEditorProps) {
  const document = useFragment(documentEditorFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const [activeKey, setActiveKey] = useState<string | null>(null);
  const [editor, setEditorInstance] = useState<Editor | null>(null);
  const { handle, failed } = useDocumentCollaboration(document.id);
  const { controller, setEditor } = useSuggestionController(document.id, mode, setActiveKey);
  const commentAnchorController: CommentAnchorController = useRef({
    onOpenThread: (id: string) => setActiveKey(id),
  });
  const onUploadFile = useUploadRichTextFile();
  const paperRef = useRef<HTMLDivElement>(null);
  const { frame, viewport, content, paper, review, scrollbar, thumb } = documentEditor();
  usePreventSuggestionTextDrop(paperRef, mode);

  const onEditor = useCallback((next: Editor | null) => {
    setEditor(next);
    setEditorInstance(next);
  }, [setEditor]);

  return (
    <ScrollArea.Root className={frame()}>
      <ScrollArea.Viewport className={viewport()}>
        <ScrollArea.Content className={content()}>
          <div ref={paperRef} className={paper()} data-theme="document">
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
                      onEditor={onEditor}
                      onUploadFile={onUploadFile}
                    />
                  )}
          </div>
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
        </ScrollArea.Content>
      </ScrollArea.Viewport>
      <ScrollArea.Scrollbar className={scrollbar()} orientation="vertical">
        <ScrollArea.Thumb className={thumb()} />
      </ScrollArea.Scrollbar>
    </ScrollArea.Root>
  );
}
