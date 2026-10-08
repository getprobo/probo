// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import type { ReactNodeViewProps } from "@tiptap/react";
import { NodeViewWrapper } from "@tiptap/react";

export function ContentUploadNodeView({ node }: ReactNodeViewProps) {
  const fileName = (node.attrs.fileName as string | null) || "File";

  return (
    <NodeViewWrapper>
      <div className="rich-upload" contentEditable={false}>
        {`Uploading ${fileName}…`}
      </div>
    </NodeViewWrapper>
  );
}
