// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { ArrowSquareOutIcon, CodeIcon, LinkIcon, TextBIcon, TextItalicIcon, TextStrikethroughIcon, TextUnderlineIcon, TrashIcon } from "@phosphor-icons/react";
import { Editor, useEditorState } from "@tiptap/react";
import { type FocusEvent, type KeyboardEvent, useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { tv } from "tailwind-variants/lite";

import { toggleStyle } from "./_lib/formatting";
import { useLinkEditor } from "./_lib/useLinkEditor";
import { BubbleMenuToolbar } from "./BubbleMenu";
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
  const { t } = useTranslation();
  const link = useLinkEditor(editor);
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
        };
      }

      return {
        isBold: current.isActive("bold"),
        isItalic: current.isActive("italic"),
        isUnderline: current.isActive("underline"),
        isStrike: current.isActive("strike"),
        isCode: current.isActive("code"),
      };
    },
  });

  useEffect(() => {
    if (link.open) {
      linkInputRef.current?.focus();
    }
  }, [link.open]);

  function submitLink() {
    if (!link.submit() && !editor.isDestroyed) {
      editor.commands.focus();
    }
  }

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter") {
      e.preventDefault();
      submitLink();
    }
    if (e.key === "Escape") {
      e.preventDefault();
      link.setOpen(false);
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
          active={link.isLink || link.open}
          onClick={() => link.setOpen(!link.open)}
        >
          <LinkIcon size={16} weight="bold" />
        </MenuButton>
      </BubbleMenuToolbar>
      {link.open && (
        <div className={linkInput()} onBlur={handleLinkInputBlur}>
          <input
            ref={linkInputRef}
            type="url"
            placeholder="https://…"
            value={link.url}
            onChange={e => link.setUrl(e.target.value)}
            onKeyDown={handleKeyDown}
            className="min-w-0 flex-1 bg-transparent text-txt-primary outline-none placeholder:text-txt-quaternary"
          />
          {link.url.trim() !== "" && (
            <button
              type="button"
              aria-label={t("ui.richEditor.toolbar.openLink")}
              onMouseDown={e => e.preventDefault()}
              onClick={link.openUrl}
              className={linkAction()}
            >
              <ArrowSquareOutIcon size={16} weight="bold" />
            </button>
          )}
          {link.isLink && (
            <button
              type="button"
              aria-label={t("ui.richEditor.toolbar.removeLink")}
              onMouseDown={e => e.preventDefault()}
              onClick={link.remove}
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
