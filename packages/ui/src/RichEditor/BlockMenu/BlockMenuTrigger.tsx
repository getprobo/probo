// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { PlusIcon } from "@phosphor-icons/react";
import { type Editor } from "@tiptap/react";

import { getSlashStorage } from "../_lib/getSlashStorage";
import { useBlockTrigger } from "../_lib/useBlockTrigger";
import { activateSlashCommand } from "../SlashCommandExtension";

import { blockMenuVariants } from "./variants";

const { trigger } = blockMenuVariants();

type BlockMenuTriggerProps = {
  editor: Editor;
  hoveredBlock: HTMLElement;
};

export function BlockMenuTrigger({ editor, hoveredBlock }: BlockMenuTriggerProps) {
  const { triggerRefs, triggerStyles, isPositioned } = useBlockTrigger(hoveredBlock, 32);

  const handleTriggerClick = () => {
    try {
      const pos = editor.view.posAtDOM(hoveredBlock, 0);
      const $pos = editor.state.doc.resolve(pos);

      const rootPos = $pos.depth >= 1 ? $pos.before(1) : pos;
      const rootNode = $pos.depth >= 1 ? $pos.node(1) : $pos.nodeAfter;

      if (rootNode && rootNode.isTextblock && rootNode.content.size === 0) {
        const textPos = rootPos + 1;

        editor.chain()
          .focus()
          .setTextSelection(textPos)
          .insertContent("/")
          .run();

        const s = getSlashStorage(editor);
        if (s) activateSlashCommand(s, textPos);
        return;
      }

      let insertPos: number;
      if ($pos.depth >= 1) {
        insertPos = rootPos + rootNode!.nodeSize;
      } else {
        const nodeAfter = $pos.nodeAfter;
        insertPos = pos + (nodeAfter?.nodeSize ?? 1);
      }

      const textPos = insertPos + 1;

      editor.chain()
        .focus()
        .insertContentAt(insertPos, { type: "paragraph" })
        .setTextSelection(textPos)
        .insertContent("/")
        .run();

      const s = getSlashStorage(editor);
      if (s) activateSlashCommand(s, textPos);
    } catch {
      // Block may no longer be in the document
    }
  };

  return (
    <button
      ref={(node) => {
        triggerRefs.setFloating(node);
      }}
      onClick={handleTriggerClick}
      onMouseDown={e => e.preventDefault()}
      type="button"
      style={{
        ...triggerStyles,
        visibility: isPositioned ? "visible" : "hidden",
      }}
      className={trigger()}
    >
      <PlusIcon size={14} weight="bold" />
    </button>
  );
}
