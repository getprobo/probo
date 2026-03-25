// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import type { Node as PmNode } from "@tiptap/pm/model";
import type { Editor } from "@tiptap/react";

type BlockNodeData = {
  node: PmNode;
  pos: number;
};

export function getBlockNode(
  editor: Editor,
  block: HTMLElement,
): BlockNodeData | null {
  try {
    const pos = editor.view.posAtDOM(block, 0);
    const $pos = editor.state.doc.resolve(pos);
    if ($pos.depth >= 1) {
      return { node: $pos.node(1), pos: $pos.before(1) };
    }
    const nodeAfter = $pos.nodeAfter;
    if (nodeAfter) {
      return { node: nodeAfter, pos };
    }
    return null;
  } catch {
    return null;
  }
}

export function isBlockNodeType(
  editor: Editor,
  block: HTMLElement,
  type: string,
  attrs?: Record<string, unknown>,
): boolean {
  const data = getBlockNode(editor, block);
  if (!data) return false;
  if (data.node.type.name !== type) return false;
  if (attrs) {
    return Object.entries(attrs).every(
      ([key, value]) => data.node.attrs[key] === value,
    );
  }
  return true;
}
