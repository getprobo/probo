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

import { Avatar as BaseAvatar } from "@base-ui/react/avatar";
import { avatarColor, avatarInitial } from "@probo/helpers";
import type { ComponentProps, ReactNode } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { avatar } from "./variants";

export type AvatarProps
  = Omit<ComponentProps<typeof BaseAvatar.Root>, "color" | "className">
    & VariantProps<typeof avatar>
    & {
      className?: string;
      // Profile image URL. When absent or it fails to load, `fallback` shows.
      src?: string | null;
      alt?: string;
      name?: string;
      email?: string;
      // Shown until/unless the image loads: user initials (string) or an icon.
      // Defaults to the first letter of `name` when that is set.
      fallback?: ReactNode;
    };

// Foundational avatar primitive (Radix "Avatar") over Base UI's image-loading
// behavior. See contrib/claude/ui.md.
export function Avatar(props: AvatarProps) {
  const {
    size,
    variant,
    color,
    highContrast,
    radius,
    className,
    src,
    alt,
    name,
    email,
    fallback,
    ...rest
  } = props;
  const colorSeed = email ?? name;
  const slots = avatar({
    size,
    variant,
    color: color ?? (colorSeed != null ? avatarColor(colorSeed) : undefined),
    highContrast,
    radius,
  });
  const resolvedSrc = src ?? undefined;
  const resolvedAlt = alt ?? name;
  const resolvedFallback = fallback ?? (name != null ? avatarInitial(name) : "?");

  return (
    <BaseAvatar.Root className={slots.root({ className })} {...rest}>
      {resolvedSrc != null && (
        <BaseAvatar.Image src={resolvedSrc} alt={resolvedAlt} className={slots.image()} />
      )}
      <BaseAvatar.Fallback className={slots.fallback()}>{resolvedFallback}</BaseAvatar.Fallback>
    </BaseAvatar.Root>
  );
}
