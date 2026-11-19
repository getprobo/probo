// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { ArrowSquareOutIcon, LinkIcon, TrashIcon } from "@phosphor-icons/react";
import { type Editor, useEditorState } from "@tiptap/react";
import { type KeyboardEvent, useState } from "react";
import { useTranslation } from "react-i18next";

import { IconButton } from "../../v2/IconButton/IconButton";
import { Popover } from "../../v2/Popover/Popover";
import { PopoverPopup } from "../../v2/Popover/PopoverPopup";
import { PopoverTrigger } from "../../v2/Popover/PopoverTrigger";
import { applyLink } from "../_lib/formatting";
import { linkEditorPluginKey, openLink } from "../LinkExtension";

import { refocusEditor } from "./refocusEditor";
import { RichEditorToolbarButton } from "./RichEditorToolbarButton";
import { richEditorToolbar } from "./variants";

type RichEditorToolbarLinkProps = {
  editor: Editor;
  disabled: boolean;
};

// Opens from the toolbar button, or when a link in the text is clicked while
// editing: the link extension then selects it and posts an editor request.
export function RichEditorToolbarLink({ editor, disabled }: RichEditorToolbarLinkProps) {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [url, setUrl] = useState("");
  const state = useEditorState({
    editor,
    selector: ({ editor: current }) => {
      if (current.isDestroyed) {
        return { empty: true, isLink: false, href: undefined, request: null };
      }

      return {
        empty: current.state.selection.empty,
        isLink: current.isActive("link"),
        href: current.getAttributes("link").href as string | undefined,
        request: linkEditorPluginKey.getState(current.state) ?? null,
      };
    },
  });
  const requestId = state.request?.id ?? 0;
  const [handledRequestId, setHandledRequestId] = useState(requestId);
  const { linkPopup, linkInput } = richEditorToolbar();

  if (requestId !== handledRequestId) {
    setHandledRequestId(requestId);
    setUrl(state.request?.href ?? "");
    setOpen(true);
  }

  function handleOpenChange(next: boolean) {
    if (next) {
      setUrl(state.href ?? "");
    }
    setOpen(next);
  }

  function submit() {
    const next = url.trim();
    setOpen(false);
    if (next && next !== (state.href ?? "")) {
      applyLink(editor, next);
    }
  }

  function remove() {
    setOpen(false);
    applyLink(editor, null);
  }

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === "Enter") {
      event.preventDefault();
      submit();
    }
  }

  return (
    <Popover open={open} onOpenChange={handleOpenChange}>
      <PopoverTrigger
        render={(
          <RichEditorToolbarButton
            label={t("ui.richEditor.toolbar.link")}
            pressed={state.isLink || open}
            disabled={disabled || (state.empty && !open)}
          />
        )}
      >
        <LinkIcon weight="bold" />
      </PopoverTrigger>
      <PopoverPopup
        className={linkPopup()}
        finalFocus={refocusEditor(editor)}
        data-rich-editor-toolbar-popup=""
      >
        <input
          type="url"
          className={linkInput()}
          aria-label={t("ui.richEditor.toolbar.linkUrl")}
          placeholder="https://…"
          value={url}
          onChange={event => setUrl(event.target.value)}
          onKeyDown={handleKeyDown}
        />
        {url.trim() !== "" && (
          <IconButton
            size={1}
            variant="ghost"
            color="neutral"
            aria-label={t("ui.richEditor.toolbar.openLink")}
            onClick={() => openLink(url.trim())}
          >
            <ArrowSquareOutIcon weight="bold" />
          </IconButton>
        )}
        {state.isLink && (
          <IconButton
            size={1}
            variant="ghost"
            color="neutral"
            aria-label={t("ui.richEditor.toolbar.removeLink")}
            onClick={remove}
          >
            <TrashIcon weight="bold" />
          </IconButton>
        )}
      </PopoverPopup>
    </Popover>
  );
}
