// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { CodeBlockIcon, ListBulletsIcon, ListNumbersIcon, QuotesIcon, TextHFourIcon, TextHOneIcon, TextHThreeIcon, TextHTwoIcon, TextTIcon } from "@phosphor-icons/react";
import { Fragment, type Node as PmNode } from "@tiptap/pm/model";
import { TextSelection } from "@tiptap/pm/state";
import { type Editor } from "@tiptap/react";

import { getBlockNode, isBlockNodeType } from "../_lib/getBlockNode";
import { MenuButton } from "../MenuButton";

import type { OptionsMenuFloating } from "./OptionsMenu";
import { optionsMenuVariants } from "./variants";

const { menu } = optionsMenuVariants();

type TargetKind = "textblock" | "wrapper";

function collectTextBlocks(node: PmNode): PmNode[] {
  const result: PmNode[] = [];
  node.forEach((child) => {
    if (child.isTextblock) {
      result.push(child);
    } else {
      result.push(...collectTextBlocks(child));
    }
  });
  return result;
}

type DecomposeResult = {
  targetContent: Fragment | null;
  remaining: PmNode[];
};

function decomposeWrapper(node: PmNode): DecomposeResult {
  const firstChild = node.firstChild;
  if (!firstChild) {
    return { targetContent: null, remaining: [] };
  }

  let targetContent: Fragment | null = null;
  const promotedSiblings: PmNode[] = [];

  if (firstChild.isTextblock) {
    targetContent = firstChild.content;
  } else {
    firstChild.forEach((child) => {
      if (targetContent === null && child.isTextblock) {
        targetContent = child.content;
      } else {
        promotedSiblings.push(child);
      }
    });
  }

  const remainingChildren: PmNode[] = [];
  node.forEach((_child, _offset, index) => {
    if (index > 0) remainingChildren.push(_child);
  });

  const remaining: PmNode[] = [...promotedSiblings];
  if (remainingChildren.length > 0) {
    remaining.push(node.copy(Fragment.from(remainingChildren)));
  }

  return { targetContent, remaining };
}

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
    targetKind: TargetKind,
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
      const { node, pos } = data;
      const schema = editor.state.schema;

      if (node.isTextblock) {
        const $near = editor.state.doc.resolve(pos + 1);
        const textPos = TextSelection.near($near).from;

        applyCommand(
          editor.chain()
            .focus()
            .setTextSelection(textPos),
        ).run();
      } else if (targetKind === "textblock") {
        const { targetContent, remaining } = decomposeWrapper(node);
        if (!targetContent) {
          setMenuOpen(false);
          return;
        }

        const target = schema.nodes.paragraph.create(null, targetContent);
        const replacements = [target, ...remaining];

        editor.chain()
          .focus()
          .command(({ tr }) => {
            tr.replaceWith(
              pos,
              pos + node.nodeSize,
              Fragment.from(replacements),
            );
            return true;
          })
          .run();

        const $near = editor.state.doc.resolve(pos + 1);
        const textPos = TextSelection.near($near).from;

        applyCommand(
          editor.chain()
            .focus()
            .setTextSelection(textPos),
        ).run();
      } else {
        const textBlocks = collectTextBlocks(node);
        if (textBlocks.length === 0) {
          setMenuOpen(false);
          return;
        }

        const paragraphs = textBlocks.map(tb =>
          schema.nodes.paragraph.create(null, tb.content),
        );

        editor.chain()
          .focus()
          .command(({ tr }) => {
            tr.replaceWith(
              pos,
              pos + node.nodeSize,
              Fragment.from(paragraphs),
            );
            return true;
          })
          .run();

        let totalSize = 0;
        for (const p of paragraphs) {
          totalSize += p.nodeSize;
        }
        const from = pos + 1;
        const to = pos + totalSize - 1;

        applyCommand(
          editor.chain()
            .focus()
            .setTextSelection({ from, to }),
        ).run();
      }
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
        onClick={() => handleAction(chain => chain.setParagraph(), "textblock")}
      >
        <TextTIcon size={16} weight="bold" />
        Text
      </MenuButton>
      <MenuButton
        active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "heading", { level: 1 })}
        onClick={() => handleAction(chain => chain.toggleHeading({ level: 1 }), "textblock")}
      >
        <TextHOneIcon size={16} weight="bold" />
        Heading 1
      </MenuButton>
      <MenuButton
        active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "heading", { level: 2 })}
        onClick={() => handleAction(chain => chain.toggleHeading({ level: 2 }), "textblock")}
      >
        <TextHTwoIcon size={16} weight="bold" />
        Heading 2
      </MenuButton>
      <MenuButton
        active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "heading", { level: 3 })}
        onClick={() => handleAction(chain => chain.toggleHeading({ level: 3 }), "textblock")}
      >
        <TextHThreeIcon size={16} weight="bold" />
        Heading 3
      </MenuButton>
      <MenuButton
        active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "heading", { level: 4 })}
        onClick={() => handleAction(chain => chain.toggleHeading({ level: 4 }), "textblock")}
      >
        <TextHFourIcon size={16} weight="bold" />
        Heading 4
      </MenuButton>
      <MenuButton
        active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "bulletList")}
        onClick={() => handleAction(chain => chain.toggleBulletList(), "wrapper")}
      >
        <ListBulletsIcon size={16} weight="bold" />
        Bullet List
      </MenuButton>
      <MenuButton
        active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "orderedList")}
        onClick={() => handleAction(chain => chain.toggleOrderedList(), "wrapper")}
      >
        <ListNumbersIcon size={16} weight="bold" />
        Ordered List
      </MenuButton>
      <MenuButton
        active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "codeBlock")}
        onClick={() => handleAction(chain => chain.toggleCodeBlock(), "textblock")}
      >
        <CodeBlockIcon size={16} weight="bold" />
        Code Block
      </MenuButton>
      <MenuButton
        active={hoveredBlock != null && isBlockNodeType(editor, hoveredBlock, "blockquote")}
        onClick={() => handleAction(chain => chain.toggleBlockquote(), "wrapper")}
      >
        <QuotesIcon size={16} weight="bold" />
        Blockquote
      </MenuButton>
    </div>
  );
}
