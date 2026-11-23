// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { Editor } from "@tiptap/react";
import { BubbleMenu as BaseBubbleMenu } from "@tiptap/react/menus";
import { type PropsWithChildren, useContext } from "react";
import { tv } from "tailwind-variants/lite";

import { OverlayPortalRootContext } from "../lib/overlayPortalRoot";

const bubbleMenuVariants = tv({
  slots: {
    root: "flex flex-col items-stretch z-20 rounded-lg border border-border-mid bg-level-0 shadow-mid",
    toolbar: "flex items-center gap-1 p-1",
  },
});

const { root, toolbar } = bubbleMenuVariants();

type BubbleMenuProps = {
  editor: Editor;
};

export function BubbleMenu({ editor, children }: PropsWithChildren<BubbleMenuProps>) {
  const portalRoot = useContext(OverlayPortalRootContext);

  return (
    <BaseBubbleMenu
      editor={editor}
      {...(portalRoot ? { appendTo: () => portalRoot } : {})}
      className={root()}
      data-rich-editor-floating=""
      onMouseDown={(e) => {
        if (e.target instanceof HTMLInputElement) {
          return;
        }

        e.preventDefault();
      }}
    >
      {children}
    </BaseBubbleMenu>
  );
}

export function BubbleMenuToolbar({ children }: PropsWithChildren) {
  return <div className={toolbar()}>{children}</div>;
}
