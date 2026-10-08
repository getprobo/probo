// Copyright (c) 2026 Probo Inc <hello@probo.com>.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import { Fragment, type Node as PmNode } from "@tiptap/pm/model";
import type { EditorView } from "@tiptap/pm/view";

import { type AttachmentKind, attachmentKind, prepareAttachment, rememberAttachmentPreview } from "./attachments";

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
  onError: (message: string) => void,
) {
  if (view.isDestroyed || !view.editable) {
    return;
  }

  const accepted = files.flatMap((raw) => {
    const file = prepareAttachment(raw);
    const kind = attachmentKind(file);
    if (!kind) {
      onError(`${file.name} is not a supported file`);
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
      rememberAttachmentPreview(uploaded.id, item.file);
      replaceUpload(view, item.id, item.kind, uploaded);
    } catch {
      onError(`Could not upload ${item.file.name}`);
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

  const attrs = kind === "image"
    ? { fileId: uploaded.id, src: null, alt: uploaded.fileName, title: null }
    : {
        fileId: uploaded.id,
        fileName: uploaded.fileName,
        mimeType: uploaded.mimeType,
        size: uploaded.size,
      };
  const node = view.state.schema.nodes[kind].create(attrs);
  const found = findUpload(view.state.doc, id);

  if (!found) {
    view.dispatch(view.state.tr.insert(view.state.doc.content.size, node));
    return;
  }

  view.dispatch(view.state.tr.replaceWith(found.pos, found.pos + found.node.nodeSize, node));
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

function blockInsertPos(doc: PmNode, pos: number) {
  const safe = Math.max(0, Math.min(pos, doc.content.size));
  const $pos = doc.resolve(safe);
  if ($pos.depth === 0) {
    return safe;
  }

  return $pos.after(1);
}
