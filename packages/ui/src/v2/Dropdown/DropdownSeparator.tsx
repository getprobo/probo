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

import { Menu } from "@base-ui/react/menu";
import type { ComponentProps } from "react";

import { dropdownSeparator } from "./variants";

export type DropdownSeparatorProps = Omit<ComponentProps<typeof Menu.Separator>, "className"> & {
  className?: string;
};

// A thin divider between groups of items.
export function DropdownSeparator(props: DropdownSeparatorProps) {
  const { className, ...rest } = props;

  return <Menu.Separator className={dropdownSeparator({ className })} {...rest} />;
}
