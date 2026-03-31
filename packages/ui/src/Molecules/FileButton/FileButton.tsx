// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import type {
  ChangeEventHandler,
  ComponentProps,
  PropsWithChildren,
  RefObject,
} from "react";

import { button, Button } from "../../Atoms/Button/Button";

type Props = PropsWithChildren<{
  onChange: ChangeEventHandler<HTMLInputElement>;
  accept?: string;
  disabled?: boolean;
  className?: string;
  ref?: RefObject<HTMLInputElement | null>;
}>
& Pick<ComponentProps<typeof Button>, "disabled" | "variant" | "icon">;

export function FileButton({
  onChange,
  children,
  icon: IconComponent,
  ref,
  accept,
  ...props
}: Props) {
  return (
    <label className={button({ ...props })}>
      {IconComponent && <IconComponent size={16} className="flex-none" />}
      {children}
      <input
        type="file"
        onChange={onChange}
        hidden
        ref={ref}
        {...(accept && { accept })}
      />
    </label>
  );
}
