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

import { insertPicture, isPictureFile, type RichEditorPicture } from "./ImageExtension";

export type RichEditorPictureUpload = (file: File) => Promise<RichEditorPicture>;

type PictureUploadStorage = {
  upload?: RichEditorPictureUpload;
};

const pictureUploadKey = new PluginKey("pictureUpload");

export function setPictureUpload(
  editor: Editor,
  upload: RichEditorPictureUpload | undefined,
) {
  const storage = (editor.storage as { pictureUpload?: PictureUploadStorage }).pictureUpload;
  if (!storage) {
    return;
  }

  storage.upload = upload;
}

function pictureFiles(list: FileList | null | undefined) {
  if (!list) {
    return [];
  }

  return [...list].filter(isPictureFile);
}

async function uploadPictures(
  view: EditorView,
  files: File[],
  pos: number,
  upload: RichEditorPictureUpload,
) {
  let at = pos;

  for (const file of files) {
    if (view.isDestroyed) {
      return;
    }

    const picture = await upload(file);
    if (view.isDestroyed) {
      return;
    }

    insertPicture(view, at, {
      src: picture.src,
      alt: picture.alt ?? file.name,
    });
    at = view.state.selection.to;
  }
}

export const PictureUploadExtension = Extension.create<object, PictureUploadStorage>({
  name: "pictureUpload",

  addStorage() {
    return {};
  },

  addProseMirrorPlugins() {
    const storage = this.storage;

    return [
      new Plugin({
        key: pictureUploadKey,
        props: {
          handlePaste(view, event) {
            const upload = storage.upload;
            const files = pictureFiles(event.clipboardData?.files);
            if (!upload || files.length === 0) {
              return false;
            }

            event.preventDefault();
            void uploadPictures(view, files, view.state.selection.from, upload).catch(() => {
              // The upload callback reports the failure.
            });
            return true;
          },

          handleDrop(view, event) {
            const upload = storage.upload;
            const files = pictureFiles(event.dataTransfer?.files);
            if (!upload || files.length === 0) {
              return false;
            }

            const coords = view.posAtCoords({
              left: event.clientX,
              top: event.clientY,
            });
            if (!coords) {
              return false;
            }

            event.preventDefault();
            void uploadPictures(view, files, coords.pos, upload).catch(() => {
              // The upload callback reports the failure.
            });
            return true;
          },
        },
      }),
    ];
  },
});
