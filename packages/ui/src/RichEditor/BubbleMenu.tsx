import type { useEditor } from "@tiptap/react";
import { BubbleMenu as BaseBubbleMenu } from "@tiptap/react/menus";
import { tv } from "tailwind-variants";

import { MenuButton } from "./MenuButton";

const bubbleMenuVariants = tv({
  base: ["flex items-center gap-1 rounded-lg border border-border-mid bg-level-0 p-1 shadow-md"],
});

type BubbleMenuProps = {
  editor: ReturnType<typeof useEditor>;
};

export function BubbleMenu(props: BubbleMenuProps) {
  const { editor } = props;

  return (
    <BaseBubbleMenu
      editor={editor}
      className={bubbleMenuVariants()}
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
    </BaseBubbleMenu>
  );
}
