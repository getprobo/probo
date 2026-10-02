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

import { Blockquote } from "@tiptap/extension-blockquote";
import { Bold } from "@tiptap/extension-bold";
import { Code } from "@tiptap/extension-code";
import { Document } from "@tiptap/extension-document";
import { HardBreak } from "@tiptap/extension-hard-break";
import { Heading } from "@tiptap/extension-heading";
import { HorizontalRule } from "@tiptap/extension-horizontal-rule";
import { Italic } from "@tiptap/extension-italic";
import { BulletList, ListItem, ListKeymap, OrderedList } from "@tiptap/extension-list";
import { Paragraph } from "@tiptap/extension-paragraph";
import { Strike } from "@tiptap/extension-strike";
import { TableKit } from "@tiptap/extension-table";
import { Text } from "@tiptap/extension-text";
import { Underline } from "@tiptap/extension-underline";
import { Dropcursor, UndoRedo } from "@tiptap/extensions";
import { type Content, Editor, EditorContent, type JSONContent, useEditor } from "@tiptap/react";
import { type ChangeEvent, type ComponentProps, type DragEvent, useCallback, useEffect, useLayoutEffect, useMemo, useRef } from "react";
import { tv } from "tailwind-variants";

import { attachmentAccept, AttachmentExtension } from "./AttachmentExtension";
import { AttachmentUploadExtension, filesFromDataTransfer, insertUploadedFile, type RichEditorAttachmentUpload, setAttachmentUpload, uploadEditorFiles } from "./AttachmentUploadExtension";
import { BlockMenu } from "./BlockMenu/BlockMenu";
import { BubbleMenu } from "./BubbleMenu";
import { CodeBlockExtension } from "./CodeBlockExtension";
import { ImageExtension } from "./ImageExtension";
import { LinkExtension } from "./LinkExtension";
import { MarkdownPasteExtension } from "./MarkdownPasteExtension";
import { OptionsMenu } from "./OptionsMenu/OptionsMenu";
import { PlaceholderExtension, setPlaceholder } from "./PlaceholderExtension";
import { SlashCommandExtension } from "./SlashCommandExtension";
import { TableCellMenu } from "./TableCellMenu/TableCellMenu";
import { TableColumnMenu } from "./TableColumnMenu/TableColumnMenu";
import { TableRowMenu } from "./TableRowMenu/TableRowMenu";
import { TableSelectionOverlay } from "./TableSelectionOverlay";

const extensions = [
  Document,
  Paragraph,
  Text,
  Heading,
  Bold,
  Italic,
  Strike,
  Underline,
  Code,
  CodeBlockExtension,
  LinkExtension,
  SlashCommandExtension,
  Blockquote,
  BulletList,
  OrderedList,
  ListItem,
  ListKeymap,
  HorizontalRule,
  HardBreak,
  Dropcursor.configure({
    color: "#0081f1",
    width: 2,
  }),
  UndoRedo,
  TableKit.configure({
    table: { resizable: true },
  }),
  MarkdownPasteExtension,
  PlaceholderExtension,
];

const richEditorVariants = tv({
  base: ["relative flex-1 min-w-0 overflow-auto py-14 pr-8 bg-level-1 shadow-base"],
  variants: {
    disabled: {
      true: "pl-8",
      false: "pl-14",
    },
  },
});

const attachmentInput = tv({
  base: "sr-only",
});

export type RichEditorAttachments = {
  upload?: RichEditorAttachmentUpload;
};

function stripNonTextMarks(node: JSONContent) {
  if (node.type !== "text") delete node.marks;
  node.content?.forEach(stripNonTextMarks);
}

type RichEditorProps = ComponentProps<"div"> & {
  content: string;
  disabled?: boolean;
  placeholder?: string;
  attachments?: RichEditorAttachments;
  onChangeContent?: (content: string) => void;
};

function parseContent(content: string): Content {
  if (!content) {
    return "";
  }

  try {
    return JSON.parse(content) as Content;
  } catch {
    return "";
  }
}

