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
import { type ComponentProps, useCallback, useContext, useEffect, useLayoutEffect, useRef } from "react";
import { tv } from "tailwind-variants";

import useToast from "../v2/Toaster/useToast";

import { fileAccept, imageAccept } from "./attachments";
import { BlockMenu } from "./BlockMenu/BlockMenu";
import { BubbleMenu } from "./BubbleMenu";
import { CodeBlockExtension } from "./CodeBlockExtension";
import { ContentUploadExtension } from "./ContentUploadExtension";
import { FileExtension } from "./FileExtension";
import { ImageExtension } from "./ImageExtension";
import { LinkExtension } from "./LinkExtension";
import { MarkdownPasteExtension } from "./MarkdownPasteExtension";
import { OptionsMenu } from "./OptionsMenu/OptionsMenu";
import { PlaceholderExtension, setPlaceholder } from "./PlaceholderExtension";
import { RichTextUploadContext } from "./RichTextUploadContext";
import { SlashCommandExtension } from "./SlashCommandExtension";
import { TableCellMenu } from "./TableCellMenu/TableCellMenu";
import { TableColumnMenu } from "./TableColumnMenu/TableColumnMenu";
import { TableRowMenu } from "./TableRowMenu/TableRowMenu";
import { TableSelectionOverlay } from "./TableSelectionOverlay";
import {
  type UploadAttachment,
  uploadAttachments,
} from "./uploadAttachments";

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
  ImageExtension,
  FileExtension,
  ContentUploadExtension,
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

function stripNonTextMarks(node: JSONContent) {
  if (node.type !== "text") delete node.marks;
  node.content?.forEach(stripNonTextMarks);
}

function withoutContentUploads(node: JSONContent): JSONContent {
  if (!node.content) {
    return node;
  }

  const content = node.content
    .filter(child => child.type !== "contentUpload")
    .map(withoutContentUploads);
  const needsBlock = node.type === "doc"
    || node.type === "blockquote"
    || node.type === "listItem"
    || node.type === "tableCell"
    || node.type === "tableHeader";

  return {
    ...node,
    content: needsBlock && content.length === 0 ? [{ type: "paragraph" }] : content,
  };
}

type RichEditorProps = ComponentProps<"div"> & {
  content: string;
  disabled?: boolean;
  placeholder?: string;
  onChangeContent?: (content: string) => void;
  onUploadFile?: UploadAttachment;
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
    onChangeContent,
    onUploadFile,
    ...divProps
  } = props;
  const contextUpload = useContext(RichTextUploadContext);
  const uploadFile = onUploadFile ?? contextUpload ?? undefined;
  const uploadRef = useRef<UploadAttachment | null>(null);
  const emittedRef = useRef<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const toast = useToast();
  const reportUploadErrorRef = useRef<(message: string) => void>(() => {});
  const canUpload = Boolean(uploadFile && !disabled);

  useLayoutEffect(() => {
    if (uploadFile == null || disabled) {
      uploadRef.current = null;
    } else {
      uploadRef.current = uploadFile;
    }

    reportUploadErrorRef.current = (message) => {
      toast.add({
        title: "Upload failed",
        description: message,
        type: "error",
      });
    };
  }, [disabled, toast, uploadFile]);

  const reportUploadError = useCallback((message: string) => {
    reportUploadErrorRef.current(message);
  }, []);

  const handleUpdate = useCallback(
    ({ editor }: { editor: Editor }) => {
      if (editor.isDestroyed) {
        return;
      }

      const json = withoutContentUploads(editor.getJSON());
      stripNonTextMarks(json);

      const next = JSON.stringify(json);
      if (next === emittedRef.current) {
        return;
      }

      emittedRef.current = next;
      onChangeContent?.(next);
    },
    [onChangeContent],
  );

  const editor = useEditor({
    editorProps: {
      attributes: {
        class: "h-full",
      },
      handlePaste(view, event) {
        const upload = uploadRef.current;
        if (!upload || !view.editable) {
          return false;
        }

        const files = Array.from(event.clipboardData?.files ?? []);
        if (files.length === 0) {
          return false;
        }

        event.preventDefault();
        void uploadAttachments(view, files, view.state.selection.from, upload, reportUploadError);

        return true;
      },
      handleDrop(view, event, _slice, moved) {
        if (moved) {
          return false;
        }

        const upload = uploadRef.current;
        if (!upload || !view.editable) {
          return false;
        }

        const files = Array.from(event.dataTransfer?.files ?? []);
        if (files.length === 0) {
          return false;
        }

        event.preventDefault();
        const coords = view.posAtCoords({ left: event.clientX, top: event.clientY });
        void uploadAttachments(
          view,
          files,
          coords?.pos ?? view.state.selection.from,
          upload,
          reportUploadError,
        );

        return true;
      },
    },
    editable: !disabled,
    extensions,
    content: parseContent(content),
    onCreate: ({ editor: created }) => {
      const json = withoutContentUploads(created.getJSON());
      stripNonTextMarks(json);
      emittedRef.current = JSON.stringify(json);
    },
    onUpdate: handleUpdate,
  });

  const openFilePicker = useCallback((kind: "image" | "file") => {
    const input = fileInputRef.current;
    const upload = uploadRef.current;
    if (!input || !upload || !editor) {
      return;
    }

    input.accept = kind === "image" ? imageAccept : fileAccept;
    input.value = "";
    input.click();
  }, [editor]);

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

  if (!editor) return null;

  return (
    <div className={richEditorVariants({ className, disabled })} {...divProps}>
      {!disabled
        && (
          <>
            <BubbleMenu editor={editor} />
            <BlockMenu
              editor={editor}
              onPickFiles={canUpload ? openFilePicker : undefined}
            />
            <OptionsMenu editor={editor} />
            <TableSelectionOverlay editor={editor} />
            <TableCellMenu editor={editor} />
            <TableColumnMenu editor={editor} />
            <TableRowMenu editor={editor} />
          </>
        )}

      <EditorContent className="h-full" editor={editor} />
      <input
        ref={fileInputRef}
        type="file"
        multiple
        className="hidden"
        onChange={(event) => {
          const upload = uploadRef.current;
          const files = Array.from(event.target.files ?? []);
          event.target.value = "";
          if (!upload || files.length === 0) {
            return;
          }

          void uploadAttachments(
            editor.view,
            files,
            editor.state.selection.from,
            upload,
            reportUploadError,
          );
        }}
      />
    </div>
  );
}
