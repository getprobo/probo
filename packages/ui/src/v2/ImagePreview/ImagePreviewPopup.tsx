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

import { Dialog as BaseDialog } from "@base-ui/react/dialog";
import { XIcon } from "@phosphor-icons/react";

import { DialogClose } from "../Dialog/DialogClose";

import { imagePreview } from "./variants";

export interface ImagePreviewPopupProps {
  src: string;
  alt: string;
  // Accessible name of the dialog. The image keeps `alt` separately.
  label: string;
  closeLabel: string;
}

// Dimmed backdrop, the image at viewport size, and a close control. Escape and
// a click on the backdrop also dismiss it, via the dialog root.
export function ImagePreviewPopup(props: ImagePreviewPopupProps) {
  const { src, alt, label, closeLabel } = props;
  const { backdrop, popup, title, image, close } = imagePreview();

  return (
    <BaseDialog.Portal>
      <BaseDialog.Backdrop className={backdrop()} />
      <BaseDialog.Popup className={popup()}>
        <BaseDialog.Title className={title()}>{label}</BaseDialog.Title>
        <img className={image()} src={src} alt={alt} />
        <DialogClose className={close()} aria-label={closeLabel}>
          <XIcon aria-hidden weight="bold" />
        </DialogClose>
      </BaseDialog.Popup>
    </BaseDialog.Portal>
  );
}
