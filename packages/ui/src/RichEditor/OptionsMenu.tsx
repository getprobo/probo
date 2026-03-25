// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import {
  autoUpdate,
  flip,
  offset,
  shift,
  useClick,
  useDismiss,
  useFloating,
  useFloatingRootContext,
  useInteractions,
} from "@floating-ui/react";
import { CodeBlockIcon, DotsSixVerticalIcon, ListBulletsIcon, ListNumbersIcon, QuotesIcon, TextHOneIcon, TextHThreeIcon, TextHTwoIcon, TextTIcon } from "@phosphor-icons/react";
import { NodeSelection, TextSelection } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { type Editor } from "@tiptap/react";
import { type DragEvent, useState } from "react";
import { tv } from "tailwind-variants";

import { getBlockNode, isBlockNodeType } from "./_lib/getBlockNode";
import { useBlockTrigger } from "./_lib/useBlockTrigger";
import { useHoveredBlock } from "./_lib/useHoveredBlock";
import { MenuButton } from "./MenuButton";

const optionsMenuVariants = tv({
  slots: {
    trigger: [
      "z-10 flex size-6 items-center justify-center",
      "rounded text-txt-tertiary hover:bg-subtle hover:text-txt-primary cursor-grab",
    ],
    menu: ["rounded-lg border border-border-mid bg-level-0 p-1 shadow-md z-20"],
  },
});

const { trigger, menu } = optionsMenuVariants();

function startDrag(view: EditorView, slice: ReturnType<NodeSelection["content"]>, node: NodeSelection) {
  view.dragging = { slice, move: true, node } as typeof view.dragging;
}

type OptionsMenuProps = {
  editor: Editor;
};

export function OptionsMenu({ editor }: OptionsMenuProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [triggerEl, setTriggerEl] = useState<Element | null>(null);
  const [dropdownEl, setDropdownEl] = useState<HTMLElement | null>(null);

  const { hoveredBlock, setHoveredBlock } = useHoveredBlock(editor, menuOpen);
  const { triggerRefs, triggerStyles, isPositioned } = useBlockTrigger(hoveredBlock, 16);

  const menuRootContext = useFloatingRootContext({
    open: menuOpen,
    onOpenChange: setMenuOpen,
    elements: { reference: triggerEl, floating: dropdownEl },
  });

  const { refs: menuRefs, floatingStyles: menuStyles } = useFloating({
    rootContext: menuRootContext,
    strategy: "fixed",
    placement: "bottom-start",
    middleware: [offset(4), flip(), shift()],
    whileElementsMounted: autoUpdate,
  });

  const click = useClick(menuRootContext);
  const dismiss = useDismiss(menuRootContext);
  const { getReferenceProps, getFloatingProps } = useInteractions([click, dismiss]);

  const shouldShow = hoveredBlock != null || menuOpen;

  if (!shouldShow) return null;

  const handleAction = (
    applyCommand: (chain: ReturnType<typeof editor.chain>) => ReturnType<typeof editor.chain>,
  ) => {
    if (!hoveredBlock) {
      setMenuOpen(false);
      return;
    }
    const data = getBlockNode(editor, hoveredBlock);
    if (!data) {
      setMenuOpen(false);
      return;
    }

    try {
      if (!data.node.isTextblock) {
        if (!data.node.firstChild) {
          const paragraph = editor.state.schema.nodes.paragraph.create();
          editor.chain()
            .focus()
            .command(({ tr }) => {
              tr.replaceWith(data.pos, data.pos + data.node.nodeSize, paragraph);
              return true;
            })
            .run();

          const $near = editor.state.doc.resolve(data.pos + 1);
          const textPos = TextSelection.near($near).from;

          applyCommand(
            editor.chain()
              .focus()
              .setTextSelection(textPos),
          ).run();

          setMenuOpen(false);
          return;
        }

        let textBlock = data.node.firstChild;
        if (!textBlock) return;
        while (!textBlock.isTextblock && textBlock.firstChild) {
          textBlock = textBlock.firstChild;
        }
        if (!textBlock.isTextblock) return;

        const firstChildSize = data.node.firstChild.nodeSize;
        const paragraph = editor.state.schema.nodes.paragraph.create(
          null,
          textBlock.content,
        );

        editor.chain()
          .focus()
          .command(({ tr }) => {
            tr.insert(data.pos, paragraph);
            const wrapperPos = data.pos + paragraph.nodeSize;
            const wrapperNode = tr.doc.nodeAt(wrapperPos);
            if (!wrapperNode) return false;

            if (wrapperNode.childCount <= 1) {
              tr.delete(wrapperPos, wrapperPos + wrapperNode.nodeSize);
            } else {
              tr.delete(wrapperPos + 1, wrapperPos + 1 + firstChildSize);
            }

            return true;
          })
          .run();
      }

      const $near = editor.state.doc.resolve(data.pos + 1);
      const textPos = TextSelection.near($near).from;

      applyCommand(
        editor.chain()
          .focus()
          .setTextSelection(textPos),
      ).run();
    } catch {
      // Block may no longer be in the document
    }

    setMenuOpen(false);
  };

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
        document.addEventListener("drop", () => wrapper.remove(), { once: true });
      }

      startDrag(view, slice, selection);
    } catch {
      // Block may no longer be in the document
    }
  };

  return (
    <>
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
        <DotsSixVerticalIcon size={20} weight="bold" />
      </button>
      {menuOpen && (
        <div
          ref={(node) => {
            setDropdownEl(node);
            menuRefs.setFloating(node);
          }}
          style={menuStyles}
          {...getFloatingProps()}
          onMouseDown={e => e.preventDefault()}
          className={menu()}
        >
          <div className="p-1 font-semibold text-sm">Turn into</div>
          <MenuButton
            active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "paragraph")}
            onClick={() => handleAction(chain => chain.setParagraph())}
          >
            <TextTIcon size={16} weight="bold" />
            Text
          </MenuButton>
          <MenuButton
            active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "heading", { level: 1 })}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 1 }))}
          >
            <TextHOneIcon size={16} weight="bold" />
            Heading 1
          </MenuButton>
          <MenuButton
            active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "heading", { level: 2 })}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 2 }))}
          >
            <TextHTwoIcon size={16} weight="bold" />
            Heading 2
          </MenuButton>
          <MenuButton
            active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "heading", { level: 3 })}
            onClick={() => handleAction(chain => chain.toggleHeading({ level: 3 }))}
          >
            <TextHThreeIcon size={16} weight="bold" />
            Heading 3
          </MenuButton>
          <MenuButton
            active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "bulletList")}
            onClick={() => handleAction(chain => chain.toggleBulletList())}
          >
            <ListBulletsIcon size={16} weight="bold" />
            Bullet List
          </MenuButton>
          <MenuButton
            active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "orderedList")}
            onClick={() => handleAction(chain => chain.toggleOrderedList())}
          >
            <ListNumbersIcon size={16} weight="bold" />
            Ordered List
          </MenuButton>
          <MenuButton
            active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "codeBlock")}
            onClick={() => handleAction(chain => chain.toggleCodeBlock())}
          >
            <CodeBlockIcon size={16} weight="bold" />
            Code Block
          </MenuButton>
          <MenuButton
            active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "blockquote")}
            onClick={() => handleAction(chain => chain.toggleBlockquote())}
          >
            <QuotesIcon size={16} weight="bold" />
            Blockquote
          </MenuButton>
        </div>
      )}
    </>
  );
}
