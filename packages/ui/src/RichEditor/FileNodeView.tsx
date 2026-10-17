// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { PaperclipIcon } from "@phosphor-icons/react";
import { fileSize } from "@probo/i18n";
import type { ReactNodeViewProps } from "@tiptap/react";
import { NodeViewWrapper } from "@tiptap/react";
import { useTranslation } from "react-i18next";

import { attachmentPath, useAttachmentParentID } from "./attachments";

export function FileNodeView({ node }: ReactNodeViewProps) {
  const { t } = useTranslation();
  const parentId = useAttachmentParentID();
  const fileId = typeof node.attrs.fileId === "string" ? node.attrs.fileId : "";
  const fileName
    = typeof node.attrs.fileName === "string" && node.attrs.fileName !== ""
      ? node.attrs.fileName
      : t("ui.richEditor.file");
  const size = Number(node.attrs.size ?? 0);
  const href = fileId ? attachmentPath(fileId, parentId) : undefined;

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
        {size > 0 && (
          <span className="rich-file-size">{fileSize(size, t)}</span>
        )}
      </a>
    </NodeViewWrapper>
  );
}
