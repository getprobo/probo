// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { CodeBlockIcon, ListBulletsIcon, ListNumbersIcon, QuotesIcon, TextHFourIcon, TextHOneIcon, TextHThreeIcon, TextHTwoIcon, TextTIcon } from "@phosphor-icons/react";
import { TextSelection } from "@tiptap/pm/state";
import { type Editor } from "@tiptap/react";

import { getBlockNode, isBlockNodeType } from "../_lib/getBlockNode";
import { MenuButton } from "../MenuButton";

import type { OptionsMenuFloating } from "./OptionsMenu";
import { optionsMenuVariants } from "./variants";

const { menu } = optionsMenuVariants();

type OptionsMenuContentProps = {
  editor: Editor;
  hoveredBlock: HTMLElement | null;
  setMenuOpen: React.Dispatch<React.SetStateAction<boolean>>;
  setDropdownEl: OptionsMenuFloating["setDropdownEl"];
  menuRefs: OptionsMenuFloating["menuRefs"];
  menuStyles: OptionsMenuFloating["menuStyles"];
  getFloatingProps: OptionsMenuFloating["getFloatingProps"];
};

export function OptionsMenuContent({
  editor,
  hoveredBlock,
  setMenuOpen,
  setDropdownEl,
  menuRefs,
  menuStyles,
  getFloatingProps,
}: OptionsMenuContentProps) {
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

  return (
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
        active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "heading", { level: 4 })}
        onClick={() => handleAction(chain => chain.toggleHeading({ level: 4 }))}
      >
        <TextHFourIcon size={16} weight="bold" />
        Heading 4
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
  );
}
