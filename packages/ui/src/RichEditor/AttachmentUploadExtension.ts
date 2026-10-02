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

import { type Editor, Extension } from "@tiptap/core";
import { Plugin, PluginKey } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";

import { insertAttachment, isAttachmentFile, type RichEditorAttachment } from "./AttachmentExtension";
import { insertPicture, isPictureMime } from "./ImageExtension";

export type RichEditorUploadedFile = {
  src: string;
  fileName: string;
  mimeType: string;
};

export type RichEditorAttachmentUpload = (file: File) => Promise<RichEditorUploadedFile>;

type AttachmentUploadStorage = {
  upload?: RichEditorAttachmentUpload;
};

const attachmentUploadKey = new PluginKey("attachmentUpload");

export function setAttachmentUpload(
  editor: Editor,
  upload: RichEditorAttachmentUpload | undefined,
) {
  const storage = (editor.storage as { attachmentUpload?: AttachmentUploadStorage }).attachmentUpload;
  if (!storage) {
    return;
  }

  storage.upload = upload;
}

function attachmentFiles(list: FileList | null | undefined) {
  if (!list) {
    return [];
  }

  return [...list].filter(isAttachmentFile);
}

function insertUploadedFile(
  view: EditorView,
  pos: number,
  uploaded: RichEditorUploadedFile,
) {
  if (isPictureMime(uploaded.mimeType)) {
    insertPicture(view, pos, {
      src: uploaded.src,
      alt: uploaded.fileName,
    });
    return;
  }

  const attachment: RichEditorAttachment = {
    href: uploaded.src,
    fileName: uploaded.fileName,
    mimeType: uploaded.mimeType,
  };
  insertAttachment(view, pos, attachment);
}

async function uploadAttachments(
  view: EditorView,
  files: File[],
  pos: number,
  upload: RichEditorAttachmentUpload,
) {
  let at = pos;

  for (const file of files) {
    if (view.isDestroyed) {
      return;
    }

    try {
      const uploaded = await upload(file);
      if (view.isDestroyed) {
        return;
      }

      insertUploadedFile(view, at, uploaded);
      at = view.state.selection.to;
    } catch {
      // The upload callback reports the failure.
    }
  }
}

export function filesFromDataTransfer(data: DataTransfer | null | undefined) {
  if (!data) {
    return [];
  }

  // Read the file items before anything else touches the payload. Chrome
  // drops the file list once text or a URI is read, and then opens the file.
  const files: File[] = [];
  for (const item of data.items) {
    if (item.kind !== "file") {
      continue;
    }

    const file = item.getAsFile();
    if (file) {
      files.push(file);
    }
  }

  if (files.length === 0) {
    files.push(...data.files);
  }

  return files;
}

export function uploadEditorFiles(
  view: EditorView,
  fileList: FileList | File[] | null | undefined,
  pos: number,
  upload: RichEditorAttachmentUpload,
) {
  const files = fileList ? [...fileList] : [];
  if (files.length === 0) {
    return false;
  }

  void uploadAttachments(view, files, pos, upload);
  return true;
}

function dropPosition(view: EditorView, event: DragEvent) {
  return view.posAtCoords({
    left: event.clientX,
    top: event.clientY,
  })?.pos ?? view.state.selection.from;
}

export const AttachmentUploadExtension = Extension.create<object, AttachmentUploadStorage>({
  name: "attachmentUpload",

  addStorage() {
    return {};
  },

  addProseMirrorPlugins() {
    const storage = this.storage;

    return [
      new Plugin({
        key: attachmentUploadKey,
        props: {
          // ProseMirror's handleDrop and handlePaste read dataTransfer text
          // before those props run. Chrome clears the file list when that
          // happens, so file drops have to be taken from the DOM event.
          handleDOMEvents: {
            drop(view, event) {
              const upload = storage.upload;
              const files = filesFromDataTransfer(event.dataTransfer);
              const fileDrag = files.length > 0
                || (event.dataTransfer?.types.includes("Files") ?? false);
              if (!upload || !fileDrag) {
                return false;
              }

              // Claim the drop before the browser opens the file in the tab.
              event.preventDefault();
              event.stopPropagation();
              if (files.length === 0) {
                return true;
              }

              void uploadAttachments(view, files, dropPosition(view, event), upload);
              return true;
            },

            paste(view, event) {
              const upload = storage.upload;
              const files = attachmentFiles(event.clipboardData?.files);
              if (!upload || files.length === 0) {
                return false;
              }

              event.preventDefault();
              event.stopPropagation();
              void uploadAttachments(view, files, view.state.selection.from, upload);
              return true;
            },
          },
        },
      }),
    ];
  },
});

export { insertUploadedFile };
