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

import { CaretDownIcon, PlusIcon } from "@phosphor-icons/react";
import type { Editor } from "@tiptap/react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { Button } from "../../v2/Button/Button";
import { Dropdown } from "../../v2/Dropdown/Dropdown";
import { DropdownItem } from "../../v2/Dropdown/DropdownItem";
import { DropdownPopup } from "../../v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "../../v2/Dropdown/DropdownTrigger";
import {
  type InsertItem,
  insertItems,
  insertMermaidFromToolbar,
  isPickerItem,
  runBlockItem,
} from "../BlockMenu/blockItems";

import { refocusEditor } from "./refocusEditor";
import { richEditorToolbar } from "./variants";

type RichEditorToolbarInsertProps = {
  editor: Editor;
  disabled: boolean;
  // Image and file entries only show when uploads are allowed.
  onPickFiles?: (kind: "image" | "file") => void;
};

export function RichEditorToolbarInsert({ editor, disabled, onPickFiles }: RichEditorToolbarInsertProps) {
  const { t } = useTranslation();
  const items = onPickFiles ? insertItems : insertItems.filter(item => !isPickerItem(item));
  const { triggerLabel, triggerCaret } = richEditorToolbar();
  // Until the text is clicked, the selection sits at the start of the
  // document. Inserting there would change the first block.
  const [hadCaret, setHadCaret] = useState(editor.isFocused);

  useEffect(() => {
    if (hadCaret) return;
    const onFocus = () => setHadCaret(true);
    editor.on("focus", onFocus);
    return () => {
      editor.off("focus", onFocus);
    };
  }, [editor, hadCaret]);

  function insert(item: InsertItem) {
    if (!hadCaret && !editor.isDestroyed) {
      editor.commands.focus("end");
    }
    if (item.id === "mermaid") {
      insertMermaidFromToolbar(editor);
    } else {
      runBlockItem(editor, item, onPickFiles);
    }
  }

  return (
    <Dropdown>
      <DropdownTrigger
        disabled={disabled}
        render={(
          <Button
            size={2}
            variant="ghost"
            color="neutral"
            highContrast
            iconStart={<PlusIcon weight="bold" />}
            iconEnd={<CaretDownIcon className={triggerCaret()} />}
          />
        )}
      >
        <span className={triggerLabel()}>{t("ui.richEditor.toolbar.insert")}</span>
      </DropdownTrigger>
      <DropdownPopup
        size={1}
        variant="soft"
        finalFocus={refocusEditor(editor)}
        data-rich-editor-toolbar-popup=""
      >
        {items.map(item => (
          <DropdownItem
            key={item.id}
            color="neutral"
            iconStart={<item.icon weight="bold" />}
            onClick={() => insert(item)}
          >
            {t(`ui.richEditor.toolbar.inserts.${item.id}`)}
          </DropdownItem>
        ))}
      </DropdownPopup>
    </Dropdown>
  );
}
