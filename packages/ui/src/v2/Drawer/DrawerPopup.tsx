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

import { Drawer as BaseDrawer } from "@base-ui/react/drawer";
import type { ComponentProps } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { drawer } from "./variants";

export type DrawerPopupProps
  = & Omit<ComponentProps<typeof BaseDrawer.Popup>, "className">
    & VariantProps<typeof drawer>
    & {
      className?: string;
    };

// Portal + dimmed backdrop + edge-aligned popup. Children compose the header /
// body / footer regions inside Base UI's Content (swipe-safe text selection).
// `ref` lands on the Popup panel (not the viewport) so nested menus can portal
// into the drawer stacking context.
export function DrawerPopup(props: DrawerPopupProps) {
  const { className, children, side, ref, ...popupProps } = props;
  const slots = drawer({ side });

  return (
    <BaseDrawer.Portal>
      <BaseDrawer.Backdrop className={slots.backdrop()} />
      <BaseDrawer.Viewport className={slots.viewport()}>
        <BaseDrawer.Popup
          ref={ref}
          className={slots.popup({ className })}
          {...popupProps}
        >
          <BaseDrawer.Content className={slots.content()}>
            {children}
          </BaseDrawer.Content>
        </BaseDrawer.Popup>
      </BaseDrawer.Viewport>
    </BaseDrawer.Portal>
  );
}
