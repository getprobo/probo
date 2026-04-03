// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { Blockquote } from "@tiptap/extension-blockquote";
import { Bold } from "@tiptap/extension-bold";
import { BulletList } from "@tiptap/extension-bullet-list";
import { Code } from "@tiptap/extension-code";
import { Document } from "@tiptap/extension-document";
import { HardBreak } from "@tiptap/extension-hard-break";
import { Heading } from "@tiptap/extension-heading";
import { HorizontalRule } from "@tiptap/extension-horizontal-rule";
import { Italic } from "@tiptap/extension-italic";
import { ListItem, OrderedList } from "@tiptap/extension-list";
import { Paragraph } from "@tiptap/extension-paragraph";
import { Strike } from "@tiptap/extension-strike";
import { TableKit } from "@tiptap/extension-table";
import { Text } from "@tiptap/extension-text";
import { Underline } from "@tiptap/extension-underline";
import { Dropcursor, UndoRedo } from "@tiptap/extensions";
import { type Content, Editor, EditorContent, useEditor } from "@tiptap/react";
import { type ComponentProps, useCallback, useEffect } from "react";
import { tv } from "tailwind-variants";

import { BlockMenu } from "./BlockMenu/BlockMenu";
import { BubbleMenu } from "./BubbleMenu";
import { CodeBlockExtension } from "./CodeBlockExtension";
import { LinkExtension } from "./LinkExtension";
import { MarkdownPasteExtension } from "./MarkdownPasteExtension";
import { OptionsMenu } from "./OptionsMenu/OptionsMenu";
import { PlaceholderExtension } from "./PlaceholderExtension";
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
  PlaceholderExtension,
  Blockquote,
  BulletList,
  OrderedList,
  ListItem,
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
];

const richEditorVariants = tv({
  base: ["relative flex-1 py-14 pr-8 bg-level-1 shadow-base"],
  variants: {
    disabled: {
      true: "pl-8",
      false: "pl-14",
    },
  },
});

type RichEditorProps = ComponentProps<"div"> & {
  content: string;
  disabled?: boolean;
  onChangeContent: (content: string) => void;
};

export function RichEditor(props: RichEditorProps) {
  const { className, content, disabled = false, onChangeContent, ...divProps } = props;

  const handleUpdate = useCallback(
    ({ editor }: { editor: Editor }) => {
      onChangeContent(JSON.stringify(editor.getJSON()));
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
    extensions,
    content: (content ? JSON.parse(content) : "") as Content,
    onUpdate: handleUpdate,
  });

  useEffect(() => {
    if (!editor) return;
    editor.setEditable(!disabled, false);
  }, [editor, disabled]);

  if (!editor) return null;

  return (
    <div className={richEditorVariants({ className, disabled })} {...divProps}>
      {!disabled
        && (
          <>
            <BubbleMenu editor={editor} />
            <BlockMenu editor={editor} />
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
