// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import type { ReactNodeViewProps } from "@tiptap/react";
import { NodeViewWrapper } from "@tiptap/react";

import { attachmentPath, attachmentPreviewURL } from "./attachments";

export function ImageNodeView({ node }: ReactNodeViewProps) {
  const fileId = node.attrs.fileId as string | null;
  const src = fileId
    ? attachmentPreviewURL(fileId) ?? attachmentPath(fileId)
    : node.attrs.src as string | null;
  if (!src) {
    return null;
  }

  const alt = (node.attrs.alt as string | null) ?? "";
  const title = (node.attrs.title as string | null) ?? undefined;

  return (
    <NodeViewWrapper>
      <img className="rich-image" src={src} alt={alt} title={title} />
    </NodeViewWrapper>
  );
}
