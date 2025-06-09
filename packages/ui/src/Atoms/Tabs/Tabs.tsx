import type { PropsWithChildren } from "react";
import { NavLink } from "react-router";
import { Root, List } from "@radix-ui/react-tabs";
import clsx from "clsx";

export function Tabs(props: PropsWithChildren) {
    return (
        <Root className="TabsRoot" defaultValue="tab1">
            <List className="TabsList" aria-label="Manage your account">
                <div
                    className="border-b border-border-low flex gap-6 text-sm font-medium text-txt-secondary"
                    {...props}
                />
            </List>
        </Root>
    );
}

export function TabLink(props: PropsWithChildren<{ to: string }>) {
    return (
        <NavLink
            className={(params) =>
                clsx(
                    "py-4 hover:text-txt-primary border-b-2 active:border-border-active -mb-[1px] active:text-txt-primary flex items-center gap-1",
                    params.isActive
                        ? "border-border-active text-txt-primary"
                        : "border-transparent",
                )
            }
            {...props}
        />
    );
}

export function TabBadge(props: PropsWithChildren) {
    return (
        <span className="py-1 px-2 text-txt-secondary text-xs font-semibold rounded-lg bg-highlight text-primary">
            {props.children}
        </span>
    );
}
