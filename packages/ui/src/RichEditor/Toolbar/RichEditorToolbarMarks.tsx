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

import {
  CodeIcon,
  type Icon,
  TextBIcon,
  TextItalicIcon,
  TextStrikethroughIcon,
  TextUnderlineIcon,
} from "@phosphor-icons/react";
import { type Editor, useEditorState } from "@tiptap/react";
import { useTranslation } from "react-i18next";

import { type FormattingStyle, toggleStyle } from "../_lib/formatting";

import { RichEditorToolbarButton } from "./RichEditorToolbarButton";

const styles: { style: FormattingStyle; icon: Icon }[] = [
  { style: "bold", icon: TextBIcon },
  { style: "italic", icon: TextItalicIcon },
  { style: "underline", icon: TextUnderlineIcon },
  { style: "strike", icon: TextStrikethroughIcon },
  { style: "code", icon: CodeIcon },
];

type RichEditorToolbarMarksProps = {
  editor: Editor;
  disabled: boolean;
};

export function RichEditorToolbarMarks({ editor, disabled }: RichEditorToolbarMarksProps) {
  const { t } = useTranslation();
  const state = useEditorState({
    editor,
    selector: ({ editor: current }) => {
      if (current.isDestroyed) {
        return { active: [] as FormattingStyle[] };
      }

      return {
        active: styles.map(({ style }) => style).filter(style => current.isActive(style)),
      };
    },
  });

  return styles
    .filter(({ style }) => editor.schema.marks[style] != null)
    .map(({ style, icon: StyleIcon }) => (
      <RichEditorToolbarButton
        key={style}
        label={t(`ui.richEditor.toolbar.styles.${style}`)}
        pressed={state.active.includes(style)}
        disabled={disabled}
        onMouseDown={event => event.preventDefault()}
        onClick={() => toggleStyle(editor, style)}
      >
        <StyleIcon weight="bold" />
      </RichEditorToolbarButton>
    ));
}
