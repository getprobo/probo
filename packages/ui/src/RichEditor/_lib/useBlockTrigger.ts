// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { autoUpdate, offset, useFloating } from "@floating-ui/react";
import { useLayoutEffect } from "react";

const TRIGGER_HEIGHT = 24;

export function useBlockTrigger(hoveredBlock: HTMLElement | null, offsetValue: number) {
  const blockHeight = hoveredBlock?.getBoundingClientRect().height ?? 0;
  // Center on a single line. Only pin to the top for tall blocks (headings,
  // lists) so the control does not sit above the first line of text.
  const triggerPlacement = blockHeight > 4 * TRIGGER_HEIGHT ? "left-start" as const : "left" as const;

  const {
    refs: triggerRefs,
    floatingStyles: triggerStyles,
    isPositioned,
  } = useFloating({
    strategy: "absolute",
    placement: triggerPlacement,
    middleware: [offset(offsetValue)],
    whileElementsMounted: (reference, floating, update) =>
      autoUpdate(reference, floating, update, { animationFrame: true }),
  });

  useLayoutEffect(() => {
    triggerRefs.setReference(hoveredBlock);
  }, [hoveredBlock, triggerRefs]);

  return { triggerRefs, triggerStyles, isPositioned };
}
