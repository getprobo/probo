// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { PaperclipIcon } from "@phosphor-icons/react";
import type { ReactNodeViewProps } from "@tiptap/react";
import { NodeViewWrapper } from "@tiptap/react";

import { attachmentPath, attachmentPreviewURL, formatFileSize } from "./attachments";

export function FileNodeView({ node }: ReactNodeViewProps) {
  const fileId = node.attrs.fileId as string | null;
  const fileName = (node.attrs.fileName as string | null) || "File";
  const size = Number(node.attrs.size ?? 0);
  const href = fileId ? attachmentPreviewURL(fileId) ?? attachmentPath(fileId) : undefined;

  return (
    <NodeViewWrapper>
      <a
        className="rich-file"
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        contentEditable={false}
      >
        <PaperclipIcon size={16} weight="bold" />
        <span className="rich-file-name">{fileName}</span>
        {size > 0 && <span className="rich-file-size">{formatFileSize(size)}</span>}
      </a>
    </NodeViewWrapper>
  );
}
