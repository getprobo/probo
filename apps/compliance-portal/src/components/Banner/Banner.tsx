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

import { XIcon } from "@phosphor-icons/react";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import type { ReactNode } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { banner } from "./variants";

export type BannerProps = {
  color: NonNullable<VariantProps<typeof banner>["color"]>;
  icon: ReactNode;
  message: ReactNode;
  actions: ReactNode;
  dismissLabel: string;
  onDismiss: () => void;
  dismissDisabled?: boolean;
};

// Presentational full-bleed portal banner: tinted band, TopBar-aligned column,
// leading icon + message, actions, and dual dismiss (mobile in the message row,
// desktop beside actions).
export function Banner({
  color,
  icon,
  message,
  actions,
  dismissLabel,
  onDismiss,
  dismissDisabled = false,
}: BannerProps) {
  const slots = banner({ color });

  return (
    <aside className={slots.root()} role="status">
      <div className={slots.inner()}>
        <div className={slots.content()}>
          <span aria-hidden className={slots.icon()}>
            {icon}
          </span>
          <div className={slots.message()}>{message}</div>
          <IconButton
            size={1}
            variant="soft"
            color={color}
            aria-label={dismissLabel}
            disabled={dismissDisabled}
            className={slots.dismissMobile()}
            onClick={onDismiss}
          >
            <XIcon />
          </IconButton>
        </div>
        <div className={slots.actions()}>
          {actions}
          <IconButton
            size={1}
            variant="soft"
            color={color}
            aria-label={dismissLabel}
            disabled={dismissDisabled}
            className={slots.dismissDesktop()}
            onClick={onDismiss}
          >
            <XIcon />
          </IconButton>
        </div>
      </div>
    </aside>
  );
}
