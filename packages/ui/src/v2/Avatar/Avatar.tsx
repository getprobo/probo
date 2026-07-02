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

import { Avatar as BaseAvatar } from "@base-ui/react/avatar";
import type { ComponentProps, ReactNode } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { avatar } from "./variants";

export type AvatarProps
  = Omit<ComponentProps<typeof BaseAvatar.Root>, "color" | "className">
    & VariantProps<typeof avatar>
    & {
      className?: string;
      // Profile image URL. When absent or it fails to load, `fallback` shows.
      src?: string;
      alt?: string;
      // Shown until/unless the image loads: user initials (string) or an icon.
      fallback: ReactNode;
    };

// Foundational avatar primitive (Radix "Avatar") over Base UI's image-loading
// behavior. See contrib/claude/ui.md.
export function Avatar(props: AvatarProps) {
  const { size, variant, color, highContrast, radius, className, src, alt, fallback, ...rest } = props;
  const slots = avatar({ size, variant, color, highContrast, radius });

  return (
    <BaseAvatar.Root className={slots.root({ className })} {...rest}>
      {src != null && (
        <BaseAvatar.Image src={src} alt={alt} className={slots.image()} />
      )}
      <BaseAvatar.Fallback className={slots.fallback()}>{fallback}</BaseAvatar.Fallback>
    </BaseAvatar.Root>
  );
}
