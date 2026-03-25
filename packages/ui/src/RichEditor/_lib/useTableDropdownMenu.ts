// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import {
  autoUpdate,
  flip,
  offset,
  shift,
  useDismiss,
  useFloating,
  useFloatingRootContext,
  useInteractions,
} from "@floating-ui/react";
import { useState } from "react";

export function useTableDropdownMenu() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [triggerEl, setTriggerEl] = useState<Element | null>(null);
  const [dropdownEl, setDropdownEl] = useState<HTMLElement | null>(null);

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

  const dismiss = useDismiss(menuRootContext);
  const { getFloatingProps } = useInteractions([dismiss]);

  return {
    menuOpen,
    setMenuOpen,
    triggerEl,
    setTriggerEl,
    dropdownEl,
    setDropdownEl,
    menuRefs,
    menuStyles,
    getFloatingProps,
  };
}
