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

import type { DialogTriggerProps } from "../Dialog/DialogTrigger";
import { DialogTrigger } from "../Dialog/DialogTrigger";

import { imagePreview } from "./variants";

// Opens the lightbox. The child is usually the small image. Pass `render` to
// use an existing control (e.g. a Button) as the trigger, per Base UI's
// polymorphism — do not clone children by hand.
export type ImagePreviewTriggerProps = Omit<DialogTriggerProps, "className"> & {
  className?: string;
};

export function ImagePreviewTrigger(props: ImagePreviewTriggerProps) {
  const { className, render, ...rest } = props;
  const { trigger } = imagePreview();

  return (
    <DialogTrigger
      className={render == null ? trigger({ className }) : className}
      render={render}
      data-image-preview-trigger=""
      {...rest}
    />
  );
}
