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
import { Text } from "@tiptap/extension-text";
import { Underline } from "@tiptap/extension-underline";
import { Dropcursor, Gapcursor, UndoRedo } from "@tiptap/extensions";
import { type Content, EditorContent, useEditor, useEditorState } from "@tiptap/react";
import { BubbleMenu } from "@tiptap/react/menus";
import { type ComponentProps, useEffect, useRef } from "react";
import { tv } from "tailwind-variants";

import { BlockMenu } from "./BlockMenu";
import { LinkExtension } from "./LinkExtension";
import { MenuButton } from "./MenuButton";
import { OptionsMenu } from "./OptionsMenu";

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
  Blockquote,
  BulletList,
  OrderedList,
  ListItem,
  HorizontalRule,
  HardBreak,
  Dropcursor,
  Gapcursor,
  UndoRedo,
];

const richEditorVariants = tv({
  slots: {
    bubbleMenu: ["flex items-center gap-1 rounded-lg border border-border-mid bg-level-0 p-1 shadow-md"],
    editor: ["h-full px-12"],
  },
});

const { bubbleMenu, editor: editorVariants } = richEditorVariants();

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
    <div className={editorVariants({ className })}>
      <BubbleMenu
        editor={editor}
        className={bubbleMenu()}
      >
        <MenuButton
          label="Bold"
          active={editor.isActive("bold")}
          onClick={() => editor.chain().focus().toggleBold().run()}
        />
        <MenuButton
          label="Italic"
          active={editor.isActive("italic")}
          onClick={() => editor.chain().focus().toggleItalic().run()}
        />
        <MenuButton
          label="Underline"
          active={editor.isActive("underline")}
          onClick={() => editor.chain().focus().toggleUnderline().run()}
        />
        <MenuButton
          label="Strike"
          active={editor.isActive("strike")}
          onClick={() => editor.chain().focus().toggleStrike().run()}
        />
        <MenuButton
          label="Code"
          active={editor.isActive("code")}
          onClick={() => editor.chain().focus().toggleCode().run()}
        />
        <MenuButton
          label="Link"
          active={editor.isActive("link")}
          onClick={() => {
            if (editor.isActive("link")) {
              editor.chain().focus().unsetLink().run();
              return;
            }
            const url = window.prompt("URL");
            if (url) {
              editor.chain().focus().setLink({ href: url }).run();
            }
          }}
        />
      </BubbleMenu>

      <BlockMenu editor={editor} />
      <OptionsMenu editor={editor} />

      <EditorContent className="h-full" editor={editor} />
    </div>
  );
}
