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
import { type ChangeEvent, type ComponentProps, useCallback, useEffect, useLayoutEffect, useMemo, useRef } from "react";
import { tv } from "tailwind-variants";

import { BlockMenu } from "./BlockMenu/BlockMenu";
import { BubbleMenu } from "./BubbleMenu";
import { CodeBlockExtension } from "./CodeBlockExtension";
import { ImageExtension, insertPicture } from "./ImageExtension";
import { LinkExtension } from "./LinkExtension";
import { MarkdownPasteExtension } from "./MarkdownPasteExtension";
import { OptionsMenu } from "./OptionsMenu/OptionsMenu";
import { PictureUploadExtension, type RichEditorPictureUpload, setPictureUpload } from "./PictureUploadExtension";
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

const pictureInput = tv({
  base: "sr-only",
});

export type RichEditorPictures = {
  upload?: RichEditorPictureUpload;
};

function stripNonTextMarks(node: JSONContent) {
  if (node.type !== "text") delete node.marks;
  node.content?.forEach(stripNonTextMarks);
}

type RichEditorProps = ComponentProps<"div"> & {
  content: string;
  disabled?: boolean;
  placeholder?: string;
  pictures?: RichEditorPictures;
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
    pictures,
    onChangeContent,
    ...divProps
  } = props;

  const uploadRef = useRef<RichEditorPictureUpload | undefined>(undefined);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const insertAtRef = useRef<number | null>(null);
  const picturesEnabled = pictures != null;
  const uploadEnabled = pictures?.upload != null && !disabled;

  const editorExtensions = useMemo(() => {
    if (!picturesEnabled) {
      return extensions;
    }

    const withPictures = [...extensions, ImageExtension];
    if (!uploadEnabled) {
      return withPictures;
    }

    return [...withPictures, PictureUploadExtension];
  }, [picturesEnabled, uploadEnabled]);

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
    const upload = disabled ? undefined : pictures?.upload;
    uploadRef.current = upload;
    if (!editor || editor.isDestroyed) {
      return;
    }

    setPictureUpload(editor, upload);
  }, [disabled, editor, pictures?.upload]);

  const openPicturePicker = useCallback(() => {
    if (!editor || editor.isDestroyed) {
      return;
    }

    insertAtRef.current = editor.state.selection.from;
    fileInputRef.current?.click();
  }, [editor]);

  const handlePictureFile = useCallback((event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    event.target.value = "";
    const upload = uploadRef.current;
    if (!file || !upload || !editor || editor.isDestroyed) {
      return;
    }

    const pos = insertAtRef.current ?? editor.state.selection.from;
    void upload(file).then(
      (picture) => {
        if (editor.isDestroyed) {
          return;
        }

        insertPicture(editor.view, pos, {
          src: picture.src,
          alt: picture.alt ?? file.name,
        });
      },
      () => {
        // The upload callback reports the failure.
      },
    );
  }, [editor]);

  if (!editor) return null;

  return (
    <div className={richEditorVariants({ className, disabled })} {...divProps}>
      {uploadEnabled && (
        <input
          ref={fileInputRef}
          className={pictureInput()}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          tabIndex={-1}
          aria-label="Upload picture"
          onChange={handlePictureFile}
        />
      )}
      {!disabled
        && (
          <>
            <BubbleMenu editor={editor} />
            <BlockMenu
              editor={editor}
              onInsertPicture={uploadEnabled ? openPicturePicker : undefined}
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
