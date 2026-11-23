// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { ArrowSquareOutIcon, CodeIcon, LinkIcon, TextBIcon, TextItalicIcon, TextStrikethroughIcon, TextUnderlineIcon, TrashIcon } from "@phosphor-icons/react";
import { Editor, useEditorState } from "@tiptap/react";
import { type FocusEvent, type KeyboardEvent, useEffect, useRef, useState } from "react";
import { tv } from "tailwind-variants/lite";

import { applyLink, toggleStyle } from "./_lib/formatting";
import { BubbleMenuToolbar } from "./BubbleMenu";
import { linkEditorPluginKey, openLink } from "./LinkExtension";
import { MenuButton } from "./MenuButton";

const bubbleMenuFormattingVariants = tv({
  slots: {
    linkInput: "flex items-center gap-1 border-t border-border-mid px-2 py-1.5",
    linkAction: "flex shrink-0 cursor-pointer items-center rounded-sm p-1 text-txt-secondary hover:bg-subtle",
  },
});

const { linkInput, linkAction } = bubbleMenuFormattingVariants();

type BubbleMenuFormattingProps = {
  editor: Editor;
};

export function BubbleMenuFormatting(props: BubbleMenuFormattingProps) {
  const { editor } = props;
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
      applyLink(editor, next);
    } else if (!editor.isDestroyed) {
      editor.commands.focus();
    }
    setShowLinkInput(false);
    setLinkUrl("");
  }

  function removeLink() {
    applyLink(editor, null);
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
    <>
      <BubbleMenuToolbar>
        <MenuButton
          active={state.isBold}
          onClick={() => toggleStyle(editor, "bold")}
        >
          <TextBIcon size={16} weight="bold" />
        </MenuButton>
        <MenuButton
          active={state.isItalic}
          onClick={() => toggleStyle(editor, "italic")}
        >
          <TextItalicIcon size={16} weight="bold" />
        </MenuButton>
        {editor.schema.marks.underline != null && (
          <MenuButton
            active={state.isUnderline}
            onClick={() => toggleStyle(editor, "underline")}
          >
            <TextUnderlineIcon size={16} weight="bold" />
          </MenuButton>
        )}
        <MenuButton
          active={state.isStrike}
          onClick={() => toggleStyle(editor, "strike")}
        >
          <TextStrikethroughIcon size={16} weight="bold" />
        </MenuButton>
        <MenuButton
          active={state.isCode}
          onClick={() => toggleStyle(editor, "code")}
        >
          <CodeIcon size={16} weight="bold" />
        </MenuButton>
        <MenuButton
          active={state.isLink || showLinkInput}
          onClick={handleLinkButtonClick}
        >
          <LinkIcon size={16} weight="bold" />
        </MenuButton>
      </BubbleMenuToolbar>
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
    </>
  );
}
