// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import {
  autoUpdate,
  flip,
  offset,
  shift,
  useFloating,
} from "@floating-ui/react";
import type { Icon } from "@phosphor-icons/react";
import { CodeBlockIcon, GridFourIcon, ListBulletsIcon, ListNumbersIcon, MinusIcon, QuotesIcon, TextHOneIcon, TextHThreeIcon, TextHTwoIcon, TextTIcon } from "@phosphor-icons/react";
import { type Editor } from "@tiptap/react";
import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";

import { getSlashStorage } from "../_lib/getSlashStorage";
import { MenuButton } from "../MenuButton";
import { deactivateSlashCommand } from "../SlashCommandExtension";

import { blockMenuVariants } from "./variants";

const { menu } = blockMenuVariants();

type ChainCommands = ReturnType<Editor["chain"]>;

type BlockItem = {
  label: string;
  icon: Icon;
  action: (chain: ChainCommands) => ChainCommands;
};

const BLOCK_ITEMS: BlockItem[] = [
  { label: "Text", icon: TextTIcon, action: chain => chain.setParagraph() },
  { label: "Heading 1", icon: TextHOneIcon, action: chain => chain.toggleHeading({ level: 1 }) },
  { label: "Heading 2", icon: TextHTwoIcon, action: chain => chain.toggleHeading({ level: 2 }) },
  { label: "Heading 3", icon: TextHThreeIcon, action: chain => chain.toggleHeading({ level: 3 }) },
  { label: "Bullet List", icon: ListBulletsIcon, action: chain => chain.toggleBulletList() },
  { label: "Ordered List", icon: ListNumbersIcon, action: chain => chain.toggleOrderedList() },
  { label: "Code Block", icon: CodeBlockIcon, action: chain => chain.toggleCodeBlock() },
  { label: "Blockquote", icon: QuotesIcon, action: chain => chain.toggleBlockquote() },
  { label: "Divider", icon: MinusIcon, action: chain => chain.setHorizontalRule() },
  { label: "Table", icon: GridFourIcon, action: chain => chain.insertTable() },
];

type BlockMenuContentProps = {
  editor: Editor;
  slashState: { active: boolean; query: string; from: number };
};

export function BlockMenuContent({ editor, slashState }: BlockMenuContentProps) {
  const [slashNav, setSlashNav] = useState({ index: 0, query: "" });
  const slashDropdownRef = useRef<HTMLDivElement | null>(null);

  const slashActiveIndex = slashState.query === slashNav.query
    ? slashNav.index
    : 0;

  const filteredItems = useMemo(() => {
    if (!slashState.active) return BLOCK_ITEMS;
    const q = slashState.query.toLowerCase();
    if (q.length === 0) return BLOCK_ITEMS;
    return BLOCK_ITEMS.filter(item => item.label.toLowerCase().includes(q));
  }, [slashState.active, slashState.query]);

  const {
    refs: slashMenuRefs,
    floatingStyles: slashMenuStyles,
  } = useFloating({
    strategy: "fixed",
    placement: "bottom-start",
    middleware: [offset(4), flip(), shift()],
    whileElementsMounted: autoUpdate,
  });

  useLayoutEffect(() => {
    if (!slashState.active) {
      slashMenuRefs.setPositionReference(null);
      return;
    }
    const coords = editor.view.coordsAtPos(slashState.from);
    slashMenuRefs.setPositionReference({
      getBoundingClientRect: () => ({
        x: coords.left,
        y: coords.top,
        top: coords.top,
        left: coords.left,
        bottom: coords.bottom,
        right: coords.left,
        width: 0,
        height: coords.bottom - coords.top,
      }),
    });
  }, [slashState.active, slashState.from, editor, slashMenuRefs]);

  const deactivateSlash = useCallback(() => {
    if (!editor) return;
    const s = getSlashStorage(editor);
    if (s) deactivateSlashCommand(s);
    setSlashNav({ index: 0, query: "" });
  }, [editor]);

  const handleSlashAction = useCallback(
    (item: BlockItem) => {
      if (!slashState.active) return;
      const { from } = slashState;
      const cursorPos = editor.state.selection.from;

      try {
        editor.chain()
          .focus()
          .deleteRange({ from, to: cursorPos })
          .run();

        item.action(editor.chain().focus()).run();
      } catch {
        // Block may no longer be in the document
      }

      deactivateSlash();
    },
    [editor, slashState, deactivateSlash],
  );

  useEffect(() => {
    if (editor.isDestroyed || !slashState.active) return;
    const editorDom = editor.view.dom;

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        e.stopImmediatePropagation();
        setSlashNav(prev => ({
          query: slashState.query,
          index: (prev.query === slashState.query ? prev.index : 0) < filteredItems.length - 1
            ? (prev.query === slashState.query ? prev.index : 0) + 1
            : 0,
        }));
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        e.stopImmediatePropagation();
        setSlashNav(prev => ({
          query: slashState.query,
          index: (prev.query === slashState.query ? prev.index : 0) > 0
            ? (prev.query === slashState.query ? prev.index : 0) - 1
            : filteredItems.length - 1,
        }));
      } else if (e.key === "Enter") {
        e.preventDefault();
        e.stopImmediatePropagation();
        const item = filteredItems[slashActiveIndex];
        if (item) {
          handleSlashAction(item);
        }
      }
    };

    editorDom.addEventListener("keydown", onKeyDown, { capture: true });
    return () => {
      editorDom.removeEventListener("keydown", onKeyDown, { capture: true });
    };
  }, [editor, slashState.active, slashState.query, filteredItems, slashActiveIndex, handleSlashAction]);

  return (
    <div
      ref={(node) => {
        slashDropdownRef.current = node;
        slashMenuRefs.setFloating(node);
      }}
      style={slashMenuStyles}
      onMouseDown={e => e.preventDefault()}
      className={menu()}
    >
      {filteredItems.length > 0
        ? filteredItems.map((item, index) => (
            <MenuButton
              key={item.label}
              active={index === slashActiveIndex}
              onClick={() => handleSlashAction(item)}
            >
              <item.icon size={16} weight="bold" />
              {item.label}
            </MenuButton>
          ))
        : (
            <div className="px-2 py-1.5 text-sm text-txt-tertiary">
              No results
            </div>
          )}
    </div>
  );
}
