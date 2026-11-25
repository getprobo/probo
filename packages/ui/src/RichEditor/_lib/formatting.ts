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

import type { Schema } from "@tiptap/pm/model";
import type { Editor } from "@tiptap/react";

import {
  applySuggestingLink,
  applySuggestingStyle,
  suggestingControllerFrom,
} from "../SuggestionExtension";
import type { SuggestionStyleName } from "../suggestionStyle";

export type FormattingStyle = Exclude<SuggestionStyleName, "link">;

const toggles: Record<FormattingStyle, (editor: Editor) => void> = {
  bold: editor => editor.chain().focus().toggleBold().run(),
  italic: editor => editor.chain().focus().toggleItalic().run(),
  underline: editor => editor.chain().focus().toggleUnderline().run(),
  strike: editor => editor.chain().focus().toggleStrike().run(),
  code: editor => editor.chain().focus().toggleCode().run(),
};

// `destroy()` nulls `schema`, so render paths must check it before reading marks.
export function hasStyleMark(editor: Editor, style: FormattingStyle): boolean {
  const schema: Schema | null = editor.schema;
  return schema != null && schema.marks[style] != null;
}

// Suggesting mode records the change as a format suggestion. Otherwise the
// mark is toggled directly.
export function toggleStyle(editor: Editor, style: FormattingStyle) {
  if (editor.isDestroyed) return;
  const controller = suggestingControllerFrom(editor);
  if (controller && applySuggestingStyle(editor.view, controller, style)) return;
  toggles[style](editor);
}

// A null href removes the link.
export function applyLink(editor: Editor, href: string | null) {
  if (editor.isDestroyed) return;
  const controller = suggestingControllerFrom(editor);
  if (controller && applySuggestingLink(editor.view, controller, href)) return;
  if (href == null) {
    editor.chain().focus().unsetLink().run();
  } else {
    editor.chain().focus().setLink({ href }).run();
  }
}
