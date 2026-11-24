// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import {
  autoUpdate,
  flip,
  offset,
  shift,
  useFloating,
} from "@floating-ui/react";
import { type Editor } from "@tiptap/react";
import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import { EditorFloatingPortal } from "../_lib/EditorFloatingPortal";
import { MenuButton } from "../MenuButton";
import { deactivateSlashCommand } from "../SlashCommandExtension";

import { type BlockItem, blockItemLabelKey, blockItems, isPickerItem, runBlockItem } from "./blockItems";
import { blockMenuVariants } from "./variants";

const { menu } = blockMenuVariants();

type BlockMenuContentProps = {
  editor: Editor;
  slashState: { active: boolean; query: string; from: number };
  onPickFiles?: (kind: "image" | "file") => void;
};

export function BlockMenuContent({ editor, slashState, onPickFiles }: BlockMenuContentProps) {
  const { t } = useTranslation();
  const [slashNav, setSlashNav] = useState({ index: 0, query: "" });
  const slashDropdownRef = useRef<HTMLDivElement | null>(null);

  const slashActiveIndex = slashState.query === slashNav.query
    ? slashNav.index
    : 0;

  const filteredItems = useMemo(() => {
    const source = onPickFiles ? blockItems : blockItems.filter(item => !isPickerItem(item));
    if (!slashState.active) return source;
    const q = slashState.query.toLowerCase();
    if (q.length === 0) return source;
    return source.filter(item =>
      [t(blockItemLabelKey(item)), item.label, item.id].some(text => text.toLowerCase().includes(q)),
    );
  }, [onPickFiles, slashState.active, slashState.query, t]);

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
    deactivateSlashCommand(editor.view);
    setSlashNav({ index: 0, query: "" });
  }, [editor]);

  const handleSlashAction = useCallback(
    (item: BlockItem) => {
      if (!slashState.active) return;

      deactivateSlash();
      runBlockItem(editor, item, onPickFiles);
    },
    [editor, slashState, deactivateSlash, onPickFiles],
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
    <EditorFloatingPortal>
      <div
        ref={(node) => {
          slashDropdownRef.current = node;
          slashMenuRefs.setFloating(node);
        }}
        style={slashMenuStyles}
        onMouseDown={e => e.preventDefault()}
        className={menu()}
        data-rich-editor-floating=""
      >
        {filteredItems.length > 0
          ? filteredItems.map((item, index) => (
              <MenuButton
                key={item.id}
                active={index === slashActiveIndex}
                onClick={() => handleSlashAction(item)}
              >
                <item.icon size={16} weight="bold" />
                {t(blockItemLabelKey(item))}
              </MenuButton>
            ))
          : (
              <div className="px-2 py-1.5 text-sm text-txt-tertiary">
                {t("ui.richEditor.slashNoResults")}
              </div>
            )}
      </div>
    </EditorFloatingPortal>
  );
}
