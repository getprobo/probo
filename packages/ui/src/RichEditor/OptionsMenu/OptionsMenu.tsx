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
import { type Editor } from "@tiptap/react";
import { useState } from "react";

import { useHoveredBlock } from "../_lib/useHoveredBlock";

import { OptionsMenuContent } from "./OptionsMenuContent";
import { OptionsMenuTrigger } from "./OptionsMenuTrigger";

export type OptionsMenuFloating = {
  setTriggerEl: React.Dispatch<React.SetStateAction<Element | null>>;
  setDropdownEl: React.Dispatch<React.SetStateAction<HTMLElement | null>>;
  menuRefs: ReturnType<typeof useFloating>["refs"];
  menuStyles: ReturnType<typeof useFloating>["floatingStyles"];
  getReferenceProps: ReturnType<typeof useInteractions>["getReferenceProps"];
  getFloatingProps: ReturnType<typeof useInteractions>["getFloatingProps"];
};

type OptionsMenuProps = {
  editor: Editor;
};

export function OptionsMenu({ editor }: OptionsMenuProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [triggerEl, setTriggerEl] = useState<Element | null>(null);
  const [dropdownEl, setDropdownEl] = useState<HTMLElement | null>(null);

  const { hoveredBlock, setHoveredBlock } = useHoveredBlock(editor, menuOpen);

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

  return (
    <>
      <OptionsMenuTrigger
        editor={editor}
        hoveredBlock={hoveredBlock}
        setHoveredBlock={setHoveredBlock}
        setTriggerEl={setTriggerEl}
        menuRefs={menuRefs}
        getReferenceProps={getReferenceProps}
      />
      {menuOpen && (
        <OptionsMenuContent
          editor={editor}
          hoveredBlock={hoveredBlock}
          setMenuOpen={setMenuOpen}
          setDropdownEl={setDropdownEl}
          menuRefs={menuRefs}
          menuStyles={menuStyles}
          getFloatingProps={getFloatingProps}
        />
      )}
    </>
  );
}
