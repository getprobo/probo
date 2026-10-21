// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { type Editor, useEditorState } from "@tiptap/react";

import { getSlashStorage } from "../_lib/getSlashStorage";
import { useHoveredBlock } from "../_lib/useHoveredBlock";

import { BlockMenuContent } from "./BlockMenuContent";
import { BlockMenuTrigger } from "./BlockMenuTrigger";

type BlockMenuProps = {
  editor: Editor;
  compact?: boolean;
  // When false, the plus handle is not mounted. The slash menu still opens.
  handles?: boolean;
  onPickFiles?: (kind: "image" | "file") => void;
};

export function BlockMenu({
  editor,
  compact = false,
  handles = true,
  onPickFiles,
}: BlockMenuProps) {
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

  return (
    <>
      {handles && (
        <BlockMenuHandle
          editor={editor}
          compact={compact}
          slashActive={slashState.active}
        />
      )}
      {slashState.active && (
        <BlockMenuContent
          editor={editor}
          slashState={slashState}
          onPickFiles={onPickFiles}
        />
      )}
    </>
  );
}

function BlockMenuHandle({
  editor,
  compact,
  slashActive,
}: {
  editor: Editor;
  compact: boolean;
  slashActive: boolean;
}) {
  const { hoveredBlock } = useHoveredBlock(editor, slashActive);

  if (hoveredBlock == null) {
    return null;
  }

  return (
    <BlockMenuTrigger
      editor={editor}
      hoveredBlock={hoveredBlock}
      compact={compact}
    />
  );
}