export function RichEditor(props: RichEditorProps) {
  const {
    className,
    content,
    disabled = false,
    placeholder,
    attachments,
    onChangeContent,
    ...divProps
  } = props;

  const uploadRef = useRef<RichEditorAttachmentUpload | undefined>(undefined);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const insertAtRef = useRef<number | null>(null);
  const attachmentsEnabled = attachments != null;
  const uploadEnabled = attachments?.upload != null && !disabled;

  const editorExtensions = useMemo(() => {
    if (!attachmentsEnabled) {
      return extensions;
    }

    const withAttachments = [...extensions, ImageExtension, AttachmentExtension];
    if (!uploadEnabled) {
      return withAttachments;
    }

    return [...withAttachments, AttachmentUploadExtension];
  }, [attachmentsEnabled, uploadEnabled]);

  const handleUpdate = useCallback(
    ({ editor }: { editor: Editor }) => {
      if (editor.isDestroyed) {
        return;
      }

      const json = editor.getJSON();
      stripNonTextMarks(json);

      onChangeContent?.(JSON.stringify(json));
    },
    [onChangeContent],
  );

  const editor = useEditor({
    editorProps: {
      attributes: {
        class: "h-full",
      },
    },
    editable: !disabled,
    extensions: editorExtensions,
    content: parseContent(content),
    onUpdate: handleUpdate,
  });

  useLayoutEffect(() => {
    if (!editor) {
      return;
    }

    setPlaceholder(editor, placeholder);
  }, [editor, placeholder]);

  useEffect(() => {
    if (!editor) {
      return;
    }

    editor.setEditable(!disabled, false);
  }, [editor, disabled]);

  useEffect(() => {
    const upload = disabled ? undefined : attachments?.upload;
    uploadRef.current = upload;
    if (!editor || editor.isDestroyed) {
      return;
    }

    setAttachmentUpload(editor, upload);
  }, [disabled, editor, attachments?.upload]);

  const openAttachmentPicker = useCallback(() => {
    if (!editor || editor.isDestroyed) {
      return;
    }

    insertAtRef.current = editor.state.selection.from;
    fileInputRef.current?.click();
  }, [editor]);

  const handleAttachmentFile = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    const upload = uploadRef.current;
    if (!file || !upload || !editor || editor.isDestroyed) {
      return;
    }

    const pos = insertAtRef.current ?? editor.state.selection.from;
    void upload(file).then(
      (uploaded) => {
        if (editor.isDestroyed) {
          return;
        }

        insertUploadedFile(editor.view, pos, uploaded);
      },
      () => {
        // The upload callback reports the failure.
      },
    );
  }, [editor]);

  const handleAttachmentDragOver = useCallback((event: DragEvent<HTMLDivElement>) => {
    if (!uploadRef.current || !event.dataTransfer) {
      return;
    }

    if (![...event.dataTransfer.types].includes("Files")) {
      return;
    }

    event.preventDefault();
    event.dataTransfer.dropEffect = "copy";
  }, []);

  const handleAttachmentDrop = useCallback((event: DragEvent<HTMLDivElement>) => {
    if (event.defaultPrevented) {
      return;
    }

    const upload = uploadRef.current;
    if (!upload || !editor || editor.isDestroyed) {
      return;
    }

    const files = filesFromDataTransfer(event.dataTransfer);
    const fileDrag = files.length > 0
      || (event.dataTransfer?.types.includes("Files") ?? false);
    if (!fileDrag) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();
    const coords = editor.view.posAtCoords({
      left: event.clientX,
      top: event.clientY,
    });
    uploadEditorFiles(
      editor.view,
      files,
      coords?.pos ?? editor.state.selection.from,
      upload,
    );
  }, [editor]);

  if (!editor) return null;

  return (
    <div
      className={richEditorVariants({ className, disabled })}
      {...divProps}
      onDragOver={handleAttachmentDragOver}
      onDrop={handleAttachmentDrop}
    >
      {uploadEnabled && (
        <input
          ref={fileInputRef}
          className={attachmentInput()}
          type="file"
          accept={attachmentAccept}
          tabIndex={-1}
          aria-label="Upload attachment"
          onChange={handleAttachmentFile}
        />
      )}
      {!disabled
        && (
          <>
            <BubbleMenu editor={editor} />
            <BlockMenu
              editor={editor}
              onInsertAttachment={uploadEnabled ? openAttachmentPicker : undefined}
            />
            <OptionsMenu editor={editor} />
            <TableSelectionOverlay editor={editor} />
            <TableCellMenu editor={editor} />
            <TableColumnMenu editor={editor} />
            <TableRowMenu editor={editor} />
          </>
        )}

      <EditorContent className="h-full" editor={editor} />
    </div>
  );
}
