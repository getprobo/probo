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

import {
  autoUpdate,
  flip,
  offset,
  shift,
  useFloating,
} from "@floating-ui/react";
import { type Editor } from "@tiptap/react";
import { useCallback, useLayoutEffect, useState, useSyncExternalStore } from "react";
import { useTranslation } from "react-i18next";

import { EditorFloatingPortal } from "./_lib/EditorFloatingPortal";
import { blockMenuVariants } from "./BlockMenu/variants";
import { getMentionStorage, type MentionSuggestionState } from "./MentionExtension";
import { MenuButton } from "./MenuButton";

const { menu } = blockMenuVariants();

function mentionOptionId(index: number): string {
  return `mention-option-${index}`;
}

const inactiveState: MentionSuggestionState = {
  active: false,
  loading: false,
  query: "",
  from: 0,
  items: [],
  select: null,
};

function useMentionState(editor: Editor): MentionSuggestionState {
  const subscribe = useCallback(
    (listener: () => void) => {
      const storage = getMentionStorage(editor);
      storage?.listeners.add(listener);

      return () => {
        storage?.listeners.delete(listener);
      };
    },
    [editor],
  );

  return useSyncExternalStore(
    subscribe,
    () => getMentionStorage(editor)?.state ?? inactiveState,
  );
}

type MentionMenuProps = {
  editor: Editor;
};

export function MentionMenu({ editor }: MentionMenuProps) {
  const { t } = useTranslation();
  const state = useMentionState(editor);
  const [nav, setNav] = useState({ index: 0, items: state.items });
  const activeIndex = nav.items === state.items ? nav.index : 0;

  const { refs, floatingStyles } = useFloating({
    strategy: "fixed",
    placement: "bottom-start",
    middleware: [offset(4), flip(), shift()],
    whileElementsMounted: autoUpdate,
  });

  useLayoutEffect(() => {
    if (!state.active || editor.isDestroyed) {
      refs.setPositionReference(null);
      return;
    }

    const coords = editor.view.coordsAtPos(state.from);
    refs.setPositionReference({
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
  }, [state.active, state.from, editor, refs]);

  useLayoutEffect(() => {
    const storage = getMentionStorage(editor);
    if (!storage) {
      return;
    }

    storage.keyDown = (event) => {
      const { items, select } = state;
      if (items.length === 0) {
        return false;
      }

      switch (event.key) {
        case "ArrowDown":
          setNav({ items, index: (activeIndex + 1) % items.length });
          return true;
        case "ArrowUp":
          setNav({ items, index: (activeIndex - 1 + items.length) % items.length });
          return true;
        case "Enter":
        case "Tab": {
          const item = items[activeIndex];
          if (item && select) {
            select(item);
          }
          return true;
        }
        default:
          return false;
      }
    };

    return () => {
      storage.keyDown = null;
    };
  }, [editor, state, activeIndex]);

  if (!state.active) {
    return null;
  }

  return (
    <EditorFloatingPortal>
      <div
        ref={(node) => {
          refs.setFloating(node);
        }}
        style={floatingStyles}
        onMouseDown={e => e.preventDefault()}
        className={menu({ className: "w-72" })}
        data-rich-editor-floating=""
        role="listbox"
        aria-label={t("ui.richEditor.mentionMembers")}
        aria-activedescendant={state.items.length > 0 ? mentionOptionId(activeIndex) : undefined}
      >
        {state.items.length > 0
          ? state.items.map((item, index) => (
              <MenuButton
                key={item.id}
                id={mentionOptionId(index)}
                role="option"
                aria-selected={index === activeIndex}
                active={index === activeIndex}
                onClick={() => state.select?.(item)}
              >
                <span className="flex min-w-0 flex-col items-start text-left">
                  <span className="truncate font-medium text-txt-primary">{item.label}</span>
                  {item.description && (
                    <span className="truncate text-xs text-txt-tertiary">{item.description}</span>
                  )}
                </span>
              </MenuButton>
            ))
          : (
              <div className="px-2 py-1.5 text-sm text-txt-tertiary">
                {state.loading
                  ? t("ui.richEditor.mentionLoading")
                  : t("ui.richEditor.mentionNoResults")}
              </div>
            )}
      </div>
    </EditorFloatingPortal>
  );
}
