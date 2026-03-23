import { CodeIcon, LinkIcon, TextBIcon, TextItalicIcon, TextStrikethroughIcon, TextUnderlineIcon } from "@phosphor-icons/react";
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
        active={editor.isActive("bold")}
        onClick={() => editor.chain().focus().toggleBold().run()}
      >
        <TextBIcon size={16} weight="bold" />
      </MenuButton>
      <MenuButton
        active={editor.isActive("italic")}
        onClick={() => editor.chain().focus().toggleItalic().run()}
      >
        <TextItalicIcon size={16} weight="bold" />
      </MenuButton>
      <MenuButton
        active={editor.isActive("underline")}
        onClick={() => editor.chain().focus().toggleUnderline().run()}
      >
        <TextUnderlineIcon size={16} weight="bold" />
      </MenuButton>
      <MenuButton
        active={editor.isActive("strike")}
        onClick={() => editor.chain().focus().toggleStrike().run()}
      >
        <TextStrikethroughIcon size={16} weight="bold" />
      </MenuButton>
      <MenuButton
        active={editor.isActive("code")}
        onClick={() => editor.chain().focus().toggleCode().run()}
      >
        <CodeIcon size={16} weight="bold" />
      </MenuButton>
      <MenuButton
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
      >
        <LinkIcon size={16} weight="bold" />
      </MenuButton>
    </BaseBubbleMenu>
  );
}
