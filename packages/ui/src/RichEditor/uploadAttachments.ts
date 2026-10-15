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

import { Fragment, type Node as PmNode } from "@tiptap/pm/model";
import type { EditorView } from "@tiptap/pm/view";

import {
  type AttachmentKind,
  attachmentKind,
  maxAttachmentSize,
  prepareAttachment,
  rememberImagePreview,
} from "./attachments";

export type AttachmentUploadError = {
  fileName: string;
  reason: "unsupported" | "tooLarge" | "failed";
};

export type UploadedAttachment = {
  id: string;
  fileName: string;
  mimeType: string;
  size: number;
};

export type UploadAttachment = (file: File) => Promise<UploadedAttachment>;

export async function uploadAttachments(
  view: EditorView,
  files: File[],
  pos: number,
  upload: UploadAttachment,
  onError: (error: AttachmentUploadError) => void,
) {
  if (view.isDestroyed || !view.editable) {
    return;
  }

  const accepted = files.flatMap((raw) => {
    const file = prepareAttachment(raw);
    if (file.size > maxAttachmentSize) {
      onError({ fileName: file.name, reason: "tooLarge" });
      return [];
    }

    const kind = attachmentKind(file);
    if (!kind) {
      onError({ fileName: file.name, reason: "unsupported" });
      return [];
    }

    return [{ file, kind, id: crypto.randomUUID() }];
  });

  if (accepted.length === 0) {
    return;
  }

  const insertAt = blockInsertPos(view.state.doc, pos);
  const placeholders = accepted.map(item => view.state.schema.nodes.contentUpload.create({
    id: item.id,
    fileName: item.file.name,
  }));

  view.dispatch(view.state.tr.insert(insertAt, Fragment.fromArray(placeholders)));

  await Promise.all(accepted.map(async (item) => {
    try {
      const uploaded = await upload(item.file);
      if (view.isDestroyed || !findUpload(view.state.doc, item.id)) {
        return;
      }

      if (item.kind === "image") {
        rememberImagePreview(view, uploaded.id, item.file);
      }

      replaceUpload(view, item.id, item.kind, uploaded);
    } catch {
      onError({ fileName: item.file.name, reason: "failed" });
      removeUpload(view, item.id);
    }
  }));
}

function replaceUpload(
  view: EditorView,
  id: string,
  kind: AttachmentKind,
  uploaded: UploadedAttachment,
) {
  if (view.isDestroyed) {
    return;
  }

  const found = findUpload(view.state.doc, id);
  if (!found) {
    return;
  }

  const attrs = kind === "image"
    ? { fileId: uploaded.id, src: null, alt: uploaded.fileName, title: null }
    : {
        fileId: uploaded.id,
        fileName: uploaded.fileName,
        mimeType: uploaded.mimeType,
        size: uploaded.size,
      };
  const node = view.state.schema.nodes[kind].create(attrs);

  const tr = view.state.tr
    .replaceWith(found.pos, found.pos + found.node.nodeSize, node)
    .setMeta("addToHistory", false);
  view.dispatch(tr);
}

function removeUpload(view: EditorView, id: string) {
  if (view.isDestroyed) {
    return;
  }

  const found = findUpload(view.state.doc, id);
  if (!found) {
    return;
  }

  view.dispatch(view.state.tr.delete(found.pos, found.pos + found.node.nodeSize));
}

function findUpload(doc: PmNode, id: string) {
  const found: { value: { pos: number; node: PmNode } | null } = { value: null };

  doc.descendants((node, pos) => {
    if (node.type.name === "contentUpload" && node.attrs.id === id) {
      found.value = { pos, node };
      return false;
    }

    return true;
  });

  return found.value;
}

// Insert at the caret. ProseMirror splits the surrounding block, so a file
// dropped in a list item or quote stays next to that caret.
function blockInsertPos(doc: PmNode, pos: number) {
  return Math.max(0, Math.min(pos, doc.content.size));
}
