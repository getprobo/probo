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
import type { ComponentProps } from "react";
import { useState } from "react";

import { OverlayPortalRootContext } from "../../lib/overlayPortalRoot";

import { dialog } from "./variants";

export type DialogPopupProps
  = & Omit<ComponentProps<typeof BaseDialog.Popup>, "className">
    & {
      className?: string;
      // When true, the popup itself does not scroll; inner regions manage overflow.
      lockScroll?: boolean;
    };

// Portal + dimmed backdrop + centered, styled popup frame. Children compose the
// header / body / footer regions.
export function DialogPopup(props: DialogPopupProps) {
  const { className, children, lockScroll = false, ...popupProps } = props;
  const { backdrop, popup, overlayRoot: overlayRootSlot } = dialog({ lockScroll });
  const [overlayRoot, setOverlayRoot] = useState<HTMLElement | null>(null);

  return (
    <BaseDialog.Portal>
      <OverlayPortalRootContext.Provider value={overlayRoot}>
        <BaseDialog.Backdrop className={backdrop()} />
        <BaseDialog.Popup className={popup({ className })} {...popupProps}>
          {children}
        </BaseDialog.Popup>
        <div ref={setOverlayRoot} className={overlayRootSlot()} />
      </OverlayPortalRootContext.Provider>
    </BaseDialog.Portal>
  );
}
