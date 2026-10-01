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

import { acceptAll } from "@probo/helpers";
import { mergeAttributes, Node } from "@tiptap/core";
import type { EditorView } from "@tiptap/pm/view";

import { insertBlockNode } from "./ImageExtension";

export type RichEditorAttachment = {
  href: string;
  fileName: string;
  mimeType: string;
};

export const attachmentAccept = [
  ...new Set(
    Object.entries(acceptAll).flatMap(([mimeType, extensions]) => [mimeType, ...extensions]),
  ),
].join(",");

function fileExtension(name: string) {
  const dot = name.lastIndexOf(".");
  if (dot <= 0) {
    return "";
  }

  return name.slice(dot).toLowerCase();
}

function allowsExtension(extensions: readonly string[], extension: string) {
  return extensions.some(allowed => allowed === extension);
}

export function isAttachmentFile(file: File) {
  const extension = fileExtension(file.name);
  if (!extension) {
    return false;
  }

  if (file.type) {
    if (!(file.type in acceptAll)) {
      return false;
    }

    return allowsExtension(acceptAll[file.type as keyof typeof acceptAll], extension);
  }

  return Object.values(acceptAll).some(extensions => allowsExtension(extensions, extension));
}

export function insertAttachment(
  view: EditorView,
  pos: number,
  attachment: RichEditorAttachment,
) {
  const type = view.state.schema.nodes.attachment;
  if (!type) {
    return;
  }

  insertBlockNode(view, pos, type.create({
    href: attachment.href,
    fileName: attachment.fileName,
    mimeType: attachment.mimeType,
  }));
}

export const AttachmentExtension = Node.create({
  name: "attachment",
  group: "block",
  atom: true,
  draggable: true,
  selectable: true,

  addAttributes() {
    return {
      href: { default: null },
      fileName: {
        default: null,
        parseHTML: element => element.getAttribute("data-file-name") || element.textContent,
        renderHTML: (attributes: { fileName?: unknown }) => {
          if (typeof attributes.fileName !== "string" || attributes.fileName === "") {
            return {};
          }

          return { "data-file-name": attributes.fileName };
        },
      },
      mimeType: {
        default: null,
        parseHTML: element => element.getAttribute("data-mime-type"),
        renderHTML: (attributes: { mimeType?: unknown }) => {
          if (typeof attributes.mimeType !== "string" || attributes.mimeType === "") {
            return {};
          }

          return { "data-mime-type": attributes.mimeType };
        },
      },
    };
  },

  parseHTML() {
    return [{ tag: "a.attachment[href]" }];
  },

  renderHTML({ node, HTMLAttributes }) {
    const fileName = typeof node.attrs.fileName === "string" && node.attrs.fileName
      ? node.attrs.fileName
      : "Attachment";
    const href = typeof node.attrs.href === "string" ? node.attrs.href : "";

    return [
      "a",
      mergeAttributes(HTMLAttributes, {
        class: "attachment",
        href,
        target: "_blank",
        rel: "noopener noreferrer",
      }),
      fileName,
    ];
  },
});
