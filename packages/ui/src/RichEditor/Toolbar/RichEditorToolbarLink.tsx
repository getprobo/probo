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
import type { Editor } from "@tiptap/react";
import { type KeyboardEvent } from "react";
import { useTranslation } from "react-i18next";

import { IconButton } from "../../v2/IconButton/IconButton";
import { Popover } from "../../v2/Popover/Popover";
import { PopoverPopup } from "../../v2/Popover/PopoverPopup";
import { PopoverTrigger } from "../../v2/Popover/PopoverTrigger";
import { useLinkEditor } from "../_lib/useLinkEditor";

import { refocusEditor } from "./refocusEditor";
import { RichEditorToolbarButton } from "./RichEditorToolbarButton";
import { richEditorToolbar } from "./variants";

type RichEditorToolbarLinkProps = {
  editor: Editor;
  disabled: boolean;
};

export function RichEditorToolbarLink({ editor, disabled }: RichEditorToolbarLinkProps) {
  const { t } = useTranslation();
  const { open, setOpen, url, setUrl, isLink, empty, submit, remove, openUrl } = useLinkEditor(editor);
  const { linkPopup, linkInput } = richEditorToolbar();

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === "Enter") {
      event.preventDefault();
      submit();
    }
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={(
          <RichEditorToolbarButton
            label={t("ui.richEditor.toolbar.link")}
            pressed={isLink || open}
            disabled={disabled || (empty && !open)}
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
            onClick={openUrl}
          >
            <ArrowSquareOutIcon weight="bold" />
          </IconButton>
        )}
        {isLink && (
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
