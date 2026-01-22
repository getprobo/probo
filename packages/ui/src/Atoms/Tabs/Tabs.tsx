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
