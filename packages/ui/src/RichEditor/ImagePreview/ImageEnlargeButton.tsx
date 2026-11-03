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
import { type ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { imagePreview } from "./variants";

export const richImageEnlargeSelector = "[data-rich-image-enlarge]";

interface ImageEnlargeButtonProps {
  src: string;
  alt: string;
  children: ReactNode;
}

export function ImageEnlargeButton({ src, alt, children }: ImageEnlargeButtonProps) {
  const { t } = useTranslation();
  const { trigger, backdrop, popup, title, image, close } = imagePreview();
  const label = alt !== ""
    ? t("ui.richEditor.enlargeImageNamed", { name: alt })
    : t("ui.richEditor.enlargeImage");

  return (
    <BaseDialog.Root>
      <BaseDialog.Trigger
        className={trigger()}
        data-rich-image-enlarge=""
        aria-label={label}
        onMouseDown={event => event.stopPropagation()}
      >
        {children}
      </BaseDialog.Trigger>
      <BaseDialog.Portal>
        <BaseDialog.Backdrop className={backdrop()} />
        <BaseDialog.Popup className={popup()}>
          <BaseDialog.Title className={title()}>{label}</BaseDialog.Title>
          <img className={image()} src={src} alt={alt} />
          <BaseDialog.Close className={close()} aria-label={t("ui.richEditor.closeImage")}>
            <XIcon aria-hidden weight="bold" />
          </BaseDialog.Close>
        </BaseDialog.Popup>
      </BaseDialog.Portal>
    </BaseDialog.Root>
  );
}
