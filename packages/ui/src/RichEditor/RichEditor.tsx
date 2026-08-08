// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

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
import { type Content, Editor, EditorContent, useEditor } from "@tiptap/react";
import { type ComponentProps, useCallback, useEffect, useMemo } from "react";
import { tv } from "tailwind-variants";

import { AutomergeTableStructureExtension } from "./AutomergeTableStructureExtension";
import { AutomergeUnknownBlockExtension } from "./AutomergeUnknownBlockExtension";
import { BlockMenu } from "./BlockMenu/BlockMenu";
import { BubbleMenu } from "./BubbleMenu";
import { CodeBlockExtension } from "./CodeBlockExtension";
import {
  createRichEditorAutomergeDocument as createAutomergeDocument,
  createRichEditorCollaborationExtension,
  richEditorAutomergeContent,
  supportsRichEditorCollaboration as supportsCollaboration,
} from "./collaboration";
import { LinkExtension } from "./LinkExtension";
import { MarkdownPasteExtension } from "./MarkdownPasteExtension";
import { OptionsMenu } from "./OptionsMenu/OptionsMenu";
import { PlaceholderExtension } from "./PlaceholderExtension";
import {
  createRichEditorPresenceExtension,
  type RichEditorCollaborationHandle,
} from "./presence";
import { SlashCommandExtension } from "./SlashCommandExtension";
import { TableCellMenu } from "./TableCellMenu/TableCellMenu";
import { TableColumnMenu } from "./TableColumnMenu/TableColumnMenu";
import { TableRowMenu } from "./TableRowMenu/TableRowMenu";
import { TableSelectionOverlay } from "./TableSelectionOverlay";

const tableExtension = TableKit.configure({
  table: { resizable: true },
});

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
  ListKeymap,
  HorizontalRule,
  HardBreak,
  Dropcursor.configure({
    color: "#0081f1",
    width: 2,
  }),
  UndoRedo,
  tableExtension,
  AutomergeUnknownBlockExtension,
  MarkdownPasteExtension,
];

export const richEditorCollaborationExtensions = [
  ...extensions.filter(extension => extension !== UndoRedo),
  AutomergeTableStructureExtension,
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

type RichEditorProps = ComponentProps<"div"> & {
  content: string;
  disabled?: boolean;
  onChangeContent: (content: string) => void;
  collaborationHandle?: RichEditorCollaborationHandle;
};

export function RichEditor(props: RichEditorProps) {
  const {
    className,
    content,
    disabled = false,
    onChangeContent,
    collaborationHandle,
    ...divProps
  } = props;

  const handleUpdate = useCallback(
    ({ editor }: { editor: Editor }) => {
      if (collaborationHandle) return;
      onChangeContent(JSON.stringify(editor.getJSON()));
    },
    [collaborationHandle, onChangeContent],
  );

  const editorExtensions = useMemo(
    () => collaborationHandle
      ? [
          ...richEditorCollaborationExtensions,
          createRichEditorCollaborationExtension(
            collaborationHandle,
            richEditorCollaborationExtensions,
          ),
          createRichEditorPresenceExtension(collaborationHandle),
        ]
      : extensions,
    [collaborationHandle],
  );
  const initialContent = collaborationHandle
    ? richEditorAutomergeContent(
        collaborationHandle,
        richEditorCollaborationExtensions,
      )
    : (content ? JSON.parse(content) : "") as Content;

  const editor = useEditor({
    editorProps: {
      attributes: {
        class: "h-full",
      },
    },
    editable: !disabled,
    extensions: editorExtensions,
    content: initialContent,
    onUpdate: handleUpdate,
    onSelectionUpdate: ({ editor }) => {
      collaborationHandle?.updatePresence?.(
        editor.state.selection.anchor,
        editor.state.selection.head,
      );
    },
  }, [collaborationHandle]);

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

export function createRichEditorAutomergeDocument(
  content: string,
) {
  return createAutomergeDocument(content, richEditorCollaborationExtensions);
}

export function supportsRichEditorCollaboration(content: string): boolean {
  return supportsCollaboration(content);
}
