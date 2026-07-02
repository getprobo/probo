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

import { SpinnerGapIcon } from "@phosphor-icons/react";
import type { ComponentProps } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { iconButton } from "./variants";

export type IconButtonProps
  = Omit<ComponentProps<"button">, "color">
    & VariantProps<typeof iconButton>
    & {
      // Shows a spinner in place of the icon and disables the button.
      loading?: boolean;
    };

// Square, icon-only action (Radix "IconButton"). The single icon is passed as
// children; always provide an `aria-label`. See contrib/claude/ui.md.
export function IconButton(props: IconButtonProps) {
  const {
    size, variant, color, highContrast, className,
    loading = false, disabled, type = "button", children, ...rest
  } = props;

  return (
    <button
      type={type}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      className={iconButton({ size, variant, color, highContrast, className })}
      {...rest}
    >
      {loading ? <SpinnerGapIcon className="animate-spin" aria-hidden /> : children}
    </button>
  );
}
