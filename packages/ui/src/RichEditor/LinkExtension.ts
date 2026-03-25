// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { Link } from "@tiptap/extension-link";
import { Plugin, PluginKey } from "@tiptap/pm/state";

export const LinkExtension = Link.extend({
  addProseMirrorPlugins: () => {
    return [
      new Plugin({
        key: new PluginKey("handleControlClick"),
        props: {
          handleKeyDown: (view, event) => {
            if (event.key === "Control" || event.key === "Meta") {
              view.dom.classList.add("pointer-on-hovered-link");
            }
          },
          handleDOMEvents: {
            keyup: (view, event) => {
              if (event.key === "Control" || event.key === "Meta") {
                view.dom.classList.remove("pointer-on-hovered-link");
              }
            },
          },
          handleClick: (_view, _, event) => {
            const { ctrlKey, metaKey } = event; // Check for Ctrl (Windows) or Cmd (Mac)
            const keyPressed = ctrlKey || metaKey;

            if (keyPressed) {
              const link = (event.target as Element | null)?.closest("a");

              if (link?.href) {
                window.open(link.href, "_blank", "noopener,noreferrer");
                return true;
              }
            }
            return false; // Let other handlers run
          },
        },
      }),
    ];
  },
}).configure({
  openOnClick: false,
});
