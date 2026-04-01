import type { FC, PropsWithChildren } from "react";
import { NavLink } from "react-router";
import { tv } from "tailwind-variants";

import { useSidebarCollapsed } from "./Sidebar";

const sidebarItem = tv({
  base: "flex items-center gap-2 w-full py-2 rounded-full",
  variants: {
    active: {
      true: "bg-sidebar-active-bg hover:bg-sidebar-active-hover-bg active:bg-sidebar-active-pressed-bg text-sidebar-text-active font-medium",
      false: "hover:bg-sidebar-hover-bg active:bg-subtle-pressed text-sidebar-text hover:text-sidebar-text-hover",
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
