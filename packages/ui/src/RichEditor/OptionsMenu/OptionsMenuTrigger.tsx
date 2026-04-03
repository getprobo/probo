// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { DotsSixVerticalIcon } from "@phosphor-icons/react";
import { NodeSelection } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { type Editor } from "@tiptap/react";
import type { DragEvent } from "react";

import { getBlockNode } from "../_lib/getBlockNode";
import { useBlockTrigger } from "../_lib/useBlockTrigger";

import type { OptionsMenuFloating } from "./OptionsMenu";
import { optionsMenuVariants } from "./variants";

const { trigger } = optionsMenuVariants();

function startDrag(view: EditorView, slice: ReturnType<NodeSelection["content"]>, node: NodeSelection) {
  view.dragging = { slice, move: true, node } as typeof view.dragging;
}

type OptionsMenuTriggerProps = {
  editor: Editor;
  hoveredBlock: HTMLElement | null;
  setHoveredBlock: (block: HTMLElement | null) => void;
  setTriggerEl: OptionsMenuFloating["setTriggerEl"];
  menuRefs: OptionsMenuFloating["menuRefs"];
  getReferenceProps: OptionsMenuFloating["getReferenceProps"];
};

export function OptionsMenuTrigger({
  editor,
  hoveredBlock,
  setHoveredBlock,
  setTriggerEl,
  menuRefs,
  getReferenceProps,
}: OptionsMenuTriggerProps) {
  const { triggerRefs, triggerStyles, isPositioned } = useBlockTrigger(hoveredBlock, 12);

  const onDragStart = (e: DragEvent<HTMLButtonElement>) => {
    if (!hoveredBlock) return;
    const data = getBlockNode(editor, hoveredBlock);
    if (!data) return;

    try {
      const view = editor.view;
      const selection = NodeSelection.create(view.state.doc, data.pos);
      const slice = selection.content();

      const { tr } = view.state;
      tr.setSelection(selection);
      view.dispatch(tr);

      if (e.dataTransfer) {
        e.dataTransfer.clearData();
        e.dataTransfer.setData("text/plain", "");
        e.dataTransfer.effectAllowed = "move";

        const wrapper = document.createElement("div");
        wrapper.append(hoveredBlock.cloneNode(true));
        wrapper.style.position = "absolute";
        wrapper.style.top = "-10000px";
        document.body.append(wrapper);
        e.dataTransfer.setDragImage(wrapper, 0, 0);
        document.addEventListener("dragend", () => wrapper.remove(), { once: true });
      }

      startDrag(view, slice, selection);
    } catch {
      // Block may no longer be in the document
    }
  };

  return (
    <button
      ref={(node) => {
        triggerRefs.setFloating(node);
        setTriggerEl(node);
        menuRefs.setReference(node);
      }}
      {...getReferenceProps()}
      draggable
      onDragStart={onDragStart}
      onDragEnd={() => setHoveredBlock(null)}
      type="button"
      style={{
        ...triggerStyles,
        visibility: isPositioned ? "visible" : "hidden",
      }}
      className={trigger()}
    >
      <DotsSixVerticalIcon size={16} weight="bold" />
    </button>
  );
}
