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
