// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { Blockquote } from "@tiptap/extension-blockquote";
import { Bold } from "@tiptap/extension-bold";
import { Code } from "@tiptap/extension-code";
import { CodeBlock } from "@tiptap/extension-code-block";
import { Document } from "@tiptap/extension-document";
import { HardBreak } from "@tiptap/extension-hard-break";
import { Heading } from "@tiptap/extension-heading";
import { HorizontalRule } from "@tiptap/extension-horizontal-rule";
import { Italic } from "@tiptap/extension-italic";
import { BulletList, ListItem, OrderedList } from "@tiptap/extension-list";
import { Paragraph } from "@tiptap/extension-paragraph";
import { Strike } from "@tiptap/extension-strike";
import { TableKit } from "@tiptap/extension-table";
import { Text } from "@tiptap/extension-text";
import { Underline } from "@tiptap/extension-underline";
import { Dropcursor, Gapcursor, UndoRedo } from "@tiptap/extensions";
import { type Content, EditorContent, useEditor, useEditorState } from "@tiptap/react";
import { type ComponentProps, useEffect, useRef } from "react";
import { tv } from "tailwind-variants";

import { BlockMenu } from "./BlockMenu";
import { BubbleMenu } from "./BubbleMenu";
import { FocusedCellExtension } from "./FocusedCellExtension";
import { LinkExtension } from "./LinkExtension";
import { OptionsMenu } from "./OptionsMenu";
import { PlaceholderExtension } from "./PlaceholderExtension";
import { SlashCommandExtension } from "./SlashCommandExtension";
import { TableCellMenu } from "./TableCellMenu";
import { TableColumnMenu } from "./TableColumnMenu";
import { TableRowMenu } from "./TableRowMenu";

const extensions = [
  Document,
  Paragraph,
  Text,
  Heading.configure({
    levels: [1, 2, 3],
  }),
  Bold,
  Italic,
  Strike,
  Underline,
  Code,
  CodeBlock,
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
  Gapcursor,
  UndoRedo,
  TableKit.configure({
    table: { resizable: true },
  }),
  FocusedCellExtension,
];

const richEditorVariants = tv({
  base: ["h-full pl-14"],
});

type RichEditorProps = ComponentProps<"div"> & {
  content: string;
  disabled?: boolean;
  onChangeContent: (content: string) => void;
};

export function RichEditor(props: RichEditorProps) {
  const { className, content, disabled = false, onChangeContent } = props;

  const previousContentRef = useRef<string>(content);

  const editor = useEditor({
    editorProps: {
      attributes: {
        class: "h-full",
      },
    },
    editable: !disabled,
    extensions,
    content: (content ? JSON.parse(content) : "") as Content,
  });

  const watchedContent = useEditorState({
    editor,
    selector: ({ editor }) => {
      return JSON.stringify(editor.getJSON());
    },
  });

  useEffect(() => {
    if (watchedContent !== previousContentRef.current) {
      previousContentRef.current = watchedContent;
      onChangeContent(watchedContent);
    }
  }, [content, watchedContent, onChangeContent]);

  return (
    <div className={richEditorVariants({ className })}>
      <BubbleMenu editor={editor} />
      <BlockMenu editor={editor} />
      <OptionsMenu editor={editor} />
      <TableCellMenu editor={editor} />
      <TableColumnMenu editor={editor} />
      <TableRowMenu editor={editor} />
      <EditorContent className="h-full" editor={editor} />
    </div>
  );
}
