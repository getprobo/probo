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

import { List, Root } from "@radix-ui/react-tabs";
import type { HTMLAttributes, PropsWithChildren } from "react";
import { NavLink, type NavLinkProps } from "react-router";
import { tv } from "tailwind-variants";

import { type AsChildProps, Slot } from "../Slot";

const cls = tv({
  slots: {
    wrapper:
            "border-b border-border-low flex gap-6 text-sm font-medium text-txt-secondary",
    item: "py-4 hover:text-txt-primary border-b-2 active:border-border-active -mb-[1px] active:text-txt-primary flex items-center gap-2 cursor-pointer",
    badge: "py-1 px-2 text-txt-secondary text-xs font-semibold rounded-lg bg-highlight",
  },
  variants: {
    active: {
      true: {
        item: "border-border-active text-txt-primary",
        wrapper: "border-border-active",
      },
      false: {
        item: "border-transparent",
      },
    },
  },
})();

export function Tabs(props: HTMLAttributes<HTMLElement>) {
  return (
    <Root>
      <List {...props} className={cls.wrapper(props)} />
    </Root>
  );
}

export function TabItem({
  asChild,
  active,
  ...props
}: AsChildProps<{ active?: boolean; onClick?: () => void }>) {
  const Component = asChild ? Slot : "div";
  return <Component {...props} className={cls.item({ active })} />;
}

export function TabLink(props: PropsWithChildren<NavLinkProps & { isActive?: () => boolean }>) {
  return (
    <NavLink
      className={params => cls.item({ active: params.isActive })}
      {...props}
    />
  );
}

export function TabBadge(props: PropsWithChildren) {
  return <span className={cls.badge()}>{props.children}</span>;
}
