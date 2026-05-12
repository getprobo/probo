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

import type { FC, PropsWithChildren } from "react";
import { NavLink } from "react-router";
import { tv } from "tailwind-variants";

import { useSidebarCollapsed } from "./Sidebar";

const sidebarItem = tv({
  base: "flex items-center gap-2 w-full py-2 rounded-full",
  variants: {
    active: {
      true: "bg-active hover:bg-active-hover active:bg-active-pressed text-txt-primary",
      false: "hover:bg-subtle-hover active:bg-subtle-pressed text-txt-tertiary",
    },
    isCollapsed: {
      true: "px-[10px]",
      false: "px-3",
    },
  },
  defaultVariants: {
    active: false,
  },
});

type Props = PropsWithChildren<{
  icon?: FC<{ size: number }>;
  label: string;
  to?: string;
}>;

export function SidebarItem(props: Props) {
  const isCollapsed = useSidebarCollapsed();
  return (
    <li>
      <NavLink
        to={props.to ?? "/"}
        className={({ isActive }) =>
          sidebarItem({ ...props, active: isActive, isCollapsed })}
      >
        {props.icon && <props.icon size={16} />}
        {isCollapsed ? null : props.label}
      </NavLink>
      {props.children && <ul className="mt-3 ml-5">{props.children}</ul>}
    </li>
  );
}
