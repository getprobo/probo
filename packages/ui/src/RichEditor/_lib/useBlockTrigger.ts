// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { autoUpdate, offset, useFloating } from "@floating-ui/react";
import { useLayoutEffect } from "react";

const TRIGGER_HEIGHT = 24;

export function useBlockTrigger(hoveredBlock: HTMLElement | null, offsetValue: number) {
  const blockHeight = hoveredBlock?.getBoundingClientRect().height ?? 0;
  const triggerPlacement = blockHeight > 2 * TRIGGER_HEIGHT ? "left-start" as const : "left" as const;

  const {
    refs: triggerRefs,
    floatingStyles: triggerStyles,
    isPositioned,
  } = useFloating({
    strategy: "fixed",
    placement: triggerPlacement,
    middleware: [offset(offsetValue)],
    whileElementsMounted: autoUpdate,
  });

  useLayoutEffect(() => {
    triggerRefs.setReference(hoveredBlock);
  }, [hoveredBlock, triggerRefs]);

  return { triggerRefs, triggerStyles, isPositioned };
}
