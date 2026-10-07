// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { ArrowSquareOutIcon, CodeIcon, LinkIcon, TextBIcon, TextItalicIcon, TextStrikethroughIcon, TextUnderlineIcon, TrashIcon } from "@phosphor-icons/react";
import { Editor, useEditorState } from "@tiptap/react";
import { BubbleMenu as BaseBubbleMenu } from "@tiptap/react/menus";
import { type FocusEvent, type KeyboardEvent, useContext, useEffect, useRef, useState } from "react";
import { tv } from "tailwind-variants/lite";

import { OverlayPortalRootContext } from "../lib/overlayPortalRoot";

import { linkEditorPluginKey, openLink } from "./LinkExtension";
import { MenuButton } from "./MenuButton";

const bubbleMenuVariants = tv({
  slots: {
    root: "flex flex-col items-stretch z-20 rounded-lg border border-border-mid bg-level-0 shadow-mid",
    toolbar: "flex items-center gap-1 p-1",
    linkInput: "flex items-center gap-1 border-t border-border-mid px-2 py-1.5",
    linkAction: "flex shrink-0 cursor-pointer items-center rounded-sm p-1 text-txt-secondary hover:bg-subtle",
  },
});

const { root, toolbar, linkInput, linkAction } = bubbleMenuVariants();

type BubbleMenuProps = {
  editor: Editor;
};

export function BubbleMenu(props: BubbleMenuProps) {
  const { editor } = props;
  const portalRoot = useContext(OverlayPortalRootContext);
  const [showLinkInput, setShowLinkInput] = useState(false);
  const [linkUrl, setLinkUrl] = useState("");
  const linkInputRef = useRef<HTMLInputElement>(null);

  const state = useEditorState({
    editor,
    selector: ({ editor: current }) => {
      if (current.isDestroyed) {
        return {
          isBold: false,
          isItalic: false,
          isUnderline: false,
          isStrike: false,
          isCode: false,
          isLink: false,
          linkHref: undefined,
          linkEditorRequest: null,
        };
      }

      return {
        isBold: current.isActive("bold"),
        isItalic: current.isActive("italic"),
        isUnderline: current.isActive("underline"),
        isStrike: current.isActive("strike"),
        isCode: current.isActive("code"),
        isLink: current.isActive("link"),
        linkHref: current.getAttributes("link").href as string | undefined,
        linkEditorRequest: linkEditorPluginKey.getState(current.state) ?? null,
      };
    },
  });
  const linkEditorRequestId = state.linkEditorRequest?.id ?? 0;
  const [handledLinkEditorRequestId, setHandledLinkEditorRequestId] = useState(
    linkEditorRequestId,
  );

  if (linkEditorRequestId !== handledLinkEditorRequestId) {
    setHandledLinkEditorRequestId(linkEditorRequestId);
    setLinkUrl(state.linkEditorRequest?.href ?? "");
    setShowLinkInput(true);
  }

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
    const next = linkUrl.trim();
    if (next && next !== (state.linkHref ?? "")) {
      editor.chain().focus().setLink({ href: next }).run();
    } else if (!editor.isDestroyed) {
      editor.commands.focus();
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
      if (!editor.isDestroyed) {
        editor.commands.focus();
      }
    }
  }

  function handleLinkInputBlur(e: FocusEvent<HTMLDivElement>) {
    if (
      e.currentTarget.contains(e.relatedTarget)
      || !document.hasFocus()
    ) {
      return;
    }

    submitLink();
  }

  return (
    <BaseBubbleMenu
      editor={editor}
      {...(portalRoot ? { appendTo: () => portalRoot } : {})}
      className={root()}
      data-rich-editor-floating=""
      onMouseDown={(e) => {
        if (e.target instanceof HTMLInputElement) {
          return;
        }

        e.preventDefault();
      }}
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
        <div className={linkInput()} onBlur={handleLinkInputBlur}>
          <input
            ref={linkInputRef}
            type="url"
            placeholder="https://…"
            value={linkUrl}
            onChange={e => setLinkUrl(e.target.value)}
            onKeyDown={handleKeyDown}
            className="min-w-0 flex-1 bg-transparent text-txt-primary outline-none placeholder:text-txt-quaternary"
          />
          {linkUrl.trim() !== "" && (
            <button
              type="button"
              aria-label="Open link in a new tab"
              onMouseDown={e => e.preventDefault()}
              onClick={() => openLink(linkUrl.trim())}
              className={linkAction()}
            >
              <ArrowSquareOutIcon size={16} weight="bold" />
            </button>
          )}
          {state.isLink && (
            <button
              type="button"
              aria-label="Remove link"
              onMouseDown={e => e.preventDefault()}
              onClick={removeLink}
              className={linkAction()}
            >
              <TrashIcon size={16} weight="bold" />
            </button>
          )}
        </div>
      )}
    </BaseBubbleMenu>
  );
}
