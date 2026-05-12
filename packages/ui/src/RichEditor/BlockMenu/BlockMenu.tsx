// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { type Editor, useEditorState } from "@tiptap/react";

import { getSlashStorage } from "../_lib/getSlashStorage";
import { useHoveredBlock } from "../_lib/useHoveredBlock";

import { BlockMenuContent } from "./BlockMenuContent";
import { BlockMenuTrigger } from "./BlockMenuTrigger";

type BlockMenuProps = {
  editor: Editor;
};

export function BlockMenu({ editor }: BlockMenuProps) {
  const slashState = useEditorState({
    editor,
    selector: ({ editor: e }) => {
      const s = getSlashStorage(e);
      return {
        active: s?.active ?? false,
        query: s?.query ?? "",
        from: s?.from ?? 0,
      };
    },
  });

  const { hoveredBlock } = useHoveredBlock(editor, slashState.active);

  return (
    <>
      {hoveredBlock != null && (
        <BlockMenuTrigger
          editor={editor}
          hoveredBlock={hoveredBlock}
        />
      )}
      {slashState.active && (
        <BlockMenuContent
          editor={editor}
          slashState={slashState}
        />
      )}
    </>
  );
}
