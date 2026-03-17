import { Blockquote } from "@tiptap/extension-blockquote";
import { Bold } from "@tiptap/extension-bold";
import { Code } from "@tiptap/extension-code";
import { CodeBlock } from "@tiptap/extension-code-block";
import { Document } from "@tiptap/extension-document";
import { HardBreak } from "@tiptap/extension-hard-break";
import { Heading } from "@tiptap/extension-heading";
import { HorizontalRule } from "@tiptap/extension-horizontal-rule";
import { Italic } from "@tiptap/extension-italic";
import { Link } from "@tiptap/extension-link";
import { BulletList, ListItem, OrderedList } from "@tiptap/extension-list";
import { Paragraph } from "@tiptap/extension-paragraph";
import { Strike } from "@tiptap/extension-strike";
import { Text } from "@tiptap/extension-text";
import { Underline } from "@tiptap/extension-underline";
import { Dropcursor, Gapcursor, UndoRedo } from "@tiptap/extensions";
import { type Content, EditorContent, useEditor, useEditorState } from "@tiptap/react";
import { BubbleMenu, FloatingMenu } from "@tiptap/react/menus";
import { useEffect } from "react";
import { tv } from "tailwind-variants";

const extensions = [
  Document,
  Paragraph,
  Text,
  Heading.configure({ levels: [1, 2, 3] }),
  Bold,
  Italic,
  Strike,
  Underline,
  Code,
  CodeBlock,
  Link.configure({ openOnClick: false }),
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
    floatingMenu: ["flex items-center gap-1 rounded-lg border border-border-mid bg-level-0 p-1 shadow-md"],
    menuButton: ["px-2 py-1 text-sm rounded-sm font-semibold bg-level-0 hover:bg-subtle"],
  },
  variants: {
    active: {
      true: {
        menuButton: ["bg-active"],
      },
    },
  },
});

const { bubbleMenu, floatingMenu, menuButton } = richEditorVariants();

type MenuButtonProps = {
  label: string;
  active?: boolean;
  onClick: () => void;
};

function MenuButton({ label, active, onClick }: MenuButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={menuButton({ active })}
    >
      {label}
    </button>
  );
}

interface RichEditorProps {
  content: string;
  onChange: (content: string) => void;
}

export function RichEditor(props: RichEditorProps) {
  const { content, onChange } = props;

  const editor = useEditor({
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
    if (watchedContent !== content) {
      onChange(watchedContent);
    }
  }, [content, watchedContent, onChange]);

  return (
    <div>
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

      <FloatingMenu
        editor={editor}
        className={floatingMenu()}
      >
        <MenuButton
          label="H1"
          active={editor.isActive("heading", { level: 1 })}
          onClick={() =>
            editor.chain().focus().toggleHeading({ level: 1 }).run()}
        />
        <MenuButton
          label="H2"
          active={editor.isActive("heading", { level: 2 })}
          onClick={() =>
            editor.chain().focus().toggleHeading({ level: 2 }).run()}
        />
        <MenuButton
          label="H3"
          active={editor.isActive("heading", { level: 3 })}
          onClick={() =>
            editor.chain().focus().toggleHeading({ level: 3 }).run()}
        />
        <MenuButton
          label="Bullet List"
          active={editor.isActive("bulletList")}
          onClick={() => editor.chain().focus().toggleBulletList().run()}
        />
        <MenuButton
          label="Ordered List"
          active={editor.isActive("orderedList")}
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
        />
        <MenuButton
          label="Code Block"
          active={editor.isActive("codeBlock")}
          onClick={() => editor.chain().focus().toggleCodeBlock().run()}
        />
        <MenuButton
          label="Blockquote"
          active={editor.isActive("blockquote")}
          onClick={() => editor.chain().focus().toggleBlockquote().run()}
        />
        <MenuButton
          label="Divider"
          onClick={() => editor.chain().focus().setHorizontalRule().run()}
        />
      </FloatingMenu>

      <EditorContent editor={editor} />
    </div>
  );
}
