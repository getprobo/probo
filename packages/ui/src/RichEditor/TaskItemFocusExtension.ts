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

import { Extension } from "@tiptap/core";
import { Plugin } from "@tiptap/pm/state";

const taskCheckboxSelector = "ul[data-type=\"taskList\"] input[type=\"checkbox\"]";

type CheckboxContainer = Pick<HTMLElement, "querySelectorAll">;

// Browsers keep native <input> elements in the tab order even inside a
// contenteditable="false" container, so the pointer-events rule in
// rich-editor.css (which only blocks mouse interaction) is not enough to
// keep a read-only document's checkboxes from being tabbed to and toggled
// with the keyboard.
export function syncTaskCheckboxFocusability(
  root: CheckboxContainer,
  editable: boolean,
): void {
  const checkboxes = root.querySelectorAll<HTMLInputElement>(
    taskCheckboxSelector,
  );
  checkboxes.forEach((checkbox) => {
    checkbox.tabIndex = editable ? 0 : -1;
  });
}

export const TaskItemFocusExtension = Extension.create({
  name: "taskItemFocus",

  addProseMirrorPlugins() {
    const { editor } = this;

    const sync = (view: { dom: CheckboxContainer }) => {
      syncTaskCheckboxFocusability(view.dom, editor.isEditable);
    };

    return [
      new Plugin({
        view(view) {
          sync(view);

          return {
            update: sync,
          };
        },
      }),
    ];
  },
});
