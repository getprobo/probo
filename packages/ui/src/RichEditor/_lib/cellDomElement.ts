// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import type { Editor } from "@tiptap/react";

export function cellDomElement(
  editor: Editor,
  cellPos: number,
): HTMLElement | null {
  const dom = editor.view.domAtPos(cellPos + 1);
  let el: Node | null = dom.node;
  if (el.nodeType === Node.TEXT_NODE) el = el.parentElement;
  while (el && !(el instanceof HTMLTableCellElement)) {
    el = (el as HTMLElement).parentElement;
  }
  return el as HTMLElement | null;
}
