// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
