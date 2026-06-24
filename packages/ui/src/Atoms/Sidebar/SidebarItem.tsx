// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
