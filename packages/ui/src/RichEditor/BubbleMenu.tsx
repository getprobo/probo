import { CodeIcon, LinkIcon, TextBIcon, TextItalicIcon, TextStrikethroughIcon, TextUnderlineIcon, TrashIcon } from "@phosphor-icons/react";
import { Editor, useEditorState } from "@tiptap/react";
import { BubbleMenu as BaseBubbleMenu } from "@tiptap/react/menus";
import { type KeyboardEvent, useEffect, useRef, useState } from "react";
import { tv } from "tailwind-variants";

import { MenuButton } from "./MenuButton";

const bubbleMenuVariants = tv({
  slots: {
    root: "flex flex-col items-stretch rounded-lg border border-border-mid bg-level-0 shadow-md",
    toolbar: "flex items-center gap-1 p-1",
    linkInput: "flex items-center gap-1 border-t border-border-mid px-2 py-1.5",
  },
});

const { root, toolbar, linkInput } = bubbleMenuVariants();

type BubbleMenuProps = {
  editor: Editor;
};

export function BubbleMenu(props: BubbleMenuProps) {
  const { editor } = props;
  const [showLinkInput, setShowLinkInput] = useState(false);
  const [linkUrl, setLinkUrl] = useState("");
  const linkInputRef = useRef<HTMLInputElement>(null);

  const state = useEditorState({
    editor,
    selector: ({ editor }) => ({
      isBold: editor.isActive("bold"),
      isItalic: editor.isActive("italic"),
      isUnderline: editor.isActive("underline"),
      isStrike: editor.isActive("strike"),
      isCode: editor.isActive("code"),
      isLink: editor.isActive("link"),
      linkHref: editor.getAttributes("link").href as string | undefined,
    }),
  });

  useEffect(() => {
    if (showLinkInput) {
      linkInputRef.current?.focus();
    }
  }, [showLinkInput]);

  function handleLinkButtonClick() {
    if (showLinkInput) {
      setShowLinkInput(false);
      setLinkUrl("");
      return;
    }
    setLinkUrl(state.linkHref ?? "");
    setShowLinkInput(true);
  }

  function submitLink() {
    if (linkUrl.trim()) {
      editor.chain().focus().setLink({ href: linkUrl.trim() }).run();
    }
    setShowLinkInput(false);
    setLinkUrl("");
  }

  function removeLink() {
    editor.chain().focus().unsetLink().run();
    setShowLinkInput(false);
    setLinkUrl("");
  }

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter") {
      e.preventDefault();
      submitLink();
    }
    if (e.key === "Escape") {
      e.preventDefault();
      setShowLinkInput(false);
      setLinkUrl("");
      editor.commands.focus();
    }
  }

  return (
    <BaseBubbleMenu
      editor={editor}
      className={root()}
    >
      <div className={toolbar()}>
        <MenuButton
          active={state.isBold}
          onClick={() => editor.chain().focus().toggleBold().run()}
        >
          <TextBIcon size={16} weight="bold" />
        </MenuButton>
        <MenuButton
          active={state.isItalic}
          onClick={() => editor.chain().focus().toggleItalic().run()}
        >
          <TextItalicIcon size={16} weight="bold" />
        </MenuButton>
        <MenuButton
          active={state.isUnderline}
          onClick={() => editor.chain().focus().toggleUnderline().run()}
        >
          <TextUnderlineIcon size={16} weight="bold" />
        </MenuButton>
        <MenuButton
          active={state.isStrike}
          onClick={() => editor.chain().focus().toggleStrike().run()}
        >
          <TextStrikethroughIcon size={16} weight="bold" />
        </MenuButton>
        <MenuButton
          active={state.isCode}
          onClick={() => editor.chain().focus().toggleCode().run()}
        >
          <CodeIcon size={16} weight="bold" />
        </MenuButton>
        <MenuButton
          active={state.isLink || showLinkInput}
          onClick={handleLinkButtonClick}
        >
          <LinkIcon size={16} weight="bold" />
        </MenuButton>
      </div>
      {showLinkInput && (
        <div className={linkInput()}>
          <input
            ref={linkInputRef}
            type="url"
            placeholder="https://…"
            value={linkUrl}
            onChange={e => setLinkUrl(e.target.value)}
            onKeyDown={handleKeyDown}
            onBlur={submitLink}
            className="min-w-0 flex-1 bg-transparent text-txt-primary outline-none placeholder:text-txt-quaternary"
          />
          {state.isLink && (
            <button
              type="button"
              onMouseDown={e => e.preventDefault()}
              onClick={removeLink}
              className="flex shrink-0 cursor-pointer items-center rounded-sm p-1 text-txt-secondary hover:bg-subtle"
            >
              <TrashIcon size={16} weight="bold" />
            </button>
          )}
        </div>
      )}
    </BaseBubbleMenu>
  );
}
