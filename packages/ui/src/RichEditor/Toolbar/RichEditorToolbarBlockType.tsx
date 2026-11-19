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

import { CaretDownIcon } from "@phosphor-icons/react";
import { type Editor, useEditorState } from "@tiptap/react";
import { useTranslation } from "react-i18next";

import { Button } from "../../v2/Button/Button";
import { Dropdown } from "../../v2/Dropdown/Dropdown";
import { DropdownPopup } from "../../v2/Dropdown/DropdownPopup";
import { DropdownRadioGroup } from "../../v2/Dropdown/DropdownRadioGroup";
import { DropdownRadioItem } from "../../v2/Dropdown/DropdownRadioItem";
import { DropdownTrigger } from "../../v2/Dropdown/DropdownTrigger";
import { currentBlockType } from "../_lib/currentBlockType";
import { blockTypeItems, runBlockItem } from "../BlockMenu/blockItems";

import { refocusEditor } from "./refocusEditor";
import { richEditorToolbar } from "./variants";

type RichEditorToolbarBlockTypeProps = {
  editor: Editor;
  disabled: boolean;
};

// Turns the block at the selection into another type. Picking the current
// type does nothing, since the list and heading commands would toggle it off.
export function RichEditorToolbarBlockType({ editor, disabled }: RichEditorToolbarBlockTypeProps) {
  const { t } = useTranslation();
  const current = useEditorState({
    editor,
    selector: ({ editor: e }) => (e.isDestroyed ? null : currentBlockType(e.state)),
  });
  const { blockTypeTrigger, triggerLabel, triggerCaret } = richEditorToolbar();

  function handleValueChange(value: unknown) {
    const item = blockTypeItems.find(candidate => candidate.id === value);
    if (item) {
      runBlockItem(editor, item);
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
            className={blockTypeTrigger()}
            iconEnd={<CaretDownIcon className={triggerCaret()} />}
          />
        )}
      >
        <span className={triggerLabel()}>
          {t(current ? `ui.richEditor.toolbar.blocks.${current}` : "ui.richEditor.toolbar.blockType")}
        </span>
      </DropdownTrigger>
      <DropdownPopup
        size={1}
        variant="soft"
        finalFocus={refocusEditor(editor)}
        data-rich-editor-toolbar-popup=""
      >
        <DropdownRadioGroup value={current} onValueChange={handleValueChange}>
          {blockTypeItems.map(item => (
            <DropdownRadioItem key={item.id} value={item.id} iconStart={<item.icon weight="bold" />}>
              {t(`ui.richEditor.toolbar.blocks.${item.id}`)}
            </DropdownRadioItem>
          ))}
        </DropdownRadioGroup>
      </DropdownPopup>
    </Dropdown>
  );
}
