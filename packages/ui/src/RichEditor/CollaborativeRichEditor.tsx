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

import { type Editor } from "@tiptap/core";
import { type ComponentProps, type MutableRefObject, useCallback, useEffect, useMemo, useRef } from "react";

import {
  createRichEditorAutomergeDocument as createAutomergeDocument,
  createRichEditorCollaborationExtension,
  richEditorAutomergeContent,
  supportsRichEditorCollaboration,
} from "./collaboration";
import {
  type CommentAnchor,
  createCommentAnchorExtension,
  refreshCommentAnchors,
  type CommentAnchorController,
} from "./commentAnchors";
import { createRichEditorHistoryExtension } from "./historyDecorations";
import {
  createRichEditorPresenceExtension,
  type RichEditorCollaborationHandle,
} from "./presence";
import { RichEditor, richEditorCollaborationExtensions } from "./RichEditor";
import {
  createSuggestingExtension,
  type SuggestingController,
} from "./SuggestionExtension";

export { supportsRichEditorCollaboration };

export function createRichEditorAutomergeDocument(content: string) {
  return createAutomergeDocument(content, richEditorCollaborationExtensions);
}

type CollaborativeRichEditorProps = Omit<
  ComponentProps<typeof RichEditor>,
  "editorExtensions" | "initialContent" | "onPresence" | "suspendContentChanges"
> & {
  collaborationHandle: RichEditorCollaborationHandle;
  suggesting?: MutableRefObject<SuggestingController>;
  commentAnchors?: CommentAnchor[];
  commentAnchorController?: CommentAnchorController;
};

export function CollaborativeRichEditor({
  collaborationHandle,
  historyHighlights,
  suggesting,
  commentAnchors,
  commentAnchorController,
  onEditor,
  ...props
}: CollaborativeRichEditorProps) {
  const editorRef = useRef<Editor | null>(null);
  const editorExtensions = useMemo(() => {
    const base = [
      ...richEditorCollaborationExtensions,
      createRichEditorCollaborationExtension(
        collaborationHandle,
        richEditorCollaborationExtensions,
      ),
      createRichEditorPresenceExtension(collaborationHandle),
    ];

    if (suggesting) {
      base.push(createSuggestingExtension(suggesting));
    }
    if (commentAnchorController) {
      base.push(createCommentAnchorExtension(collaborationHandle, commentAnchorController));
    }
    if (!historyHighlights?.length) {
      return base;
    }

    return [...base, createRichEditorHistoryExtension(historyHighlights)];
  }, [collaborationHandle, commentAnchorController, historyHighlights, suggesting]);

  useEffect(() => {
    if (!editorRef.current || !commentAnchors) return;
    refreshCommentAnchors(editorRef.current, commentAnchors);
  }, [commentAnchors]);

  const initialContent = useMemo(
    () => richEditorAutomergeContent(
      collaborationHandle,
      richEditorCollaborationExtensions,
    ),
    [collaborationHandle],
  );

  const onPresence = useCallback((anchor: number, head: number) => {
    collaborationHandle.updatePresence?.(anchor, head);
  }, [collaborationHandle]);

  const handleEditor = useCallback((editor: Editor | null) => {
    editorRef.current = editor;
    if (editor && commentAnchors) {
      refreshCommentAnchors(editor, commentAnchors);
    }
    onEditor?.(editor);
  }, [commentAnchors, onEditor]);

  return (
    <RichEditor
      {...props}
      editorExtensions={editorExtensions}
      historyHighlights={undefined}
      initialContent={initialContent}
      onEditor={handleEditor}
      onPresence={onPresence}
      suspendContentChanges
    />
  );
}
