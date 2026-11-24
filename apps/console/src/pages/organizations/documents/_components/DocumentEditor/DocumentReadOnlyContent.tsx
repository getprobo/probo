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

import { createPaginationExtension } from "@probo/ui/src/RichEditor/pagination";
import { RichEditor, richEditorExtensions } from "@probo/ui/src/RichEditor/RichEditor";
import type { Editor } from "@tiptap/core";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { documentPageLayout, documentPageSize } from "../../_lib/documentPageLayout";

interface DocumentReadOnlyContentProps {
  // ProseMirror JSON.
  content: string;
  orientation: string;
  // Must be stable: a new callback rebuilds the editor.
  onPages: (pages: number) => void;
}

// A read-only document body laid out on pages like the live editor. New
// content replaces the shown one in place, so a refetch neither remounts the
// editor nor moves the scroll.
export function DocumentReadOnlyContent({ content, orientation, onPages }: DocumentReadOnlyContentProps) {
  const [initialContent] = useState(content);
  const [initialDocument] = useState(() => parseContent(content));
  const contentRef = useRef(initialDocument);
  // TipTap reports its editor asynchronously, possibly before this component
  // mounts, so the editor is held in a ref rather than in state.
  const editorRef = useRef<Editor | null>(null);
  const extensions = useMemo(
    () => [
      ...richEditorExtensions,
      createPaginationExtension({
        layout: documentPageLayout(documentPageSize(orientation)),
        onPages,
      }),
    ],
    [onPages, orientation],
  );

  // A rebuilt editor starts from the initial content, so it catches up too.
  const onEditor = useCallback((next: Editor | null) => {
    editorRef.current = next;
    if (next != null && contentRef.current != null) {
      showContent(next, contentRef.current);
    }
  }, []);

  // Malformed content keeps the last good one on screen.
  useEffect(() => {
    const parsed = parseContent(content);
    if (parsed == null) {
      return;
    }
    contentRef.current = parsed;
    if (editorRef.current != null) {
      showContent(editorRef.current, parsed);
    }
  }, [content]);

  return (
    <RichEditor
      className="w-full"
      content={initialContent}
      page={false}
      disabled
      editorExtensions={extensions}
      onEditor={onEditor}
    />
  );
}

function parseContent(content: string): object | null {
  try {
    const parsed: unknown = JSON.parse(content);
    return typeof parsed === "object" && parsed !== null ? parsed : null;
  } catch {
    return null;
  }
}

function showContent(editor: Editor, content: object) {
  if (editor.isDestroyed || JSON.stringify(editor.getJSON()) === JSON.stringify(content)) {
    return;
  }
  editor.commands.setContent(content, { emitUpdate: false });
}
